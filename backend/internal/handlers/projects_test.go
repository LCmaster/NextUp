package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
)

func TestCreateProject_Success(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"name":"My Project","owner_id":"` + validUUID + `"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookie(req, validUUID)
	rr := do(r, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateProject_MissingName(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	body := `{"owner_id":"` + validUUID + `"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookie(req, validUUID)
	rr := do(r, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetProject_Success(t *testing.T) {
	q := &MockQuerier{
		GetProjectByIDFn: func(_ context.Context, _ pgtype.UUID) (db.Project, error) {
			return db.Project{Name: "Test Project"}, nil
		},
	}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/projects/"+validUUID, nil)
	addAuthCookie(req, validUUID)
	rr := do(r, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Test Project") {
		t.Errorf("expected body to contain 'Test Project', got: %s", rr.Body.String())
	}
}

func TestDeleteProject_Success(t *testing.T) {
	q := &MockQuerier{}
	r := newTestRouter(q)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/projects/"+validUUID, nil)
	addAuthCookie(req, validUUID)
	rr := do(r, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
