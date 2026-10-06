package documents

import (
	"database/sql"
	"strconv"
	"strings"

	"vecai/internal/db"
)

// Postgres is the PostgreSQL-backed documents repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

const docSelect = `SELECT d.*, p.name AS project_name, u.name AS uploaded_by_name
	FROM project_documents d JOIN projects p ON p.id = d.project_id LEFT JOIN users u ON u.id = d.uploaded_by`

func (p *Postgres) ListDocuments(projectID, category, q any) ([]map[string]any, error) {
	sqlText := docSelect
	args := []any{}
	conds := []string{}
	i := 1
	if projectID != nil {
		conds = append(conds, "d.project_id = $"+strconv.Itoa(i))
		args = append(args, projectID)
		i++
	}
	if category != nil {
		conds = append(conds, "d.category = $"+strconv.Itoa(i))
		args = append(args, category)
		i++
	}
	if q != nil {
		conds = append(conds, "d.name LIKE $"+strconv.Itoa(i))
		args = append(args, q)
		i++
	}
	if len(conds) > 0 {
		sqlText += " WHERE " + strings.Join(conds, " AND ")
	}
	sqlText += " ORDER BY d.created_at DESC"
	return db.QueryMaps(p.database, sqlText, args...)
}

func (p *Postgres) GetDocument(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, docSelect+" WHERE d.id = $1", id)
}

func (p *Postgres) CreateDocument(projectID int64, uploadedBy int64, name, category, version, originalName, storedName, mimeType string, sizeBytes int64, format string) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO project_documents (project_id, uploaded_by, name, category, version, original_name, stored_name, mime_type, size_bytes, format) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		projectID, uploadedBy, name, category, version, originalName, storedName, mimeType, sizeBytes, format).Scan(&id)
	return id, err
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
