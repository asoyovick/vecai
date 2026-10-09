package auth

import (
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for service tests.
type fakeRepo struct {
	users         map[string]*UserRow
	byID          map[int64]*UserRow
	byToken       map[string]*UserRow
	sessions      map[string]int64
	nextID        int64
	emailInUse    bool
	profileErr    error
	prefs         map[int64]string
	organisation  map[string]any
	organisationE error
	members       []map[string]any
	assignments   []map[string]any
	deletedAll    bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:        map[string]*UserRow{},
		byID:         map[int64]*UserRow{},
		byToken:      map[string]*UserRow{},
		sessions:     map[string]int64{},
		prefs:        map[int64]string{},
		nextID:       1,
		organisation: map[string]any{"name": "VECAI", "business_type": "Construction Company", "country": "Kenya", "currency": "KES"},
	}
}

func (f *fakeRepo) FindUserByEmail(email string) (*UserRow, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, errors.New("no rows")
	}
	return u, nil
}

func (f *fakeRepo) FindUserByID(id int64) (*UserRow, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, errors.New("no rows")
	}
	return u, nil
}

func (f *fakeRepo) EmailInUse(email string, excludeID int64) (bool, error) {
	u, ok := f.users[email]
	return ok && u.ID != excludeID, nil
}

func (f *fakeRepo) CreateUser(email, passwordHash, name, role, jobTitle, initials, color, phone string) (int64, error) {
	f.nextID++
	row := &UserRow{User: User{ID: f.nextID, Email: email, Name: name, Role: role, JobTitle: jobTitle, Initials: initials, Color: color, Phone: phone}, PasswordHash: passwordHash}
	f.users[email] = row
	f.byID[row.ID] = row
	return row.ID, nil
}

func (f *fakeRepo) UpdateUserProfile(id int64, name, email, phone, jobTitle, initials string) error {
	if f.profileErr != nil {
		return f.profileErr
	}
	row := f.byID[id]
	row.Name, row.Email, row.Phone, row.JobTitle, row.Initials = name, email, phone, jobTitle, initials
	return nil
}

func (f *fakeRepo) UpdateUserPassword(id int64, passwordHash string) error {
	f.byID[id].PasswordHash = passwordHash
	return nil
}

func (f *fakeRepo) InsertSession(tokenHash string, userID int64, expiresAt string) error {
	f.sessions[tokenHash] = userID
	return nil
}

func (f *fakeRepo) DeleteSession(tokenHash string) error {
	delete(f.sessions, tokenHash)
	return nil
}

func (f *fakeRepo) DeleteSessionsForUser(userID int64) error {
	for k, v := range f.sessions {
		if v == userID {
			delete(f.sessions, k)
		}
	}
	f.deletedAll = true
	return nil
}

func (f *fakeRepo) FindUserBySessionToken(tokenHash string) (*UserRow, error) {
	uid, ok := f.sessions[tokenHash]
	if !ok {
		return nil, errors.New("no rows")
	}
	return f.byID[uid], nil
}

func (f *fakeRepo) ListTeamMembers() ([]map[string]any, error) { return f.members, nil }

func (f *fakeRepo) ListProjectAssignments() ([]map[string]any, error) {
	return f.assignments, nil
}

func (f *fakeRepo) AddProjectMember(projectID, userID int64) error { return nil }

func (f *fakeRepo) ProjectExists(projectID int64) (bool, error) { return projectID == 1, nil }

func (f *fakeRepo) GetPreferences(userID int64) (string, error) {
	raw, ok := f.prefs[userID]
	if !ok {
		return "", errors.New("no rows")
	}
	return raw, nil
}

func (f *fakeRepo) UpsertPreferences(userID int64, prefs string) error {
	f.prefs[userID] = prefs
	return nil
}

func (f *fakeRepo) GetOrganisation() (map[string]any, error) {
	return f.organisation, f.organisationE
}

func (f *fakeRepo) UpdateOrganisation(name, businessType, country, currency string) error {
	f.organisation["name"] = name
	return nil
}

func (f *fakeRepo) LogActivity(projectID, actorID any, name, typ, text string) {}

func seedUser(f *fakeRepo, email, password, role string) *UserRow {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	row := &UserRow{User: User{ID: f.nextID, Email: email, Name: "Test User", Role: role}, PasswordHash: hash}
	f.nextID++
	f.users[email] = row
	f.byID[row.ID] = row
	return row
}

