package project

import (
	"context"
	"testing"
	"time"
)

type fakeRepo struct {
	projects map[string]*Project
	members  []Member
}

func (f *fakeRepo) CreateProject(_ context.Context, p *Project) error {
	f.projects[p.ID] = p
	return nil
}
func (f *fakeRepo) GetProjectByID(_ context.Context, id string) (*Project, error) {
	p, ok := f.projects[id]
	if !ok {
		return nil, ErrProjectNotFound
	}
	return p, nil
}
func (f *fakeRepo) ListProjectsByOwner(_ context.Context, ownerID string) ([]*Project, error) {
	var out []*Project
	for _, p := range f.projects {
		if p.OwnerID == ownerID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (f *fakeRepo) ListProjectsByUser(_ context.Context, userID string) ([]*Project, error) {
	seen := map[string]bool{}
	var out []*Project
	for _, p := range f.projects {
		if p.OwnerID == userID {
			seen[p.ID] = true
			out = append(out, p)
		}
	}
	for _, m := range f.members {
		if m.UserID == userID && !seen[m.ProjectID] {
			if p, ok := f.projects[m.ProjectID]; ok {
				seen[p.ID] = true
				out = append(out, p)
			}
		}
	}
	return out, nil
}
func (f *fakeRepo) ListAllProjects(_ context.Context) ([]*Project, error) { return nil, nil }
func (f *fakeRepo) UpdateProject(_ context.Context, p *Project) error {
	f.projects[p.ID] = p
	return nil
}
func (f *fakeRepo) DeleteProject(_ context.Context, id string) error {
	delete(f.projects, id)
	return nil
}
func (f *fakeRepo) AddBlueprint(context.Context, *Blueprint) error { return nil }
func (f *fakeRepo) GetBlueprintsByProject(context.Context, string) ([]*Blueprint, error) {
	return nil, nil
}
func (f *fakeRepo) AddSpec(context.Context, *Spec) error { return nil }
func (f *fakeRepo) GetSpecsByProject(context.Context, string) ([]*Spec, error) {
	return nil, nil
}
func (f *fakeRepo) AddMember(_ context.Context, m *Member) error {
	f.members = append(f.members, *m)
	return nil
}
func (f *fakeRepo) GetMembersByProject(context.Context, string) ([]*Member, error) { return nil, nil }
func (f *fakeRepo) GetMemberByID(context.Context, string, string) (*Member, error) {
	return nil, ErrMemberNotFound
}
func (f *fakeRepo) GetMemberByUserID(_ context.Context, projectID, userID string) (*Member, error) {
	for i := range f.members {
		if f.members[i].ProjectID == projectID && f.members[i].UserID == userID {
			m := f.members[i]
			return &m, nil
		}
	}
	return nil, ErrMemberNotFound
}
func (f *fakeRepo) IsMember(_ context.Context, projectID, userID string) (bool, error) {
	for _, m := range f.members {
		if m.ProjectID == projectID && m.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}
func (f *fakeRepo) DeleteMember(context.Context, string, string) error { return nil }

func TestListProjects_IncludesOwnedAndMemberProjects(t *testing.T) {
	repo := &fakeRepo{
		projects: map[string]*Project{
			"p-owned":  {ID: "p-owned", OwnerID: "user-1", Name: "Owned", CreatedAt: time.Now()},
			"p-member": {ID: "p-member", OwnerID: "user-2", Name: "Member", CreatedAt: time.Now().Add(-time.Hour)},
			"p-other":  {ID: "p-other", OwnerID: "user-2", Name: "Other", CreatedAt: time.Now().Add(-2 * time.Hour)},
		},
		members: []Member{
			{ProjectID: "p-member", UserID: "user-1", Role: "viewer"},
		},
	}
	svc := NewService(repo)

	got, err := svc.ListProjects(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	ids := map[string]bool{}
	for _, p := range got {
		ids[p.ID] = true
	}
	if !ids["p-owned"] || !ids["p-member"] {
		t.Fatalf("expected owned and member projects, got %#v", ids)
	}
	if ids["p-other"] {
		t.Fatalf("did not expect unrelated project")
	}
}
