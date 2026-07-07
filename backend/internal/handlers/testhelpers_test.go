package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/handlers"
	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

// ── Auth Helper ───────────────────────────────────────────────────────────────

func addAuthCookie(req *http.Request, userID string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("secret"))
	req.AddCookie(&http.Cookie{
		Name:  "nextup_session",
		Value: tokenString,
	})
}

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
	ListProjectsByMemberFn        func(ctx context.Context, userID pgtype.UUID) ([]db.Project, error)
	ListTicketsByProjectAndUserFn func(ctx context.Context, arg db.ListTicketsByProjectAndUserParams) ([]db.Ticket, error)
	UpdateProjectFn               func(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error)
	UpdateTicketFn                func(ctx context.Context, arg db.UpdateTicketParams) (db.Ticket, error)
	UpdateUserFn                  func(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	
	// Project Members & Invites
	AddProjectMemberFn        func(ctx context.Context, arg db.AddProjectMemberParams) (db.ProjectMember, error)
	CreateProjectInviteFn     func(ctx context.Context, arg db.CreateProjectInviteParams) (db.ProjectInvite, error)
	DeleteProjectInviteFn     func(ctx context.Context, id pgtype.UUID) error
	GetProjectInviteByTokenFn func(ctx context.Context, token string) (db.ProjectInvite, error)
	GetProjectMemberFn        func(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error)
	ListProjectInvitesFn      func(ctx context.Context, projectID pgtype.UUID) ([]db.ProjectInvite, error)
	ListProjectMembersFn      func(ctx context.Context, projectID pgtype.UUID) ([]db.ListProjectMembersRow, error)
	ListUsersFn               func(ctx context.Context) ([]db.ListUsersRow, error)
	RemoveProjectMemberFn     func(ctx context.Context, arg db.RemoveProjectMemberParams) error
	UpdateProjectMemberRoleFn func(ctx context.Context, arg db.UpdateProjectMemberRoleParams) (db.ProjectMember, error)
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
	r.Use(apimiddleware.Auth([]byte("secret")))
	r.Route("/api/v1", func(r chi.Router) {
		handlers.RegisterUserRoutes(r, q, []byte("secret"))
		projectSvc := services.NewProjectService(q, hub)
		handlers.RegisterProjectRoutes(r, projectSvc)
		handlers.RegisterProjectMemberRoutes(r, q, hub)
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
func (m *MockQuerier) AddProjectMember(ctx context.Context, arg db.AddProjectMemberParams) (db.ProjectMember, error) {
	if m.AddProjectMemberFn != nil { return m.AddProjectMemberFn(ctx, arg) }
	return db.ProjectMember{}, nil 
}
func (m *MockQuerier) CreateProjectInvite(ctx context.Context, arg db.CreateProjectInviteParams) (db.ProjectInvite, error) {
	if m.CreateProjectInviteFn != nil { return m.CreateProjectInviteFn(ctx, arg) }
	return db.ProjectInvite{}, nil 
}
func (m *MockQuerier) DeleteProjectInvite(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteProjectInviteFn != nil { return m.DeleteProjectInviteFn(ctx, id) }
	return nil 
}
func (m *MockQuerier) GetProjectInviteByToken(ctx context.Context, token string) (db.ProjectInvite, error) {
	if m.GetProjectInviteByTokenFn != nil { return m.GetProjectInviteByTokenFn(ctx, token) }
	return db.ProjectInvite{}, nil 
}
func (m *MockQuerier) GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error) {
	if m.GetProjectMemberFn != nil { return m.GetProjectMemberFn(ctx, arg) }
	return db.ProjectMember{Role: "owner"}, nil 
}
func (m *MockQuerier) ListProjectInvites(ctx context.Context, projectID pgtype.UUID) ([]db.ProjectInvite, error) {
	if m.ListProjectInvitesFn != nil { return m.ListProjectInvitesFn(ctx, projectID) }
	return nil, nil 
}
func (m *MockQuerier) ListProjectMembers(ctx context.Context, projectID pgtype.UUID) ([]db.ListProjectMembersRow, error) {
	if m.ListProjectMembersFn != nil { return m.ListProjectMembersFn(ctx, projectID) }
	return nil, nil 
}
func (m *MockQuerier) ListProjectsByMember(ctx context.Context, userID pgtype.UUID) ([]db.Project, error) {
	if m.ListProjectsByMemberFn != nil { return m.ListProjectsByMemberFn(ctx, userID) }
	return nil, nil 
}
func (m *MockQuerier) ListTicketsByProjectAndUser(ctx context.Context, arg db.ListTicketsByProjectAndUserParams) ([]db.Ticket, error) {
	if m.ListTicketsByProjectAndUserFn != nil { return m.ListTicketsByProjectAndUserFn(ctx, arg) }
	return nil, nil 
}
func (m *MockQuerier) ListUsers(ctx context.Context) ([]db.ListUsersRow, error) {
	if m.ListUsersFn != nil { return m.ListUsersFn(ctx) }
	return nil, nil 
}
func (m *MockQuerier) RemoveProjectMember(ctx context.Context, arg db.RemoveProjectMemberParams) error {
	if m.RemoveProjectMemberFn != nil { return m.RemoveProjectMemberFn(ctx, arg) }
	return nil 
}
func (m *MockQuerier) UpdateProjectMemberRole(ctx context.Context, arg db.UpdateProjectMemberRoleParams) (db.ProjectMember, error) {
	if m.UpdateProjectMemberRoleFn != nil { return m.UpdateProjectMemberRoleFn(ctx, arg) }
	return db.ProjectMember{}, nil 
}
