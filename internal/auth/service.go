package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/mail"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("Incorrect email or password.")
	ErrEmailTaken         = errors.New("A team member with that email already exists.")
	ErrCurrentPassword    = errors.New("Your current password is incorrect.")
)

// Service contains the authentication and user-management business rules.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

const demoColors = 8

func (s *Service) Login(email, password string) (*User, string, error) {
	row, err := s.repo.FindUserByEmail(email)
	if err != nil || !VerifyPassword(password, row.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}
	return s.issueSession(row.User)
}

func (s *Service) Register(name, email, password, jobTitle string) (*User, string, error) {
	if _, err := validateRegistration(name, email, password); err != nil {
		return nil, "", err
	}
	taken, err := s.repo.EmailInUse(email, 0)
	if err != nil {
		return nil, "", err
	}
	if taken {
		return nil, "", ErrEmailTaken
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}
	id, err := s.repo.CreateUser(email, hash, name, "member", jobTitle, initialsFor(name), "#1e3a5f", "")
	if err != nil {
		return nil, "", err
	}
	row, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, "", err
	}
	return s.issueSession(row.User)
}

func (s *Service) issueSession(u User) (*User, string, error) {
	token, err := newToken()
	if err != nil {
		return nil, "", err
	}
	if err := s.repo.InsertSession(tokenHash(token), u.ID, sessionExpiry()); err != nil {
		return nil, "", err
	}
	return &u, token, nil
}

func sessionExpiry() string {
	return time.Now().Add(sessionTTL).UTC().Format("2006-01-02 15:04:05")
}

func (s *Service) Logout(token string) {
	if token != "" {
		s.repo.DeleteSession(tokenHash(token))
	}
}

func (s *Service) UpdateProfile(u *User, name, email, phone, jobTitle string) (*UserRow, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("Enter a valid email address.")
	}
	taken, err := s.repo.EmailInUse(email, u.ID)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, errors.New("That email address is already in use.")
	}
	if err := s.repo.UpdateUserProfile(u.ID, name, email, phone, jobTitle, initialsFor(name)); err != nil {
		return nil, err
	}
	return s.repo.FindUserByID(u.ID)
}

func (s *Service) ChangePassword(u *User, currentPassword, newPassword string) error {
	if err := ValidateNewPassword(newPassword); err != nil {
		return err
	}
	row, err := s.repo.FindUserByID(u.ID)
	if err != nil || !VerifyPassword(currentPassword, row.PasswordHash) {
		return ErrCurrentPassword
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateUserPassword(u.ID, hash); err != nil {
		return err
	}
	return s.repo.DeleteSessionsForUser(u.ID)
}

var defaultPrefs = map[string]bool{
	"material-alerts": true, "task-reminders": true, "site-updates": true,
	"document-uploads": false, "budget-thresholds": true, "weekly-summary": true,
}

func (s *Service) GetPreferences(userID int64) (map[string]bool, error) {
	prefs := map[string]bool{}
	for k, v := range defaultPrefs {
		prefs[k] = v
	}
	raw, err := s.repo.GetPreferences(userID)
	if err == nil {
		var stored map[string]bool
		if json.Unmarshal([]byte(raw), &stored) == nil {
			for k, v := range stored {
				prefs[k] = v
			}
		}
	}
	return prefs, nil
}

func (s *Service) PatchPreferences(userID int64, body map[string]any) (map[string]bool, error) {
	next := map[string]bool{}
	raw, err := s.repo.GetPreferences(userID)
	if err == nil {
		json.Unmarshal([]byte(raw), &next)
	}
	for key := range defaultPrefs {
		if v, ok := body[key].(bool); ok {
			next[key] = v
		}
	}
	data, _ := json.Marshal(next)
	if err := s.repo.UpsertPreferences(userID, string(data)); err != nil {
		return nil, err
	}
	for k, v := range defaultPrefs {
		if _, ok := next[k]; !ok {
			next[k] = v
		}
	}
	return next, nil
}

func (s *Service) GetOrganisation() (map[string]any, error) {
	return s.repo.GetOrganisation()
}

func (s *Service) UpdateOrganisation(name, businessType, country, currency string) error {
	return s.repo.UpdateOrganisation(name, businessType, country, currency)
}

func (s *Service) Team() ([]map[string]any, []map[string]any, error) {
	members, err := s.repo.ListTeamMembers()
	if err != nil {
		return nil, nil, err
	}
	if members == nil {
		members = []map[string]any{}
	}
	assignments, err := s.repo.ListProjectAssignments()
	if err != nil {
		return nil, nil, err
	}
	if assignments == nil {
		assignments = []map[string]any{}
	}
	return members, assignments, nil
}

// InviteMember creates an account with a temporary password (no SMTP in this deploy).
func (s *Service) InviteMember(admin *User, name, email, jobTitle, role, phone string, projectID *int64) (map[string]any, string, error) {
	taken, err := s.repo.EmailInUse(email, 0)
	if err != nil {
		return nil, "", err
	}
	if taken {
		return nil, "", ErrEmailTaken
	}
	tokenBuf := make([]byte, 4)
	rand.Read(tokenBuf)
	tempPassword := "vecai-" + hex.EncodeToString(tokenBuf)
	hash, err := HashPassword(tempPassword)
	if err != nil {
		return nil, "", err
	}
	id, err := s.repo.CreateUser(email, hash, name, role, jobTitle, initialsFor(name), inviteColor(), phone)
	if err != nil {
		return nil, "", err
	}
	if projectID != nil {
		if exists, _ := s.repo.ProjectExists(*projectID); exists {
			s.repo.AddProjectMember(*projectID, id)
			s.repo.LogActivity(*projectID, admin.ID, admin.Name, "project", "Invited "+name+" to the project team")
		}
	}
	row, err := s.repo.FindUserByID(id)
	if err != nil {
		return nil, "", err
	}
	return row.publicMap(), tempPassword, nil
}

var inviteColors = []string{"#1e3a5f", "#0891b2", "#7c3aed", "#db2777", "#059669", "#d97706", "#dc2626", "#0d9488"}

func inviteColor() string {
	b := make([]byte, 1)
	rand.Read(b)
	return inviteColors[int(b[0])%len(inviteColors)]
}
