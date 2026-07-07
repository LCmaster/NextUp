package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

type ProjectService struct {
	queries db.Querier
	hub     *ws.Hub
}

func NewProjectService(queries db.Querier, hub *ws.Hub) *ProjectService {
	return &ProjectService{queries: queries, hub: hub}
}

func (s *ProjectService) CreateProject(ctx context.Context, name, description string, userID pgtype.UUID) (db.Project, error) {
	desc := pgtype.Text{}
	if description != "" {
		desc = pgtype.Text{String: description, Valid: true}
	}

	project, err := s.queries.CreateProject(ctx, db.CreateProjectParams{
		Name:        name,
		Description: desc,
		OwnerID:     userID,
	})
	if err != nil {
		return db.Project{}, fmt.Errorf("failed to create project: %w", err)
	}

	// Add creator as owner
	_, err = s.queries.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
	})
	if err != nil {
		// Log error, but project was created
	}

	s.hub.BroadcastEvent("project.created", project)
	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, userID pgtype.UUID) ([]db.Project, error) {
	projects, err := s.queries.ListProjectsByMember(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id, userID pgtype.UUID) (db.Project, error) {
	_, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil {
		return db.Project{}, fmt.Errorf("forbidden: %w", err)
	}

	project, err := s.queries.GetProjectByID(ctx, id)
	if err != nil {
		return db.Project{}, fmt.Errorf("not found: %w", err)
	}

	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id, userID pgtype.UUID, name, description string) (db.Project, error) {
	member, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return db.Project{}, fmt.Errorf("forbidden: insufficient permissions")
	}

	desc := pgtype.Text{}
	if description != "" {
		desc = pgtype.Text{String: description, Valid: true}
	}

	project, err := s.queries.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id,
		Name:        name,
		Description: desc,
	})
	if err != nil {
		return db.Project{}, fmt.Errorf("failed to update project: %w", err)
	}

	s.hub.BroadcastEvent("project.updated", project)
	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id, userID pgtype.UUID) error {
	member, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil || member.Role != "owner" {
		return fmt.Errorf("forbidden: only owner can delete")
	}

	if err := s.queries.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	s.hub.BroadcastEvent("project.deleted", map[string]string{"id": id.String()})
	return nil
}
