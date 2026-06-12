package postgres

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/MintzyG/boxd"
	_ "github.com/lib/pq"
)

const (
	defaultUser     = "test"
	defaultPassword = "pass"
	defaultDB       = "testdb"
)

// PostgresContainer wraps a boxd Container with Postgres-specific helpers.
type PostgresContainer struct {
	*boxd.Container
	User     string
	Password string
	Database string
}

// Run starts a Postgres 16 container with sane defaults.
// Additional boxd options can be passed to override env, ports, wait strategy, etc.
func Run(t *testing.T, opts ...boxd.Option) *PostgresContainer {
	t.Helper()

	defaults := []boxd.Option{
		boxd.WithImage("postgres:16"),
		boxd.WithEnv("POSTGRES_USER", defaultUser),
		boxd.WithEnv("POSTGRES_PASSWORD", defaultPassword),
		boxd.WithEnv("POSTGRES_DB", defaultDB),
		boxd.WithHealthCheck(boxd.HealthCheck{
			Test:     []string{"CMD-SHELL", "pg_isready -U " + defaultUser},
			Interval: 2 * time.Second,
			Timeout:  1 * time.Second,
			Retries:  5,
		}),
		boxd.WithPort("5432/tcp"),
		boxd.WithWait(boxd.WaitForHealthy(30 * time.Second)),
	}

	c := boxd.Run(t, append(defaults, opts...)...)

	return &PostgresContainer{
		Container: c,
		User:      defaultUser,
		Password:  defaultPassword,
		Database:  defaultDB,
	}
}

// ConnStr returns the DSN for connecting to the container.
func (c *PostgresContainer) ConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Ports["5432/tcp"], c.Database,
	)
}

// DB opens and returns a *sql.DB connected to the container.
// The connection is closed automatically when the test ends.
func (c *PostgresContainer) DB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", c.ConnStr())
	if err != nil {
		t.Fatal("open db:", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatal("ping db:", err)
	}
	return db
}

// Clean truncates the given tables and restarts their sequences.
// Useful for resetting state between subtests.
func (c *PostgresContainer) Clean(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("clean %q: %v", table, err)
		}
	}
}
