package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

type projectHandler struct {
	queries *db.Queries
	hub     *ws.Hub
}

// RegisterProjectRoutes sets up project-related routes.
func RegisterProjectRoutes(r chi.Router, queries *db.Queries, hub *ws.Hub) {
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
	OwnerID     string `json:"owner_id"`
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

	if req.Name == "" || req.OwnerID == "" {
		respondError(w, http.StatusBadRequest, "Name and owner_id are required")
		return
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(req.OwnerID); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid owner_id format")
		return
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	project, err := h.queries.CreateProject(r.Context(), db.CreateProjectParams{
		Name:        req.Name,
		Description: description,
		OwnerID:     ownerID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	h.hub.BroadcastEvent("project.created", project)
	respondJSON(w, http.StatusCreated, project)
}

func (h *projectHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	ownerIDStr := r.URL.Query().Get("owner_id")
	if ownerIDStr == "" {
		respondError(w, http.StatusBadRequest, "owner_id query parameter required")
		return
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(ownerIDStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid owner_id format")
		return
	}

	projects, err := h.queries.ListProjectsByOwner(r.Context(), ownerID)
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

	if err := h.queries.DeleteProject(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	h.hub.BroadcastEvent("project.deleted", map[string]string{"id": chi.URLParam(r, "id")})
	respondJSON(w, http.StatusNoContent, nil)
}
