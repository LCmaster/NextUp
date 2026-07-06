package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

type todoHandler struct {
	queries *db.Queries
	hub     *ws.Hub
}

// RegisterTodoRoutes sets up todo-related routes.
func RegisterTodoRoutes(r chi.Router, queries *db.Queries, hub *ws.Hub) {
	h := &todoHandler{queries: queries, hub: hub}

	r.Route("/todos", func(r chi.Router) {
		r.Post("/", h.createTodo)
		r.Get("/", h.listTodos)
		r.Get("/{id}", h.getTodo)
		r.Put("/{id}", h.updateTodo)
		r.Delete("/{id}", h.deleteTodo)
	})
}

type createTodoRequest struct {
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
}

type updateTodoRequest struct {
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
}

func (h *todoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	var req createTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProjectID == "" || req.Title == "" {
		respondError(w, http.StatusBadRequest, "project_id and title are required")
		return
	}

	projectID := pgtype.UUID{}
	if err := projectID.Scan(req.ProjectID); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project_id format")
		return
	}

	todo, err := h.queries.CreateTodo(r.Context(), db.CreateTodoParams{
		ProjectID: projectID,
		Title:     req.Title,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create todo")
		return
	}

	h.hub.BroadcastEvent("todo.created", todo)
	respondJSON(w, http.StatusCreated, todo)
}

func (h *todoHandler) listTodos(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project_id query parameter required")
		return
	}

	projectID := pgtype.UUID{}
	if err := projectID.Scan(projectIDStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project_id format")
		return
	}

	todos, err := h.queries.ListTodosByProject(r.Context(), projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list todos")
		return
	}

	respondJSON(w, http.StatusOK, todos)
}

func (h *todoHandler) getTodo(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	todo, err := h.queries.GetTodoByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Todo not found")
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

func (h *todoHandler) updateTodo(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	var req updateTodoRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	todo, err := h.queries.UpdateTodo(r.Context(), db.UpdateTodoParams{
		ID:          id,
		Title:       req.Title,
		IsCompleted: req.IsCompleted,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update todo")
		return
	}

	h.hub.BroadcastEvent("todo.updated", todo)
	respondJSON(w, http.StatusOK, todo)
}

func (h *todoHandler) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	if err := h.queries.DeleteTodo(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete todo")
		return
	}

	h.hub.BroadcastEvent("todo.deleted", map[string]string{"id": chi.URLParam(r, "id")})
	respondJSON(w, http.StatusNoContent, nil)
}
