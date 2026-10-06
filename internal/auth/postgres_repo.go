package auth

import (
	"database/sql"

	"vecai/internal/db"
)

// Postgres is the PostgreSQL-backed user/session repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

func scanUser(row *sql.Row) (*UserRow, error) {
	var u UserRow
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.JobTitle, &u.Initials, &u.Color, &u.Phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

const userColumns = `id, email, password_hash, name, role, job_title, initials, color, phone`

func (p *Postgres) FindUserByEmail(email string) (*UserRow, error) {
	return scanUser(p.database.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = $1`, email))
}

func (p *Postgres) FindUserByID(id int64) (*UserRow, error) {
	return scanUser(p.database.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (p *Postgres) EmailInUse(email string, excludeID int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2`, email, excludeID).Scan(&c)
	return c > 0, err
}

func (p *Postgres) CreateUser(email string, passwordHash string, name string, role string, jobTitle string, initials string, color string, phone string) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO users (email, password_hash, name, role, job_title, initials, color, phone) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		email, passwordHash, name, role, jobTitle, initials, color, phone).Scan(&id)
	return id, err
}

func (p *Postgres) UpdateUserProfile(id int64, name string, email string, phone string, jobTitle string, initials string) error {
	_, err := p.database.Exec(
		`UPDATE users SET name = $1, email = $2, phone = $3, job_title = $4, initials = $5, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $6`,
		name, email, phone, jobTitle, initials, id)
	return err
}

func (p *Postgres) UpdateUserPassword(id int64, passwordHash string) error {
	_, err := p.database.Exec(
		`UPDATE users SET password_hash = $1, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $2`,
		passwordHash, id)
	return err
}

func (p *Postgres) InsertSession(tokenHash string, userID int64, expiresAt string) error {
	_, err := p.database.Exec(`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`, tokenHash, userID, expiresAt)
	return err
}

func (p *Postgres) DeleteSession(tokenHash string) error {
	_, err := p.database.Exec(`DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (p *Postgres) DeleteSessionsForUser(userID int64) error {
	_, err := p.database.Exec(`DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (p *Postgres) FindUserBySessionToken(tokenHash string) (*UserRow, error) {
	return scanUser(p.database.QueryRow(
		`SELECT u.id, u.email, u.password_hash, u.name, u.role, u.job_title, u.initials, u.color, u.phone
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')`, tokenHash))
}

func (p *Postgres) ListTeamMembers() ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT u.id, u.name, u.email, u.role, u.job_title, u.initials, u.color, u.phone, u.created_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.user_id = u.id) AS project_count
		FROM users u ORDER BY u.name`)
}

func (p *Postgres) ListProjectAssignments() ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT pm.user_id, p.id AS project_id, p.name AS project_name FROM project_members pm JOIN projects p ON p.id = pm.project_id`)
}

func (p *Postgres) AddProjectMember(projectID int64, userID int64) error {
	_, err := p.database.Exec(`INSERT INTO project_members (project_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, projectID, userID)
	return err
}

func (p *Postgres) ProjectExists(projectID int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = $1`, projectID).Scan(&c)
	return c > 0, err
}

func (p *Postgres) GetPreferences(userID int64) (string, error) {
	var raw string
	err := p.database.QueryRow(`SELECT prefs FROM user_preferences WHERE user_id = $1`, userID).Scan(&raw)
	return raw, err
}

func (p *Postgres) UpsertPreferences(userID int64, prefs string) error {
	_, err := p.database.Exec(`INSERT INTO user_preferences (user_id, prefs) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET prefs = excluded.prefs`, userID, prefs)
	return err
}

func (p *Postgres) GetOrganisation() (map[string]any, error) {
	return db.QueryMap(p.database, `SELECT name, business_type, country, currency FROM organisation WHERE id = 1`)
}

func (p *Postgres) UpdateOrganisation(name string, businessType string, country string, currency string) error {
	_, err := p.database.Exec(`UPDATE organisation SET name = $1, business_type = $2, country = $3, currency = $4 WHERE id = 1`, name, businessType, country, currency)
	return err
}

func (p *Postgres) LogActivity(projectID any, actorID any, name string, typ string, text string) {
	p.database.Exec(`INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES ($1, $2, $3, $4, $5)`,
		projectID, actorID, name, typ, text)
}
