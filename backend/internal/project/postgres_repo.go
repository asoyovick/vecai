package project

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

const projectColumns = `
	id, owner_id, name, COALESCE(description, ''), COALESCE(location, ''),
	COALESCE(client_name, ''), status, COALESCE(TO_CHAR(start_date, 'YYYY-MM-DD'), ''),
	COALESCE(TO_CHAR(expected_end_date, 'YYYY-MM-DD'), ''), created_at, updated_at
`

func scanProject(row interface{ Scan(dest ...any) error }) (*Project, error) {
	p := &Project{}
	err := row.Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Location,
		&p.ClientName, &p.Status, &p.StartDate, &p.EndDate, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	return p, err
}

func (r *PostgresRepository) CreateProject(ctx context.Context, p *Project) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = p.CreatedAt
	if p.Status == "" {
		p.Status = "Planning"
	}

	query := `
		INSERT INTO projects (id, owner_id, name, description, location, client_name, status, start_date, expected_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, '')::DATE, NULLIF($9, '')::DATE, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.OwnerID, p.Name, p.Description, p.Location, p.ClientName,
		p.Status, p.StartDate, p.EndDate, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetProjectByID(ctx context.Context, id string) (*Project, error) {
	query := `SELECT ` + projectColumns + ` FROM projects WHERE id = $1`
	return scanProject(r.db.QueryRowContext(ctx, query, id))
}

func (r *PostgresRepository) ListProjectsByOwner(ctx context.Context, ownerID string) ([]*Project, error) {
	query := `SELECT ` + projectColumns + ` FROM projects WHERE owner_id = $1 ORDER BY created_at DESC`
	return r.queryProjects(ctx, query, ownerID)
}

func (r *PostgresRepository) ListAllProjects(ctx context.Context) ([]*Project, error) {
	query := `SELECT ` + projectColumns + ` FROM projects ORDER BY created_at DESC`
	return r.queryProjects(ctx, query)
}

func (r *PostgresRepository) queryProjects(ctx context.Context, query string, args ...any) ([]*Project, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]*Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *PostgresRepository) UpdateProject(ctx context.Context, p *Project) error {
	query := `
		UPDATE projects
		SET name = $2, description = NULLIF($3, ''), location = NULLIF($4, ''),
		    client_name = NULLIF($5, ''), status = $6, start_date = NULLIF($7, '')::DATE,
		    expected_end_date = NULLIF($8, '')::DATE, updated_at = $9
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Location, p.ClientName,
		p.Status, p.StartDate, p.EndDate, time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteProject(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *PostgresRepository) AddBlueprint(ctx context.Context, b *Blueprint) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO blueprints (id, project_id, title, description, file_path, version, uploader_id, created_at, updated_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.ProjectID, b.Title, b.Description, b.FilePath, b.Version, b.UploaderID, b.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetBlueprintsByProject(ctx context.Context, projectID string) ([]*Blueprint, error) {
	query := `
		SELECT id, project_id, title, COALESCE(description, ''), file_path, version, uploader_id, created_at
		FROM blueprints
		WHERE project_id = $1
		ORDER BY version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blueprints := make([]*Blueprint, 0)
	for rows.Next() {
		b := &Blueprint{}
		if err := rows.Scan(&b.ID, &b.ProjectID, &b.Title, &b.Description, &b.FilePath, &b.Version, &b.UploaderID, &b.CreatedAt); err != nil {
			return nil, err
		}
		blueprints = append(blueprints, b)
	}
	return blueprints, rows.Err()
}

func (r *PostgresRepository) AddSpec(ctx context.Context, s *Spec) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO specs (id, project_id, title, description, section_code, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.ProjectID, s.Title, s.Description, s.SectionCode, s.CreatedBy, s.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetSpecsByProject(ctx context.Context, projectID string) ([]*Spec, error) {
	query := `
		SELECT id, project_id, title, COALESCE(description, ''), section_code, created_by, created_at
		FROM specs
		WHERE project_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	specs := make([]*Spec, 0)
	for rows.Next() {
		s := &Spec{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Title, &s.Description, &s.SectionCode, &s.CreatedBy, &s.CreatedAt); err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, rows.Err()
}

func (r *PostgresRepository) AddMember(ctx context.Context, m *Member) error {
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO project_members (id, project_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.ProjectID, m.UserID, m.Role, m.JoinedAt)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return ErrDuplicateProject // reused as "already a member"
		case "23503": // foreign_key_violation
			return ErrProjectNotFound
		}
	}
	return err
}

const memberColumns = `
	pm.id, pm.project_id, pm.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), pm.role, pm.joined_at
`

func scanMember(row interface{ Scan(dest ...any) error }) (*Member, error) {
	m := &Member{}
	err := row.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMemberNotFound
	}
	return m, err
}

func (r *PostgresRepository) GetMembersByProject(ctx context.Context, projectID string) ([]*Member, error) {
	query := `
		SELECT ` + memberColumns + `
		FROM project_members pm
		LEFT JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*Member, 0)
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *PostgresRepository) GetMemberByID(ctx context.Context, projectID, memberID string) (*Member, error) {
	query := `
		SELECT ` + memberColumns + `
		FROM project_members pm
		LEFT JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1 AND pm.id = $2
	`
	return scanMember(r.db.QueryRowContext(ctx, query, projectID, memberID))
}

func (r *PostgresRepository) DeleteMember(ctx context.Context, projectID, memberID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM project_members WHERE project_id = $1 AND id = $2`, projectID, memberID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrMemberNotFound
	}
	return nil
}
