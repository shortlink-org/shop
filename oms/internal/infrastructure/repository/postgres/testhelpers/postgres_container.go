//go:build integration

package testhelpers

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer holds the container and connection pool for tests.
type PostgresContainer struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
	ConnStr   string
}

// SetupPostgresContainer creates a PostgreSQL container for integration tests.
// It returns a PostgresContainer with an active connection pool and a cleanup function.
func SetupPostgresContainer(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to ping database: %v", err)
	}

	pc := &PostgresContainer{
		Container: container,
		Pool:      pool,
		ConnStr:   connStr,
	}

	// Register cleanup
	t.Cleanup(func() {
		pool.Close()
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	return pc
}

// RunMigrations executes the provided SQL migration statements.
func (pc *PostgresContainer) RunMigrations(t *testing.T, migrations ...string) {
	t.Helper()

	ctx := context.Background()
	for i, migration := range migrations {
		if _, err := pc.Pool.Exec(ctx, migration); err != nil {
			t.Fatalf("failed to run migration %d: %v", i+1, err)
		}
	}
}

// TestDB implements db.DB interface for testing purposes.
type TestDB struct {
	pool *pgxpool.Pool
}

// NewTestDB creates a new TestDB wrapper around a pgxpool.Pool.
func NewTestDB(pool *pgxpool.Pool) *TestDB {
	return &TestDB{pool: pool}
}

// Init is a no-op for test database (connection is already established).
func (t *TestDB) Init(_ context.Context) error {
	return nil
}

// GetConn returns the underlying connection pool.
func (t *TestDB) GetConn() any {
	return t.pool
}

// DB returns a TestDB that implements db.DB interface.
func (pc *PostgresContainer) DB() *TestDB {
	return NewTestDB(pc.Pool)
}
