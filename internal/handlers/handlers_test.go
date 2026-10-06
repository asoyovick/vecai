package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"vecai/internal/config"
	"vecai/internal/db"
	"vecai/internal/handlers"
	"vecai/internal/routes"
)

func newTestServer(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"), "../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	cfg := &config.Config{DatabasePath: filepath.Join(dir, "test.db"), UploadsDir: filepath.Join(dir, "uploads")}
	return routes.New(handlers.New(database, cfg)), database
}

func login(t *testing.T, h http.Handler) []*http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": "david.ochieng@vecai.co.ke", "password": "demo1234"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()
}

func authed(t *testing.T, method, url string, body any, cookies []*http.Cookie) *http.Request {
	t.Helper()
	var r *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		r = httptest.NewRequest(method, url, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, url, nil)
	}
	for _, c := range cookies {
		r.AddCookie(c)
	}
	return r
}

func TestLoginAndMe(t *testing.T) {
	h, _ := newTestServer(t)
	cookies := login(t, h)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authed(t, "GET", "/api/auth/me", nil, cookies))
	if rec.Code != 200 {
		t.Fatalf("me: %d", rec.Code)
	}
}

func TestUnauthenticatedRequestsRejected(t *testing.T) {
	h, _ := newTestServer(t)
	for _, url := range []string{"/api/projects", "/api/dashboard", "/api/tasks", "/api/team"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
		if rec.Code != 401 {
			t.Fatalf("%s: expected 401, got %d", url, rec.Code)
		}
	}
}

func TestProjectLifecycle(t *testing.T) {
	h, _ := newTestServer(t)
	cookies := login(t, h)

	// Create (admin)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authed(t, "POST", "/api/projects", map[string]any{"name": "Test Residence", "status": "Design", "budgetTotal": 1000000}, cookies))
	if rec.Code != 201 {
		t.Fatalf("create project: %d %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	json.Unmarshal(rec.Body.Bytes(), &created)
	project := created["project"].(map[string]any)
	id := project["id"].(float64)

	// Read back
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/projects/1", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("get project: %d", rec2.Code)
	}

	// Patch progress
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("PATCH", "/api/projects/1", bytes.NewReader(mustJSON(map[string]any{"progress": 40})))
	req3.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req3.AddCookie(c)
	}
	h.ServeHTTP(rec3, req3)
	if rec3.Code != 200 {
		t.Fatalf("patch: %d %s", rec3.Code, rec3.Body.String())
	}

	// Invalid status rejected
	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("PATCH", "/api/projects/1", bytes.NewReader(mustJSON(map[string]any{"status": "Bogus"})))
	req4.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req4.AddCookie(c)
	}
	h.ServeHTTP(rec4, req4)
	if rec4.Code != 400 {
		t.Fatalf("expected 400, got %d", rec4.Code)
	}
	_ = id
}

func TestViewerCannotCreateTask(t *testing.T) {
	h, _ := newTestServer(t)
	admin := login(t, h)
	// Invite a viewer
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/team/invite", bytes.NewReader(mustJSON(map[string]any{"name": "View Only", "email": "viewer@example.com", "jobTitle": "Auditor", "role": "viewer"})))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range admin {
		req.AddCookie(c)
	}
	h.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("invite: %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	pw := resp["temporaryPassword"].(string)

	rec2 := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"email": "viewer@example.com", "password": pw})
	req2 := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("viewer login: %d", rec2.Code)
	}
	viewerCookies := rec2.Result().Cookies()

	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/tasks", bytes.NewReader(mustJSON(map[string]any{"projectId": 1, "name": "x", "dueDate": "2026-11-01"})))
	req3.Header.Set("Content-Type", "application/json")
	for _, c := range viewerCookies {
		req3.AddCookie(c)
	}
	h.ServeHTTP(rec3, req3)
	if rec3.Code != 403 {
		t.Fatalf("expected 403 for viewer, got %d", rec3.Code)
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
