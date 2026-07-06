package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
)

const validUUID = "12345678-1234-1234-1234-123456789012"

func TestCreateTicket_Success(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"project_id":"` + validUUID + `","title":"Fix login bug","priority":"high"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tickets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := do(r, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateTicket_MissingTitle(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"project_id":"` + validUUID + `"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tickets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := do(r, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateTicket_InvalidProjectID(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"project_id":"not-a-uuid","title":"Fix bug"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/tickets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := do(r, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetTicket_NotFound(t *testing.T) {
	q := &MockQuerier{
		GetTicketByIDFn: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{}, pgx.ErrNoRows
		},
	}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tickets/"+validUUID, nil)
	rr := do(r, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetTicket_InternalError(t *testing.T) {
	q := &MockQuerier{
		GetTicketByIDFn: func(_ context.Context, _ pgtype.UUID) (db.Ticket, error) {
			return db.Ticket{}, errors.New("db connection lost")
		},
	}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tickets/"+validUUID, nil)
	rr := do(r, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestUpdateTicket_Success(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"title":"Updated title","status":"in_progress","priority":"medium"}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/tickets/"+validUUID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := do(r, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteTicket_Success(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/tickets/"+validUUID, nil)
	rr := do(r, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestListTickets_ByProject(t *testing.T) {
	q := &MockQuerier{
		ListTicketsByProjectFn: func(_ context.Context, _ pgtype.UUID) ([]db.Ticket, error) {
			return []db.Ticket{
				{Title: "T1", Status: "todo", Priority: "medium"},
				{Title: "T2", Status: "done", Priority: "low"},
			}, nil
		},
	}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tickets?project_id="+validUUID, nil)
	rr := do(r, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "T1") {
		t.Errorf("expected body to contain 'T1', got: %s", rr.Body.String())
	}
}
