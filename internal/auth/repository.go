package auth

// Repository is the persistence contract the auth service depends on.
type Repository interface {
	FindUserByEmail(email string) (*UserRow, error)
	FindUserByID(id int64) (*UserRow, error)
	EmailInUse(email string, excludeID int64) (bool, error)
	CreateUser(email string, passwordHash string, name string, role string, jobTitle string, initials string, color string, phone string) (int64, error)
	UpdateUserProfile(id int64, name string, email string, phone string, jobTitle string, initials string) error
	UpdateUserPassword(id int64, passwordHash string) error
	InsertSession(tokenHash string, userID int64, expiresAt string) error
	DeleteSession(tokenHash string) error
	DeleteSessionsForUser(userID int64) error
	FindUserBySessionToken(tokenHash string) (*UserRow, error)
	ListTeamMembers() ([]map[string]any, error)
	ListProjectAssignments() ([]map[string]any, error)
	AddProjectMember(projectID int64, userID int64) error
	ProjectExists(projectID int64) (bool, error)
	GetPreferences(userID int64) (string, error)
	UpsertPreferences(userID int64, prefs string) error
	GetOrganisation() (map[string]any, error)
	UpdateOrganisation(name string, businessType string, country string, currency string) error
	LogActivity(projectID any, actorID any, name string, typ string, text string)
}
