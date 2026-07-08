package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

type ticketHandler struct {
	queries db.Querier
	hub     *ws.Hub
	svc     *services.TicketService
}

// RegisterTicketRoutes sets up ticket-related routes.
func RegisterTicketRoutes(r chi.Router, queries db.Querier, hub *ws.Hub, svc *services.TicketService) {
	h := &ticketHandler{queries: queries, hub: hub, svc: svc}

	r.Route("/tickets", func(r chi.Router) {
		r.Post("/", h.createTicket)
		r.Get("/", h.listTickets)
		r.Get("/{id}", h.getTicket)
		r.Put("/{id}", h.updateTicket)
		r.Delete("/{id}", h.deleteTicket)
		r.Post("/{id}/breakdown", h.breakdownTicket)
	})
}

type createTicketRequest struct {
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Priority    string `json:"priority,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}

type updateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}



func (h *ticketHandler) createTicket(w http.ResponseWriter, r *http.Request) {
	var req createTicketRequest
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

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	// Validate user is in project
	_, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	// Defaults are handled by the database; only set when explicitly provided.
	status := req.Status
	if status == "" {
		status = services.StatusTodo
	}
	priority := req.Priority
	if priority == "" {
		priority = services.PriorityMedium
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	assigneeID := pgtype.UUID{}
	if req.AssigneeID != "" {
		if err := assigneeID.Scan(req.AssigneeID); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid assignee_id format")
			return
		}
	}

	parentID := pgtype.UUID{}
	if req.ParentID != "" {
		if err := parentID.Scan(req.ParentID); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid parent_id format")
			return
		}
	}

	ticket, err := h.queries.CreateTicket(r.Context(), db.CreateTicketParams{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: description,
		Status:      status,
		Priority:    priority,
		AssigneeID:  assigneeID,
		ParentID:    parentID,
		CreatorID:   userID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create ticket")
		return
	}

	h.hub.BroadcastEvent("ticket.created", ticket)
	respondJSON(w, http.StatusCreated, ticket)
}

func (h *ticketHandler) listTickets(w http.ResponseWriter, r *http.Request) {
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

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	tickets, err := h.queries.ListTicketsByProjectAndUser(r.Context(), db.ListTicketsByProjectAndUserParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list tickets")
		return
	}
	if tickets == nil {
		tickets = make([]db.Ticket, 0)
	}
	respondJSON(w, http.StatusOK, tickets)
}

func (h *ticketHandler) getTicket(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	ticket, err := h.queries.GetTicketByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Ticket not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to fetch ticket")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	if canView, _ := h.svc.CanViewTicket(r.Context(), ticket, userID); !canView {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	respondJSON(w, http.StatusOK, ticket)
}

func (h *ticketHandler) updateTicket(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	ticketToUpdate, err := h.queries.GetTicketByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Ticket not found")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	canView, role := h.svc.CanViewTicket(r.Context(), ticketToUpdate, userID)
	if !canView {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	isAssignee := ticketToUpdate.AssigneeID == userID
	canEdit := role == "owner" || role == "admin" || isAssignee

	if !canEdit {
		respondError(w, http.StatusForbidden, "Forbidden: insufficient permissions to edit ticket")
		return
	}

	var req updateTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Only admins and owners can reassign tickets
	if req.AssigneeID != "" && req.AssigneeID != ticketToUpdate.AssigneeID.String() && role != "owner" && role != "admin" {
		respondError(w, http.StatusForbidden, "Forbidden: only admins can assign tickets")
		return
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description = pgtype.Text{String: req.Description, Valid: true}
	}

	assigneeID := pgtype.UUID{}
	if req.AssigneeID != "" {
		if err := assigneeID.Scan(req.AssigneeID); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid assignee_id format")
			return
		}
	}

	parentID := pgtype.UUID{}
	if req.ParentID != "" {
		if err := parentID.Scan(req.ParentID); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid parent_id format")
			return
		}
	}

	ticket, err := h.queries.UpdateTicket(r.Context(), db.UpdateTicketParams{
		ID:          id,
		Title:       req.Title,
		Description: description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  assigneeID,
		ParentID:    parentID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update ticket")
		return
	}

	h.hub.BroadcastEvent("ticket.updated", ticket)
	respondJSON(w, http.StatusOK, ticket)
}

func (h *ticketHandler) deleteTicket(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	ticketToDelete, err := h.queries.GetTicketByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Ticket not found")
		return
	}

	userIDStr, _ := apimiddleware.UserIDFromContext(r.Context())
	userID := pgtype.UUID{}
	userID.Scan(userIDStr)

	canView, role := h.svc.CanViewTicket(r.Context(), ticketToDelete, userID)
	if !canView || (role != "owner" && role != "admin" && ticketToDelete.CreatorID != userID) {
		respondError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := h.queries.DeleteTicket(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete ticket")
		return
	}

	h.hub.BroadcastEvent("ticket.deleted", map[string]string{"id": chi.URLParam(r, "id")})
	respondJSON(w, http.StatusNoContent, nil)
}

// breakdownTicket delegates to TicketService so the Gemini client is shared
// and the HTTP request context is propagated to cancel the AI call if the
// client disconnects.
func (h *ticketHandler) breakdownTicket(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	tickets, err := h.svc.Breakdown(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tickets == nil {
		tickets = make([]db.Ticket, 0)
	}
	respondJSON(w, http.StatusOK, tickets)
}
