package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/option"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

type ticketHandler struct {
	queries *db.Queries
	hub     *ws.Hub
}

// RegisterTicketRoutes sets up ticket-related routes.
func RegisterTicketRoutes(r chi.Router, queries *db.Queries, hub *ws.Hub) {
	h := &ticketHandler{queries: queries, hub: hub}

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

	status := req.Status
	if status == "" {
		status = "todo"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
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

	tickets, err := h.queries.ListTicketsByProject(r.Context(), projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list tickets")
		return
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
		respondError(w, http.StatusNotFound, "Ticket not found")
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

	var req updateTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
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

	if err := h.queries.DeleteTicket(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete ticket")
		return
	}

	h.hub.BroadcastEvent("ticket.deleted", map[string]string{"id": chi.URLParam(r, "id")})
	respondJSON(w, http.StatusNoContent, nil)
}

func (h *ticketHandler) breakdownTicket(w http.ResponseWriter, r *http.Request) {
	id := pgtype.UUID{}
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ticket ID")
		return
	}

	ticket, err := h.queries.GetTicketByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Ticket not found")
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		respondError(w, http.StatusInternalServerError, "GEMINI_API_KEY not configured")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to initialize AI client")
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")
	model.ResponseMIMEType = "application/json"

	prompt := "Break down the following ticket into 3-5 smaller, actionable sub-tasks. " +
		"Return ONLY a JSON array of objects with 'title' and 'description' keys.\n\n" +
		"Ticket Title: " + ticket.Title + "\n"
	if ticket.Description.Valid {
		prompt += "Ticket Description: " + ticket.Description.String + "\n"
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate breakdown")
		return
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		respondError(w, http.StatusInternalServerError, "No response from AI")
		return
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Invalid AI response format")
		return
	}

	var generatedTasks []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	jsonText := strings.TrimPrefix(string(text), "```json\n")
	jsonText = strings.TrimSuffix(jsonText, "\n```")

	if err := json.Unmarshal([]byte(jsonText), &generatedTasks); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	var createdTickets []db.Ticket
	for _, task := range generatedTasks {
		desc := pgtype.Text{}
		if task.Description != "" {
			desc = pgtype.Text{String: task.Description, Valid: true}
		}

		newTicket, err := h.queries.CreateTicket(r.Context(), db.CreateTicketParams{
			ProjectID:   ticket.ProjectID,
			Title:       task.Title,
			Description: desc,
			Status:      "todo",
			Priority:    "medium",
			AssigneeID:  pgtype.UUID{},
			ParentID:    id,
		})
		if err == nil {
			createdTickets = append(createdTickets, newTicket)
			h.hub.BroadcastEvent("ticket.created", newTicket)
		}
	}

	respondJSON(w, http.StatusOK, createdTickets)
}