func TestServiceLogin(t *testing.T) {
	repo := newFakeRepo()
	seedUser(repo, "a@b.co", "demo1234", "admin")
	svc := NewService(repo)

	u, token, err := svc.Login("a@b.co", "demo1234")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if u.Email != "a@b.co" || token == "" {
		t.Fatalf("unexpected login result: %+v", u)
	}
	if _, _, err := svc.Login("a@b.co", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v", err)
	}
	if _, _, err := svc.Login("missing@b.co", "demo1234"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown email error = %v", err)
	}
}

func TestServiceRegister(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	u, token, err := svc.Register("New User", "new@b.co", "passw0rd1", "Planner")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.Role != "member" || u.Initials != "NU" || token == "" {
		t.Fatalf("unexpected registration result: %+v", u)
	}
	if _, _, err := svc.Register("Dupe", "new@b.co", "passw0rd1", ""); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate registration error = %v", err)
	}
	if _, _, err := svc.Register("Bad", "new@b.co", "short", ""); err == nil {
		t.Fatal("weak password accepted")
	}
}

func TestServiceChangePassword(t *testing.T) {
	repo := newFakeRepo()
	row := seedUser(repo, "a@b.co", "oldpass1", "member")
	svc := NewService(repo)

	if err := svc.ChangePassword(&row.User, "wrong", "newpass1"); !errors.Is(err, ErrCurrentPassword) {
		t.Fatalf("wrong current password error = %v", err)
	}
	if err := svc.ChangePassword(&row.User, "oldpass1", "weak"); err == nil {
		t.Fatal("weak new password accepted")
	}
	if err := svc.ChangePassword(&row.User, "oldpass1", "newpass1"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if !VerifyPassword("newpass1", repo.byID[row.ID].PasswordHash) {
		t.Fatal("stored hash does not match the new password")
	}
	if !repo.deletedAll {
		t.Fatal("changing the password must revoke existing sessions")
	}
}

func TestServiceUpdateProfile(t *testing.T) {
	repo := newFakeRepo()
	row := seedUser(repo, "a@b.co", "demo1234", "member")
	svc := NewService(repo)

	if _, err := svc.UpdateProfile(&row.User, "Renamed", "bad-email", "", ""); err == nil {
		t.Fatal("invalid email accepted")
	}
	updated, err := svc.UpdateProfile(&row.User, "Renamed User", "renamed@b.co", "+254700000000", "Lead")
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Name != "Renamed User" || updated.Initials != "RU" || updated.Phone != "+254700000000" {
		t.Fatalf("unexpected profile: %+v", updated)
	}
}

func TestServicePreferences(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	prefs, err := svc.GetPreferences(9)
	if err != nil {
		t.Fatalf("get preferences: %v", err)
	}
	if !prefs["material-alerts"] || prefs["document-uploads"] {
		t.Fatalf("unexpected defaults: %+v", prefs)
	}
	next, err := svc.PatchPreferences(9, map[string]any{"document-uploads": true, "unknown": true})
	if err != nil {
		t.Fatalf("patch preferences: %v", err)
	}
	if !next["document-uploads"] || !next["material-alerts"] {
		t.Fatalf("unexpected patched preferences: %+v", next)
	}
	stored, err := svc.GetPreferences(9)
	if err != nil || !stored["document-uploads"] {
		t.Fatalf("preferences not persisted: %+v (%v)", stored, err)
	}
}

func TestServiceInviteMember(t *testing.T) {
	repo := newFakeRepo()
	admin := seedUser(repo, "admin@b.co", "demo1234", "admin")
	svc := NewService(repo)

	member, tempPassword, err := svc.InviteMember(&admin.User, "Grace Wanjiru", "grace@b.co", "Surveyor", "member", "", nil)
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if member["initials"] != "GW" || tempPassword == "" {
		t.Fatalf("unexpected invite result: %+v / %q", member, tempPassword)
	}
	if _, _, err := svc.InviteMember(&admin.User, "Dup", "grace@b.co", "Surveyor", "member", "", nil); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate invite error = %v", err)
	}
}

func TestServiceOrganisation(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)

	row, err := svc.GetOrganisation()
	if err != nil || row["name"] != "VECAI" {
		t.Fatalf("organisation: %+v (%v)", row, err)
	}
	if err := svc.UpdateOrganisation("VECAI Ltd", "Construction Company", "Kenya", "KES"); err != nil {
		t.Fatalf("update organisation: %v", err)
	}
	if row, _ = svc.GetOrganisation(); row["name"] != "VECAI Ltd" {
		t.Fatalf("organisation not updated: %+v", row)
	}
}
