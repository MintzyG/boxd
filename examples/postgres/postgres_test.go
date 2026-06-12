package postgres_test

import (
	"testing"

	"github.com/MintzyG/boxd/examples/postgres"
)

func TestPostgres(t *testing.T) {
	pg := postgres.Run(t)
	db := pg.DB(t)

	if _, err := db.Exec(`CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`INSERT INTO users (name) VALUES ($1), ($2)`, "alice", "bob"); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 users, got %d", count)
	}

	pg.Clean(t, db, "users")

	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 users after clean, got %d", count)
	}
}
