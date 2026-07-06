package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/ws"
)

// AIGenerator is an interface over the Gemini model, enabling test mocking.
type AIGenerator interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// TicketService encapsulates business logic for tickets.
type TicketService struct {
	queries db.Querier
	hub     *ws.Hub
	ai      AIGenerator
}

// NewTicketService creates a TicketService. ai may be nil if no API key is configured.
func NewTicketService(queries db.Querier, hub *ws.Hub, ai AIGenerator) *TicketService {
	return &TicketService{queries: queries, hub: hub, ai: ai}
}

// generatedTask is the expected shape of AI-generated sub-tasks.
type generatedTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Breakdown uses the AI model to decompose a ticket into sub-tasks, persists
// them, and broadcasts each creation event. The caller's context is used for
// all external calls so the work is cancelled if the client disconnects.
func (s *TicketService) Breakdown(ctx context.Context, ticketID pgtype.UUID) ([]db.Ticket, error) {
	if s.ai == nil {
		return nil, fmt.Errorf("AI breakdown is not configured: GEMINI_API_KEY is missing")
	}

	ticket, err := s.queries.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	prompt := "Break down the following ticket into 2-5 smaller, actionable sub-tasks. " +
		"Return ONLY a JSON array of objects with 'title' and 'description' keys.\n\n" +
		"Ticket Title: " + ticket.Title + "\n"
	if ticket.Description.Valid {
		prompt += "Ticket Description: " + ticket.Description.String + "\n"
	}

	resp, err := s.ai.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected AI response format")
	}

	// Strip optional markdown fences the model may include despite JSON mode.
	jsonText := strings.TrimPrefix(string(text), "```json\n")
	jsonText = strings.TrimSuffix(jsonText, "\n```")

	var tasks []generatedTask
	if err := json.Unmarshal([]byte(jsonText), &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	created := make([]db.Ticket, 0, len(tasks))
	for _, task := range tasks {
		desc := pgtype.Text{}
		if task.Description != "" {
			desc = pgtype.Text{String: task.Description, Valid: true}
		}

		newTicket, err := s.queries.CreateTicket(ctx, db.CreateTicketParams{
			ProjectID:   ticket.ProjectID,
			Title:       task.Title,
			Description: desc,
			Status:      "todo",
			Priority:    "medium",
			AssigneeID:  pgtype.UUID{},
			ParentID:    ticketID,
		})
		if err != nil {
			log.Printf("services: failed to create sub-task %q: %v", task.Title, err)
			continue
		}

		created = append(created, newTicket)
		s.hub.BroadcastEvent("ticket.created", newTicket)
	}

	return created, nil
}
