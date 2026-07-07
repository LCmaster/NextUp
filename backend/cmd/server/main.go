package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/handlers"
	"github.com/LCmaster/NextUp/internal/mailer"
	apimiddleware "github.com/LCmaster/NextUp/internal/middleware"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

func main() {
	// Set up structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Read config from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	
	// Configure Mailer
	var m mailer.Mailer
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		smtpPort := os.Getenv("SMTP_PORT")
		smtpUser := os.Getenv("SMTP_USER")
		smtpPass := os.Getenv("SMTP_PASS")
		smtpFrom := os.Getenv("SMTP_FROM")
		m = mailer.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
		slog.Info("SMTP mailer initialized", "host", smtpHost)
	} else {
		m = mailer.NewMockMailer()
		slog.Info("SMTP_HOST not set, using mock mailer")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		slog.Error("Unable to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to database")

	// Run migrations
	if err := runMigrations(databaseURL); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations applied")

	// Initialize Gemini AI client once at startup.
	// If GEMINI_API_KEY is absent the server still starts; breakdown endpoint
	// will return a clear error rather than panicking.
	var aiModel *genai.GenerativeModel
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		aiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			slog.Error("Failed to create Gemini client", "error", err)
			os.Exit(1)
		}
		defer aiClient.Close()
		m := aiClient.GenerativeModel("gemini-flash-latest")
		m.ResponseMIMEType = "application/json"
		aiModel = m
		slog.Info("Gemini AI client initialized")
	} else {
		slog.Warn("GEMINI_API_KEY not set — AI breakdown endpoint will be unavailable")
	}

	// Initialize dependencies
	queries := db.New(pool)
	hub := ws.NewHub()
	go hub.Run()

	jwtSecretBytes := []byte(jwtSecret)
	ticketSvc := services.NewTicketService(queries, hub, aiModel)

	// Build router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:80", "http://localhost"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// User routes — login/setup are public, /me and /logout are protected inside RegisterUserRoutes
		handlers.RegisterUserRoutes(r, queries, jwtSecretBytes)

		// All project and ticket routes require authentication
		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.Auth(jwtSecretBytes))
			projectSvc := services.NewProjectService(queries, pool, hub)
			handlers.RegisterProjectRoutes(r, projectSvc)
			handlers.RegisterProjectMemberRoutes(r, queries, hub, m, frontendURL)
			handlers.RegisterTicketRoutes(r, queries, hub, ticketSvc)
		})
	})

	// WebSocket endpoint
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	go func() {
		slog.Info("Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("Server exited")
}
