package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LCmaster/NextUp/internal/db"
	"github.com/LCmaster/NextUp/internal/services"
	"github.com/LCmaster/NextUp/internal/ws"
)

func TestProjectService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, cleanup, err := setupTestDB(ctx)
	if err != nil {
		t.Fatalf("setupTestDB failed: %v", err)
	}
	defer cleanup()

	// 1. Manually run migrations (simplified for test: create projects and project_members tables)
	// A real project might use golang-migrate or embedded sqlc schema here.
	schema := `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    github_link VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);
`
	if _, err := pool.Exec(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	queries := db.New(pool)
	hub := ws.NewHub()
	ps := services.NewProjectService(queries, hub)

	// Create a user first
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		FirstName:    "Test",
		LastName:     "User",
		Email:        "test@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 2. Test CreateProject
	proj, err := ps.CreateProject(ctx, "Integration Project", "Description", user.ID)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if proj.Name != "Integration Project" {
		t.Errorf("expected Integration Project, got %s", proj.Name)
	}

	// 3. Test ListProjects (should return 1)
	projects, err := ps.ListProjects(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	// 4. Test UpdateProject
	proj, err = ps.UpdateProject(ctx, proj.ID, user.ID, "Updated Project", "Updated Desc")
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if proj.Name != "Updated Project" {
		t.Errorf("expected Updated Project, got %s", proj.Name)
	}

	// 5. Test DeleteProject (Soft Delete)
	err = ps.DeleteProject(ctx, proj.ID, user.ID)
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify soft delete by checking ListProjects again
	projectsAfterDelete, err := ps.ListProjects(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListProjects after delete failed: %v", err)
	}
	if len(projectsAfterDelete) != 0 {
		t.Errorf("expected 0 projects after delete, got %d", len(projectsAfterDelete))
	}

	// Verify it still exists in DB but with deleted_at set
	var deletedAt pgtype.Timestamptz
	err = pool.QueryRow(ctx, "SELECT deleted_at FROM projects WHERE id = $1", proj.ID).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("failed to query deleted project: %v", err)
	}
	if !deletedAt.Valid {
		t.Errorf("expected deleted_at to be valid, got invalid")
	}
}
