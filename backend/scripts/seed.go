package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/LCmaster/NextUp/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/nextup?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	log.Println("Seeding database...")

	// 1. Create a User
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		FirstName:    "Demo",
		LastName:     "User",
		Email:        "demo@example.com",
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		log.Printf("User might already exist: %v", err)
		user, _ = queries.GetUserByEmail(ctx, "demo@example.com")
	}

	if user.ID.Valid {
		log.Printf("Created User: %s %s (ID: %s)", user.FirstName, user.LastName, user.ID.String())
	}

	// 2. Create a Project
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		Name:        "NextUp Refactor",
		Description: pgtype.Text{String: "Migrating to new architecture", Valid: true},
		OwnerID:     user.ID,
	})
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	log.Printf("Created Project: %s (ID: %s)", project.Name, project.ID.String())

	// 3. Add User as Owner
	_, _ = queries.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    user.ID,
		Role:      "owner",
	})

	// 4. Create Tickets
	tickets := []db.CreateTicketParams{
		{
			ProjectID:   project.ID,
			Title:       "Set up Database seeding",
			Description: pgtype.Text{String: "Write a Go script to populate local DB.", Valid: true},
			Status:      "done",
			Priority:    "high",
			AssigneeID:  user.ID,
			CreatorID:   user.ID,
		},
		{
			ProjectID:   project.ID,
			Title:       "Implement Soft Deletes",
			Description: pgtype.Text{String: "Add deleted_at columns.", Valid: true},
			Status:      "in_progress",
			Priority:    "medium",
			AssigneeID:  user.ID,
			CreatorID:   user.ID,
		},
		{
			ProjectID:   project.ID,
			Title:       "End-to-End Tests",
			Description: pgtype.Text{String: "Use Playwright for frontend tests.", Valid: true},
			Status:      "todo",
			Priority:    "high",
			CreatorID:   user.ID,
		},
	}

	for _, t := range tickets {
		_, err := queries.CreateTicket(ctx, t)
		if err != nil {
			log.Printf("Failed to create ticket %q: %v", t.Title, err)
		}
	}

	log.Println("Database seeded successfully!")
}
