package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/LCmaster/NextUp/internal/db"
)

type userHandler struct {
	queries db.Querier
}

// RegisterUserRoutes sets up user-related routes.
func RegisterUserRoutes(r chi.Router, queries db.Querier) {
	h := &userHandler{queries: queries}

	r.Route("/users", func(r chi.Router) {
		r.Get("/setup-status", h.getSetupStatus)
		r.Post("/setup", h.setupAccount)
		r.Post("/login", h.login)
		r.Get("/me", h.getProfile)
	})
}

type setupRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	GithubLink string `json:"github_link,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// getSetupStatus checks if the first user has been created.
func (h *userHandler) getSetupStatus(w http.ResponseWriter, r *http.Request) {
	count, err := h.queries.CountUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check setup status")
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"is_setup": count > 0})
}

// setupAccount creates the first user account.
func (h *userHandler) setupAccount(w http.ResponseWriter, r *http.Request) {
	// Check if a user already exists
	count, err := h.queries.CountUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check setup status")
		return
	}
	if count > 0 {
		respondError(w, http.StatusConflict, "Account already set up")
		return
	}

	var req setupRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "First name, last name, email, and password are required")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	githubLink := pgtype.Text{}
	if req.GithubLink != "" {
		githubLink = pgtype.Text{String: req.GithubLink, Valid: true}
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		GithubLink:   githubLink,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// login authenticates a user with email and password.
func (h *userHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// getProfile returns the current user's profile (returns the first user for now).
func (h *userHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	count, err := h.queries.CountUsers(r.Context())
	if err != nil || count == 0 {
		respondError(w, http.StatusNotFound, "No user found")
		return
	}

	// For single-user mode, get user by email from query param or return first user
	email := r.URL.Query().Get("email")
	if email == "" {
		respondError(w, http.StatusBadRequest, "Email query parameter required")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}
