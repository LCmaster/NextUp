package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
	"github.com/LCmaster/NextUp/internal/services"
)

type projectHandler struct {
	svc *services.ProjectService
}

// RegisterProjectRoutes sets up project-related routes.
func RegisterProjectRoutes(r chi.Router, svc *services.ProjectService) {
	h := &projectHandler{svc: svc}

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

	project, err := h.svc.CreateProject(r.Context(), req.Name, req.Description, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

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

	projects, err := h.svc.ListProjects(r.Context(), userID)
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

	project, err := h.svc.GetProject(r.Context(), id, userID)
	if err != nil {
		// Note: The service might return forbidden or not found, but we map it loosely here
		if err.Error() == "forbidden: forbidden" || err.Error() == "forbidden: no rows in result set" {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
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

	var req updateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.svc.UpdateProject(r.Context(), id, userID, req.Name, req.Description)
	if err != nil {
		if err.Error() == "forbidden: insufficient permissions" {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update project")
		return
	}

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

	if err := h.svc.DeleteProject(r.Context(), id, userID); err != nil {
		if err.Error() == "forbidden: only owner can delete" {
			respondError(w, http.StatusForbidden, "Forbidden: Only owner can delete")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
