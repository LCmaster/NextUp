package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/handlers"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

// ── MockQuerier ───────────────────────────────────────────────────────────────

// MockQuerier implements db.Querier; only the fields used by the tested handler
// path need to be set — all other methods return zero values.
type MockQuerier struct {
	CountUsersFn          func(ctx context.Context) (int64, error)
	CreateProjectFn       func(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
	CreateTicketFn        func(ctx context.Context, arg db.CreateTicketParams) (db.Ticket, error)
	CreateUserFn          func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	DeleteProjectFn       func(ctx context.Context, id pgtype.UUID) error
	DeleteTicketFn        func(ctx context.Context, id pgtype.UUID) error
	GetProjectByIDFn      func(ctx context.Context, id pgtype.UUID) (db.Project, error)
	GetTicketByIDFn       func(ctx context.Context, id pgtype.UUID) (db.Ticket, error)
	GetUserByEmailFn      func(ctx context.Context, email string) (db.User, error)
	GetUserByIDFn         func(ctx context.Context, id pgtype.UUID) (db.User, error)
	ListProjectsByOwnerFn func(ctx context.Context, ownerID pgtype.UUID) ([]db.Project, error)
	ListTicketsByProjectFn func(ctx context.Context, projectID pgtype.UUID) ([]db.Ticket, error)
	UpdateProjectFn       func(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error)
	UpdateTicketFn        func(ctx context.Context, arg db.UpdateTicketParams) (db.Ticket, error)
	UpdateUserFn          func(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
}

func (m *MockQuerier) CountUsers(ctx context.Context) (int64, error) {
	if m.CountUsersFn != nil {
		return m.CountUsersFn(ctx)
	}
	return 0, nil
}
func (m *MockQuerier) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	if m.CreateProjectFn != nil {
		return m.CreateProjectFn(ctx, arg)
	}
	return db.Project{Name: arg.Name}, nil
}
func (m *MockQuerier) CreateTicket(ctx context.Context, arg db.CreateTicketParams) (db.Ticket, error) {
	if m.CreateTicketFn != nil {
		return m.CreateTicketFn(ctx, arg)
	}
	return db.Ticket{Title: arg.Title, Status: arg.Status, Priority: arg.Priority}, nil
}
func (m *MockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, arg)
	}
	return db.User{Email: arg.Email}, nil
}
func (m *MockQuerier) DeleteProject(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteProjectFn != nil {
		return m.DeleteProjectFn(ctx, id)
	}
	return nil
}
func (m *MockQuerier) DeleteTicket(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteTicketFn != nil {
		return m.DeleteTicketFn(ctx, id)
	}
	return nil
}
func (m *MockQuerier) GetProjectByID(ctx context.Context, id pgtype.UUID) (db.Project, error) {
	if m.GetProjectByIDFn != nil {
		return m.GetProjectByIDFn(ctx, id)
	}
	return db.Project{}, nil
}
func (m *MockQuerier) GetTicketByID(ctx context.Context, id pgtype.UUID) (db.Ticket, error) {
	if m.GetTicketByIDFn != nil {
		return m.GetTicketByIDFn(ctx, id)
	}
	return db.Ticket{}, nil
}
func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return db.User{}, nil
}
func (m *MockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return db.User{}, nil
}
func (m *MockQuerier) ListProjectsByOwner(ctx context.Context, ownerID pgtype.UUID) ([]db.Project, error) {
	if m.ListProjectsByOwnerFn != nil {
		return m.ListProjectsByOwnerFn(ctx, ownerID)
	}
	return nil, nil
}
func (m *MockQuerier) ListTicketsByProject(ctx context.Context, projectID pgtype.UUID) ([]db.Ticket, error) {
	if m.ListTicketsByProjectFn != nil {
		return m.ListTicketsByProjectFn(ctx, projectID)
	}
	return nil, nil
}
func (m *MockQuerier) UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error) {
	if m.UpdateProjectFn != nil {
		return m.UpdateProjectFn(ctx, arg)
	}
	return db.Project{Name: arg.Name}, nil
}
func (m *MockQuerier) UpdateTicket(ctx context.Context, arg db.UpdateTicketParams) (db.Ticket, error) {
	if m.UpdateTicketFn != nil {
		return m.UpdateTicketFn(ctx, arg)
	}
	return db.Ticket{Title: arg.Title, Status: arg.Status}, nil
}
func (m *MockQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(ctx, arg)
	}
	return db.User{}, nil
}

// ── Router helper ─────────────────────────────────────────────────────────────

// newTestRouter builds a chi router wired with all handlers using the provided
// mock querier and a no-op Hub (not running, broadcast channel is buffered).
func newTestRouter(q *MockQuerier) http.Handler {
	hub := ws.NewHub()
	// NB: hub.Run() is intentionally not called; the buffered broadcast
	// channel (256) absorbs events without blocking in tests.
	svc := services.NewTicketService(q, hub, nil) // AI disabled in handler tests
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		handlers.RegisterUserRoutes(r, q)
		handlers.RegisterProjectRoutes(r, q, hub)
		handlers.RegisterTicketRoutes(r, q, hub, svc)
	})
	return r
}

// do is a convenience wrapper that executes a request against the test router
// and returns the response recorder.
func do(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}
