package db

import (
	"database/sql"
	"testing"
)

func TestPoolOpen(t *testing.T) {
	database, err := Open("postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	defer database.DB.Close()
	if err := database.DB.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestQueryMapsNormalisesValues(t *testing.T) {
	database, err := Open("postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	defer database.DB.Close()

	rows, err := QueryMaps(database.DB, "SELECT 1::bigint AS n, 'x' AS s, NULL AS nil")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d", len(rows))
	}
	if v, ok := rows[0]["n"].(int64); !ok || v != 1 {
		t.Fatalf("integer column scanned as %T (%v)", rows[0]["n"], rows[0]["n"])
	}
	if v, ok := rows[0]["s"].(string); !ok || v != "x" {
		t.Fatalf("text column scanned as %T (%v)", rows[0]["s"], rows[0]["s"])
	}
	if rows[0]["nil"] != nil {
		t.Fatalf("null column scanned as %T (%v)", rows[0]["nil"], rows[0]["nil"])
	}

	if _, err := QueryMap(database.DB, "SELECT 1 AS x WHERE false"); err != sql.ErrNoRows {
		t.Fatalf("empty QueryMap error = %v, want sql.ErrNoRows", err)
	}
}
