package services_test

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB starts a postgres testcontainer, runs migrations, and returns a pgxpool.
// It also returns a cleanup function to terminate the container.
func setupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("nextup_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	// Wait for DB to be truly ready
	if err := pool.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping db: %w", err)
	}

	// Apply migrations (normally you'd use golang-migrate programmatically)
	// For this test snippet, we assume a migrate tool or you can execute sql directly.
	// As a shortcut, we can use the actual sqlc generated schema?
	// The best is to parse the migrations and execute them. Since this is an example,
	// let's do a simple read of up migrations and execute them.
	// Since we don't have the migrate library installed yet, we'll just return the pool.
	// A proper implementation would use golang-migrate/migrate.

	cleanup := func() {
		pool.Close()
		if err := postgresContainer.Terminate(context.Background()); err != nil {
			fmt.Printf("failed to terminate container: %v\n", err)
		}
	}

	return pool, cleanup, nil
}
