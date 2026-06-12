// Integration test against a real Postgres container using boxd.Run directly.
// Not picked up by the normal test run (underscore prefix excludes this directory
// from the Go build system). Run explicitly with: go test ./...
// from inside this directory.
package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/MintzyG/boxd"
	_ "github.com/lib/pq"
)

func TestPostgres(t *testing.T) {
	c := boxd.Run(t,
		boxd.WithImage("postgres:16"),
		boxd.WithLogs(boxd.LogAlways),
		boxd.WithEnv("POSTGRES_PASSWORD", "pass"),
		boxd.WithHealthCheck(boxd.HealthCheck{
			Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
			Interval: 2 * time.Second,
			Timeout:  1 * time.Second,
			Retries:  5,
		}),
		boxd.WithPort("5432/tcp"),
		boxd.WithWait(boxd.WaitForHealthy(30*time.Second)),
	)

	connStr := fmt.Sprintf("postgres://postgres:pass@%s:%s/postgres?sslmode=disable", c.Host, c.Port("5432"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatal(err)
	}

	t.Log("got:", result)
}
