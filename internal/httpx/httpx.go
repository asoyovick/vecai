// Package httpx holds the small helpers shared by every handler:
// JSON responses, API errors, validation and row scanning.
package httpx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string { return e.Message }

func Err(status int, msg string) *APIError { return &APIError{Status: status, Message: msg} }

func JSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		JSON(w, 500, map[string]string{"error": "failed to encode response"})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	w.Write(body)
}

func Error(w http.ResponseWriter, err error) {
	if ae, ok := err.(*APIError); ok {
		JSON(w, ae.Status, map[string]string{"error": ae.Message})
		return
	}
	JSON(w, 500, map[string]string{"error": "Something went wrong on the server. Please try again."})
}

// Decode reads a JSON request body with a 1 MB cap.
func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			return Err(413, "Request body too large.")
		}
		return Err(400, "Invalid JSON body.")
	}
	return nil
}

// ── Field validation (mirrors the original API's error messages) ─────────────

func AsString(value any, field string, required bool, max int) (string, error) {
	if value == nil {
		if required {
			return "", Err(400, field+" is required.")
		}
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", Err(400, field+" must be a string.")
	}
	s = strings.TrimSpace(s)
	if required && s == "" {
		return "", Err(400, field+" is required.")
	}
	if max > 0 && len(s) > max {
		return "", Err(400, fmt.Sprintf("%s must be at most %d characters.", field, max))
	}
	return s, nil
}

func AsNumber(value any, field string, required bool, min float64, integer bool) (float64, error) {
	if value == nil || value == "" {
		if required {
			return 0, Err(400, field+" is required.")
		}
		return 0, nil
	}
	var n float64
	switch v := value.(type) {
	case float64:
		n = v
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, Err(400, field+" must be a number.")
		}
		n = parsed
	default:
		return 0, Err(400, field+" must be a number.")
	}
	if integer && n != float64(int64(n)) {
		return 0, Err(400, field+" must be a whole number.")
	}
	if n < min {
		return 0, Err(400, fmt.Sprintf("%s must be at least %v.", field, min))
	}
	return n, nil
}

func AsOneOf(value any, field string, allowed []string) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", Err(400, fmt.Sprintf("%s must be one of: %s.", field, strings.Join(allowed, ", ")))
	}
	for _, a := range allowed {
		if s == a {
			return s, nil
		}
	}
	return "", Err(400, fmt.Sprintf("%s must be one of: %s.", field, strings.Join(allowed, ", ")))
}

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func AsDate(value any, field string, required bool) (*string, error) {
	if value == nil || value == "" {
		if required {
			return nil, Err(400, field+" is required.")
		}
		return nil, nil
	}
	s, ok := value.(string)
	if !ok || !dateRe.MatchString(s) {
		return nil, Err(400, field+" must be a date (YYYY-MM-DD).")
	}
	return &s, nil
}

var EmailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// ── Row scanning ─────────────────────────────────────────────────────────────

// QueryMaps runs a query and returns rows as maps, converting SQLite values
// to JSON-friendly types (strings, floats, nils).
func QueryMaps(db *sql.DB, q string, args ...any) ([]map[string]any, error) {
	rows, err := db.Query(q, args...)
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
			m[c] = normalize(vals[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func QueryMap(db *sql.DB, q string, args ...any) (map[string]any, error) {
	rows, err := QueryMaps(db, q, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	return rows[0], nil
}

func normalize(v any) any {
	switch x := v.(type) {
	case []byte:
		return string(x)
	case int64:
		return x
	case float64:
		return x
	case nil:
		return nil
	default:
		return x
	}
}

// ── Activity + project bookkeeping ───────────────────────────────────────────

type Actor struct {
	ID   int64
	Name string
}

func LogActivity(db *sql.DB, projectID any, actor *Actor, typ, text string) {
	var pid any
	if projectID != nil {
		pid = projectID
	}
	var aid any
	name := "System"
	if actor != nil {
		aid = actor.ID
		name = actor.Name
	}
	db.Exec("INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES (?, ?, ?, ?, ?)",
		pid, aid, name, typ, text)
}

func TouchProject(db *sql.DB, projectID int64) {
	db.Exec("UPDATE projects SET updated_at = datetime('now') WHERE id = ?", projectID)
}
