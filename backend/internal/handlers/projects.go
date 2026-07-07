package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
	"github.com/LCmaster/NextUp/internal/ws"
)

type projectHandler struct {
	queries db.Querier
	hub     *ws.Hub
}

// RegisterProjectRoutes sets up project-related routes.
func RegisterProjectRoutes(r chi.Router, queries db.Querier, hub *ws.Hub) {
	h := &projectHandler{queries: queries, hub: hub}

	r.Route("/projects", func(r chi.Router) {
		r.Post("/", h.createProject)
		r.Get("/", h.listProjects)
		r.Get("/{id}", h.getProject)
		r.Put("/{id}", h.updateProject)
		r.Delete("/{id}", h.deleteProject)
	})
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (h *projectHandler) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	userIDStr, ok := apimiddleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	project, err := h.queries.CreateProject(r.Context(), db.CreateProjectParams{
		Name:        req.Name,
		Description: description,
		OwnerID:     userID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	// Add creator as owner
	_, err = h.queries.AddProjectMember(r.Context(), db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
	})
	if err != nil {
		// Log error, but project was created
	}

	h.hub.BroadcastEvent("project.created", project)
	respondJSON(w, http.StatusCreated, project)
}

func (h *projectHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := apimiddleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	projects, err := h.queries.ListProjectsByMember(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	respondJSON(w, http.StatusOK, projects)
}

func (h *projectHandler) getProject(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	_, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	project, err := h.queries.GetProjectByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

func (h *projectHandler) updateProject(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req updateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	project, err := h.queries.UpdateProject(r.Context(), db.UpdateProjectParams{
		ID:          id,
		Name:        req.Name,
		Description: description,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update project")
		return
	}

	h.hub.BroadcastEvent("project.updated", project)
	respondJSON(w, http.StatusOK, project)
}

func (h *projectHandler) deleteProject(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: id,
		UserID:    userID,
	})
	if err != nil || member.Role != "owner" {
		respondError(w, http.StatusForbidden, "Forbidden: Only owner can delete")
		return
	}

	if err := h.queries.DeleteProject(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	h.hub.BroadcastEvent("project.deleted", map[string]string{"id": chi.URLParam(r, "id")})
	respondJSON(w, http.StatusNoContent, nil)
}
