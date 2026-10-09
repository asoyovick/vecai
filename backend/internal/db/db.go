// Package db owns the PostgreSQL connection and the migration runner.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// Pool bundles the connection pool so wiring code passes one value.
type Pool struct{ DB *sql.DB }

// Close releases the pool's connections.
func (p *Pool) Close() error { return p.DB.Close() }

// Open opens a PostgreSQL connection pool.
func Open(databaseURL string) (*Pool, error) {
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	return &Pool{DB: database}, nil
}

// Migrate applies db/migrations/*.up.sql files in filename order, tracking
// applied versions in schema_migrations.
func Migrate(database *sql.DB, dir string) error {
	if _, err := database.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS'))
	)`); err != nil {
		return err
	}
	applied := map[string]bool{}
	rows, err := database.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	for _, f := range files {
		version := strings.TrimSuffix(f, ".up.sql")
		if applied[version] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return err
		}
		tx, err := database.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(data)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", f, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// SeedIfEmpty loads demo data when the users table is empty.
func SeedIfEmpty(database *sql.DB, path string) error {
	var c int
	if err := database.QueryRow("SELECT COUNT(*) FROM users").Scan(&c); err != nil {
		return err
	}
	if c > 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read seed: %w", err)
	}
	if _, err := database.Exec(string(data)); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	return nil
}

// QueryMaps runs a query and returns rows as maps of JSON-friendly values.
func QueryMaps(database *sql.DB, q string, args ...any) ([]map[string]any, error) {
	rows, err := database.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := map[string]any{}
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				m[c] = string(b)
			} else {
				m[c] = vals[i]
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// QueryMap returns the first row as a map, or sql.ErrNoRows.
func QueryMap(database *sql.DB, q string, args ...any) (map[string]any, error) {
	rows, err := QueryMaps(database, q, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}
