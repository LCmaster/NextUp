package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

// Beginner is satisfied by *pgxpool.Pool and *pgx.Conn — it is used to open
// a transaction without importing pgxpool directly in this package.
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type ProjectService struct {
	queries db.Querier
	pool    Beginner
	hub     *ws.Hub
}

// NewProjectService creates a ProjectService. pool is used exclusively to open
// transactions for atomic operations (e.g. CreateProject); pass nil in unit
// tests that only exercise non-transactional paths.
func NewProjectService(queries db.Querier, pool Beginner, hub *ws.Hub) *ProjectService {
	return &ProjectService{queries: queries, pool: pool, hub: hub}
}

func (s *ProjectService) CreateProject(ctx context.Context, name, description string, userID pgtype.UUID) (db.Project, error) {
	desc := pgtype.Text{}
	if description != "" {
		desc = pgtype.Text{String: description, Valid: true}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Project{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				slog.Error("failed to rollback CreateProject transaction", "error", rbErr)
			}
		}
	}()

	qtx := db.New(tx)

	project, err := qtx.CreateProject(ctx, db.CreateProjectParams{
		Name:        name,
		Description: desc,
		OwnerID:     userID,
	})
	if err != nil {
		return db.Project{}, fmt.Errorf("failed to create project: %w", err)
	}

	_, err = qtx.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
	})
	if err != nil {
		return db.Project{}, fmt.Errorf("failed to add project owner: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return db.Project{}, fmt.Errorf("failed to commit transaction: %w", err)
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
		return db.Project{}, fmt.Errorf("%w: %v", ErrForbidden, err)
	}

	project, err := s.queries.GetProjectByID(ctx, id)
	if err != nil {
		return db.Project{}, fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	return project, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id, userID pgtype.UUID, name, description string) (db.Project, error) {
	member, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return db.Project{}, ErrForbidden
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
		return ErrForbidden
	}

	if err := s.queries.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	s.hub.BroadcastEvent("project.deleted", map[string]string{"id": id.String()})
	return nil
}

func (s *ProjectService) TransferOwnership(ctx context.Context, projectID, newOwnerID, callerID pgtype.UUID) error {
	callerRole, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    callerID,
	})
	if err != nil || callerRole.Role != "owner" {
		return ErrForbidden
	}

	targetRole, err := s.queries.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    newOwnerID,
	})
	if err != nil || targetRole.Role != "admin" {
		return fmt.Errorf("new owner must be an admin")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				slog.Error("failed to rollback TransferOwnership transaction", "error", rbErr)
			}
		}
	}()

	qtx := db.New(tx)

	_, err = qtx.UpdateProjectMemberRole(ctx, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    callerID,
		Role:      "admin",
	})
	if err != nil {
		return fmt.Errorf("failed to demote owner: %w", err)
	}

	_, err = qtx.UpdateProjectMemberRole(ctx, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    newOwnerID,
		Role:      "owner",
	})
	if err != nil {
		return fmt.Errorf("failed to promote new owner: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
