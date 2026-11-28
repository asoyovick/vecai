package project

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/auth"
)

type fakeRepo struct {
	projects map[string]*Project
	members  map[string]*Member // key: projectID|userID
}

func (f *fakeRepo) CreateProject(context.Context, *Project) error { return nil }
func (f *fakeRepo) GetProjectByID(_ context.Context, id string) (*Project, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, ErrProjectNotFound
	}
	return p, nil
}
func (f *fakeRepo) ListProjectsByOwner(context.Context, string) ([]*Project, error) { return nil, nil }
func (f *fakeRepo) ListAllProjects(context.Context) ([]*Project, error)             { return nil, nil }
func (f *fakeRepo) UpdateProject(context.Context, *Project) error                   { return nil }
func (f *fakeRepo) DeleteProject(context.Context, string) error                     { return nil }
func (f *fakeRepo) AddBlueprint(context.Context, *Blueprint) error                  { return nil }
func (f *fakeRepo) GetBlueprintsByProject(context.Context, string) ([]*Blueprint, error) {
	return nil, nil
}
func (f *fakeRepo) AddSpec(context.Context, *Spec) error                           { return nil }
func (f *fakeRepo) GetSpecsByProject(context.Context, string) ([]*Spec, error)     { return nil, nil }
func (f *fakeRepo) AddMember(context.Context, *Member) error                       { return nil }
func (f *fakeRepo) GetMembersByProject(context.Context, string) ([]*Member, error) { return nil, nil }
func (f *fakeRepo) GetMemberByID(context.Context, string, string) (*Member, error) {
	return nil, ErrMemberNotFound
}
func (f *fakeRepo) GetMemberByUserID(_ context.Context, projectID, userID string) (*Member, error) {
	m, ok := f.members[projectID+"|"+userID]
	if !ok {
		return nil, ErrMemberNotFound
	}
	return m, nil
}
func (f *fakeRepo) DeleteMember(context.Context, string, string) error { return nil }

func TestAuthorizeProjectAccess(t *testing.T) {
	repo := &fakeRepo{
		projects: map[string]*Project{"p1": {ID: "p1", OwnerID: "owner"}},
		members:  map[string]*Member{"p1|member": {ID: "m1", ProjectID: "p1", UserID: "member"}},
	}
	svc := NewService(repo)

	if _, err := svc.AuthorizeProjectAccess(context.Background(), "p1", "owner"); err != nil {
		t.Fatalf("owner should have access: %v", err)
	}
	if _, err := svc.AuthorizeProjectAccess(context.Background(), "p1", "member"); err != nil {
		t.Fatalf("member should have access: %v", err)
	}
	if _, err := svc.AuthorizeProjectAccess(context.Background(), "p1", "stranger"); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("stranger should be denied, got: %v", err)
	}
	if _, err := svc.AuthorizeProjectAccess(context.Background(), "missing", "owner"); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("missing project should be not found, got: %v", err)
	}
}

func TestRequireProjectAccessMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeRepo{
		projects: map[string]*Project{"p1": {ID: "p1", OwnerID: "owner"}},
		members:  map[string]*Member{"p1|member": {ID: "m1", ProjectID: "p1", UserID: "member"}},
	}
	svc := NewService(repo)

	tests := []struct {
		name   string
		userID string
		want   int
	}{
		{name: "owner", userID: "owner", want: http.StatusOK},
		{name: "member", userID: "member", want: http.StatusOK},
		{name: "stranger", userID: "stranger", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				c.Set(string(auth.UserIDContextKey), tt.userID)
				c.Set(string(auth.UserRoleContextKey), auth.RoleOwner)
				c.Next()
			})
			r.GET("/projects/:id", RequireProjectAccess(svc), func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/projects/p1", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("want %d, got %d: %s", tt.want, w.Code, w.Body.String())
			}
		})
	}

	t.Run("missing project", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(string(auth.UserIDContextKey), "owner")
			c.Set(string(auth.UserRoleContextKey), auth.RoleOwner)
			c.Next()
		})
		r.GET("/projects/:id", RequireProjectAccess(svc), func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/projects/missing", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}
