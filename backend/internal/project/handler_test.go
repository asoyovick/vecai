package project

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/auth"
)

func setupProjectRouter(repo Repository) (*gin.Engine, *Service) {
	gin.SetMode(gin.TestMode)
	svc := NewService(repo)
	h := NewHandler(svc)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(auth.OptionalAuthMiddleware(auth.SecretFromEnv()))
	h.RegisterRoutes(api)
	return r, svc
}

func authHeader(t *testing.T, id string, role auth.Role) string {
	t.Helper()
	token, err := auth.GenerateToken(auth.User{ID: id, Email: id + "@example.com", Role: role})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	return "Bearer " + token
}

func TestProjectRouteAccess(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	repo := &fakeRepo{
		projects: map[string]*Project{
			"p1": {ID: "p1", OwnerID: "owner", Name: "P1"},
		},
		members: []Member{
			{ID: "m1", ProjectID: "p1", UserID: "member", Role: "engineer"},
		},
	}
	router, _ := setupProjectRouter(repo)

	type tc struct {
		name       string
		method     string
		path       string
		body       string
		user       string
		wantStatus int
	}
	tests := []tc{
		// GET project
		{name: "get owner", method: http.MethodGet, path: "/api/v1/projects/p1", user: "owner", wantStatus: http.StatusOK},
		{name: "get member", method: http.MethodGet, path: "/api/v1/projects/p1", user: "member", wantStatus: http.StatusOK},
		{name: "get stranger", method: http.MethodGet, path: "/api/v1/projects/p1", user: "stranger", wantStatus: http.StatusForbidden},
		// PUT project
		{name: "put owner", method: http.MethodPut, path: "/api/v1/projects/p1", body: `{"name":"New"}`, user: "owner", wantStatus: http.StatusOK},
		{name: "put member", method: http.MethodPut, path: "/api/v1/projects/p1", body: `{"name":"New"}`, user: "member", wantStatus: http.StatusForbidden},
		{name: "put stranger", method: http.MethodPut, path: "/api/v1/projects/p1", body: `{"name":"New"}`, user: "stranger", wantStatus: http.StatusForbidden},
		// DELETE project
		{name: "delete member", method: http.MethodDelete, path: "/api/v1/projects/p1", user: "member", wantStatus: http.StatusForbidden},
		{name: "delete stranger", method: http.MethodDelete, path: "/api/v1/projects/p1", user: "stranger", wantStatus: http.StatusForbidden},
		// Members
		{name: "list members owner", method: http.MethodGet, path: "/api/v1/projects/p1/members", user: "owner", wantStatus: http.StatusOK},
		{name: "list members member", method: http.MethodGet, path: "/api/v1/projects/p1/members", user: "member", wantStatus: http.StatusOK},
		{name: "list members stranger", method: http.MethodGet, path: "/api/v1/projects/p1/members", user: "stranger", wantStatus: http.StatusForbidden},
		{name: "invite member as member", method: http.MethodPost, path: "/api/v1/projects/p1/members", body: `{"user_id":"u3","role":"engineer"}`, user: "member", wantStatus: http.StatusForbidden},
		{name: "invite member as stranger", method: http.MethodPost, path: "/api/v1/projects/p1/members", body: `{"user_id":"u3","role":"engineer"}`, user: "stranger", wantStatus: http.StatusForbidden},
		// Blueprints/specs
		{name: "get blueprints owner", method: http.MethodGet, path: "/api/v1/projects/p1/blueprints", user: "owner", wantStatus: http.StatusOK},
		{name: "get blueprints member", method: http.MethodGet, path: "/api/v1/projects/p1/blueprints", user: "member", wantStatus: http.StatusOK},
		{name: "get blueprints stranger", method: http.MethodGet, path: "/api/v1/projects/p1/blueprints", user: "stranger", wantStatus: http.StatusForbidden},
		{name: "get specs stranger", method: http.MethodGet, path: "/api/v1/projects/p1/specs", user: "stranger", wantStatus: http.StatusForbidden},
		{name: "post blueprint member", method: http.MethodPost, path: "/api/v1/projects/p1/blueprints", body: `{"title":"B","file_path":"f"}`, user: "member", wantStatus: http.StatusCreated},
		{name: "post blueprint stranger", method: http.MethodPost, path: "/api/v1/projects/p1/blueprints", body: `{"title":"B","file_path":"f"}`, user: "stranger", wantStatus: http.StatusForbidden},
		{name: "post spec member", method: http.MethodPost, path: "/api/v1/projects/p1/specs", body: `{"title":"S","section_code":"A1"}`, user: "member", wantStatus: http.StatusCreated},
		{name: "post spec stranger", method: http.MethodPost, path: "/api/v1/projects/p1/specs", body: `{"title":"S","section_code":"A1"}`, user: "stranger", wantStatus: http.StatusForbidden},
		{name: "remove member stranger", method: http.MethodDelete, path: "/api/v1/projects/p1/members/m1", user: "stranger", wantStatus: http.StatusForbidden},
		{name: "remove member as member", method: http.MethodDelete, path: "/api/v1/projects/p1/members/m1", user: "member", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			req.Header.Set("Authorization", authHeader(t, tt.user, auth.RoleOwner))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestProjectRouteAccess_DeleteOwner(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	repo := &fakeRepo{projects: map[string]*Project{"p1": {ID: "p1", OwnerID: "owner", Name: "P1"}}, members: []Member{}}
	router, _ := setupProjectRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/p1", nil)
	req.Header.Set("Authorization", authHeader(t, "owner", auth.RoleOwner))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want=200 body=%s", w.Code, w.Body.String())
	}
}
