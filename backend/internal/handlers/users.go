package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/LCmaster/NextUp/internal/db"
	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
)

type userHandler struct {
	queries   db.Querier
	jwtSecret []byte
}

// RegisterUserRoutes sets up user-related routes.
func RegisterUserRoutes(r chi.Router, queries db.Querier, jwtSecret []byte) {
	h := &userHandler{queries: queries, jwtSecret: jwtSecret}

	r.Route("/users", func(r chi.Router) {
		// Public routes — no auth required
		r.Get("/setup-status", h.getSetupStatus)
		r.Post("/register", h.register)
		r.Post("/login", h.login)

		// Protected — requires a valid JWT cookie
		r.With(apimiddleware.Auth(jwtSecret)).Post("/logout", h.logout)
		r.With(apimiddleware.Auth(jwtSecret)).Get("/me", h.getProfile)
		r.With(apimiddleware.Auth(jwtSecret)).Get("/", h.listUsers)
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

// register creates a user account and issues a session cookie.
func (h *userHandler) register(w http.ResponseWriter, r *http.Request) {

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

	// Issue JWT session cookie
	cookie, err := apimiddleware.NewJWTCookie(user.ID.String(), h.jwtSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, cookie)

	respondJSON(w, http.StatusCreated, user)
}

// login authenticates a user with email and password and issues a session cookie.
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

	// Issue JWT session cookie
	cookie, err := apimiddleware.NewJWTCookie(user.ID.String(), h.jwtSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}
	http.SetCookie(w, cookie)

	respondJSON(w, http.StatusOK, user)
}

// logout clears the session cookie.
func (h *userHandler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "nextup_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// getProfile returns the authenticated user's profile using the user ID from the JWT context.
func (h *userHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := apimiddleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var pgID pgtype.UUID
	if err := pgID.Scan(userID); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), pgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// listUsers returns all users (for member selection in the UI).
func (h *userHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.queries.ListUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	respondJSON(w, http.StatusOK, users)
}

