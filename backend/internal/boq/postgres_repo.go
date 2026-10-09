// Package boq exposes the bill-of-quantities for each project.
package boq

import (
	"database/sql"

	"strconv"
	"strings"

	"vecai/internal/db"
)

// Repository is the persistence contract for BOQ items.
type Repository interface {
	ListItems(projectID any) ([]map[string]any, error)
	GetItem(id int64) (map[string]any, error)
	CreateItem(projectID int64, itemNo, description, unit string, quantity float64, unitRate int64) (int64, error)
	UpdateItem(id int64, updates []FieldUpdate) error
	DeleteItem(id int64) error
	ProjectExists(id int64) (bool, error)
	TouchProject(projectID int64) error
	LogActivity(projectID any, actorID any, name, typ, text string)
}

// Postgres is the PostgreSQL-backed BOQ repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

func (p *Postgres) ListItems(projectID any) ([]map[string]any, error) {
	q := `SELECT id, project_id, item_no, description, unit, quantity, unit_rate,
		(quantity * unit_rate)::bigint AS amount, created_at, updated_at
		FROM boq_items`
	args := []any{}
	if projectID != nil {
		q += " WHERE project_id = $1"
		args = append(args, projectID)
	}
	q += " ORDER BY id"
	return db.QueryMaps(p.database, q, args...)
}

func (p *Postgres) GetItem(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, `SELECT id, project_id, item_no, description, unit, quantity, unit_rate,
		(quantity * unit_rate)::bigint AS amount, created_at, updated_at FROM boq_items WHERE id = $1`, id)
}

func (p *Postgres) CreateItem(projectID int64, itemNo, description, unit string, quantity float64, unitRate int64) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO boq_items (project_id, item_no, description, unit, quantity, unit_rate) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		projectID, itemNo, description, unit, quantity, unitRate).Scan(&id)
	return id, err
}

func (p *Postgres) UpdateItem(id int64, updates []FieldUpdate) error {
	args := []any{}
	sets := []string{}
	for i, u := range updates {
		sets = append(sets, u.Column+" = $"+strconv.Itoa(i+1))
		args = append(args, u.Value)
	}
	args = append(args, id)
	_, err := p.database.Exec(`UPDATE boq_items SET `+strings.Join(sets, ", ")+`, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $`+strconv.Itoa(len(args)), args...)
	return err
}

func (p *Postgres) DeleteItem(id int64) error {
	_, err := p.database.Exec(`DELETE FROM boq_items WHERE id = $1`, id)
	return err
}

func (p *Postgres) ProjectExists(id int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = $1`, id).Scan(&c)
	return c > 0, err
}

func (p *Postgres) TouchProject(projectID int64) error {
	_, err := p.database.Exec(`UPDATE projects SET updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $1`, projectID)
	return err
}

func (p *Postgres) LogActivity(projectID any, actorID any, name, typ, text string) {
	p.database.Exec(`INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES ($1, $2, $3, $4, $5)`,
		projectID, actorID, name, typ, text)
}
