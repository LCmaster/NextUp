package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

func mustUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	id := pgtype.UUID{}
	if err := id.Scan("00000000-0000-0000-0000-000000000001"); err != nil {
		t.Fatalf("bad uuid: %v", err)
	}
	return id
}

func mustText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// ── MockQuerier ───────────────────────────────────────────────────────────────

// mockQuerier is a minimal implementation of db.Querier for service tests.
type mockQuerier struct {
	getTicketByID func(ctx context.Context, id pgtype.UUID) (db.Ticket, error)
	createTicket  func(ctx context.Context, arg db.CreateTicketParams) (db.Ticket, error)
	// All other methods return zero values.
	db.Querier
}

func (m *mockQuerier) GetTicketByID(ctx context.Context, id pgtype.UUID) (db.Ticket, error) {
	if m.getTicketByID != nil {
		return m.getTicketByID(ctx, id)
	}
	return db.Ticket{}, nil
}

func (m *mockQuerier) CreateTicket(ctx context.Context, arg db.CreateTicketParams) (db.Ticket, error) {
	if m.createTicket != nil {
		return m.createTicket(ctx, arg)
	}
	return db.Ticket{Title: arg.Title, Status: arg.Status, Priority: arg.Priority}, nil
}

// ── MockAI ────────────────────────────────────────────────────────────────────

type mockAI struct {
	generateContent func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

func (m *mockAI) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.generateContent(ctx, parts...)
}

// makeAIResponse builds a *genai.GenerateContentResponse whose first candidate
// contains the provided JSON string as a Text part.
func makeAIResponse(jsonStr string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []genai.Part{genai.Text(jsonStr)},
				},
			},
		},
	}
}

// newTestHub creates a Hub whose Run loop is started so BroadcastEvent doesn't
// block on the buffered channel.
func newTestHub(t *testing.T) *ws.Hub {
	t.Helper()
	hub := ws.NewHub()
	go hub.Run()
	return hub
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestBreakdown_NoAI(t *testing.T) {
	svc := services.NewTicketService(&mockQuerier{}, newTestHub(t), nil)
	_, err := svc.Breakdown(context.Background(), mustUUID(t))
	if err == nil {
		t.Fatal("expected error when AI is nil, got nil")
	}
}

func TestBreakdown_TicketNotFound(t *testing.T) {
	q := &mockQuerier{
		getTicketByID: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{}, errors.New("no rows")
		},
	}
	ai := &mockAI{}
	svc := services.NewTicketService(q, newTestHub(t), ai)
	_, err := svc.Breakdown(context.Background(), mustUUID(t))
	if err == nil {
		t.Fatal("expected error when ticket not found, got nil")
	}
}

func TestBreakdown_AIFailure(t *testing.T) {
	q := &mockQuerier{
		getTicketByID: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{Title: "Fix bug"}, nil
		},
	}
	ai := &mockAI{
		generateContent: func(_ context.Context, _ ...genai.Part) (*genai.GenerateContentResponse, error) {
			return nil, errors.New("AI error")
		},
	}
	svc := services.NewTicketService(q, newTestHub(t), ai)
	_, err := svc.Breakdown(context.Background(), mustUUID(t))
	if err == nil {
		t.Fatal("expected error from AI failure, got nil")
	}
}

func TestBreakdown_MalformedJSON(t *testing.T) {
	q := &mockQuerier{
		getTicketByID: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{Title: "Fix bug"}, nil
		},
	}
	ai := &mockAI{
		generateContent: func(_ context.Context, _ ...genai.Part) (*genai.GenerateContentResponse, error) {
			return makeAIResponse("not valid json at all"), nil
		},
	}
	svc := services.NewTicketService(q, newTestHub(t), ai)
	_, err := svc.Breakdown(context.Background(), mustUUID(t))
	if err == nil {
		t.Fatal("expected parse error for malformed JSON, got nil")
	}
}

func TestBreakdown_CreateSubTasksSuccess(t *testing.T) {
	parentID := mustUUID(t)
	projectID := mustUUID(t)

	type taskShape struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	tasks := []taskShape{
		{Title: "Sub-task A", Description: "Do A"},
		{Title: "Sub-task B", Description: "Do B"},
	}
	tasksJSON, _ := json.Marshal(tasks)

	var createdTitles []string

	q := &mockQuerier{
		getTicketByID: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{
				ID:        parentID,
				ProjectID: projectID,
				Title:     "Parent ticket",
				Description: mustText("Some description"),
			}, nil
		},
		createTicket: func(_ context.Context, arg db.CreateTicketParams) (db.Ticket, error) {
			createdTitles = append(createdTitles, arg.Title)
			return db.Ticket{
				Title:    arg.Title,
				Status:   arg.Status,
				Priority: arg.Priority,
				ParentID: arg.ParentID,
			}, nil
		},
	}
	ai := &mockAI{
		generateContent: func(_ context.Context, _ ...genai.Part) (*genai.GenerateContentResponse, error) {
			return makeAIResponse(string(tasksJSON)), nil
		},
	}

	svc := services.NewTicketService(q, newTestHub(t), ai)
	results, err := svc.Breakdown(context.Background(), parentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 sub-tasks, got %d", len(results))
	}
	if createdTitles[0] != "Sub-task A" || createdTitles[1] != "Sub-task B" {
		t.Errorf("unexpected sub-task titles: %v", createdTitles)
	}
	// All created tickets should have "todo" status and "medium" priority.
	for _, r := range results {
		if r.Status != "todo" {
			t.Errorf("expected status 'todo', got %q", r.Status)
		}
		if r.Priority != "medium" {
			t.Errorf("expected priority 'medium', got %q", r.Priority)
		}
	}
}
