package project

import (
	"database/sql"
	"strconv"
	"strings"

	"vecai/internal/db"
)

// Postgres is the PostgreSQL-backed project/task repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

const taskSelect = `SELECT t.*, p.name AS project_name, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
	FROM tasks t JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = t.assignee_id`

func (p *Postgres) ListProjects() ([]map[string]any, error) {
	return db.QueryMaps(p.database, "SELECT * FROM projects ORDER BY created_at DESC")
}

func (p *Postgres) GetProject(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, "SELECT * FROM projects WHERE id = $1", id)
}

func (p *Postgres) CreateProject(name, location, status string, startDate, expectedCompletion any, client, contractor string, budgetTotal int64, description string) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO projects (name, location, status, start_date, expected_completion, client, contractor, budget_total, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		name, location, status, startDate, expectedCompletion, client, contractor, budgetTotal, description).Scan(&id)
	return id, err
}

func (p *Postgres) UpdateProject(id int64, updates []FieldUpdate) error {
	args := []any{}
	sets := []string{}
	for i, u := range updates {
		sets = append(sets, u.Column+" = $"+strconv.Itoa(i+1))
		args = append(args, u.Value)
	}
	args = append(args, id)
	_, err := p.database.Exec(`UPDATE projects SET `+strings.Join(sets, ", ")+`, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $`+strconv.Itoa(len(args)), args...)
	return err
}

func (p *Postgres) AddProjectMember(projectID int64, userID int64) error {
	_, err := p.database.Exec(`INSERT INTO project_members (project_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, projectID, userID)
	return err
}

func (p *Postgres) BudgetMonths(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, "SELECT ym, planned, actual FROM budget_months WHERE project_id = $1 ORDER BY ym", projectID)
}

func (p *Postgres) TasksByProject(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT t.*, u.name AS assignee_name, u.initials AS assignee_initials, u.color AS assignee_color
		FROM tasks t LEFT JOIN users u ON u.id = t.assignee_id WHERE t.project_id = $1 ORDER BY t.due_date IS NULL, t.due_date, t.id`, projectID)
}

func (p *Postgres) MaterialsByProject(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT m.*, s.name AS supplier_name FROM materials m LEFT JOIN suppliers s ON s.id = m.supplier_id
		WHERE m.project_id = $1 ORDER BY m.name`, projectID)
}

func (p *Postgres) DocumentsByProject(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT d.*, u.name AS uploaded_by_name FROM project_documents d LEFT JOIN users u ON u.id = d.uploaded_by
		WHERE d.project_id = $1 ORDER BY d.created_at DESC`, projectID)
}

func (p *Postgres) ActivityByProject(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, "SELECT * FROM activity_log WHERE project_id = $1 ORDER BY created_at DESC, id DESC LIMIT 30", projectID)
}

func (p *Postgres) ProjectTeam(projectID int64) ([]map[string]any, error) {
	return db.QueryMaps(p.database, `SELECT u.id, u.name, u.role, u.job_title AS roleLabel, u.initials, u.color, u.email, u.phone
		FROM project_members pm JOIN users u ON u.id = pm.user_id WHERE pm.project_id = $1 ORDER BY u.name`, projectID)
}

func (p *Postgres) ProjectStats(projectID int64) (int, int64, int64, int64, error) {
	var total int
	var completed, dueWeek, blocked sql.NullInt64
	err := p.database.QueryRow(`SELECT COUNT(*),
		SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END),
		SUM(CASE WHEN status != 'completed' AND due_date IS NOT NULL AND due_date <= to_char((now() AT TIME ZONE 'UTC')::date + INTERVAL '7 days', 'YYYY-MM-DD') THEN 1 ELSE 0 END),
		SUM(CASE WHEN status = 'blocked' THEN 1 ELSE 0 END)
		FROM tasks WHERE project_id = $1`, projectID).Scan(&total, &completed, &dueWeek, &blocked)
	return total, completed.Int64, dueWeek.Int64, blocked.Int64, err
}

func (p *Postgres) MaterialAlerts(projectID int64) (int, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM materials WHERE project_id = $1 AND status IN ('critical','low')`, projectID).Scan(&c)
	return c, err
}

func (p *Postgres) RecentDocuments(projectID int64) (int, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM project_documents WHERE project_id = $1 AND created_at >= to_char(now() AT TIME ZONE 'UTC' - INTERVAL '14 days', 'YYYY-MM-DD HH24:MI:SS')`, projectID).Scan(&c)
	return c, err
}

func (p *Postgres) TouchProject(projectID int64) error {
	_, err := p.database.Exec(`UPDATE projects SET updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $1`, projectID)
	return err
}

func (p *Postgres) ListTasks(projectID any, status any, q any) ([]map[string]any, error) {
	sqlText := taskSelect
	args := []any{}
	conds := []string{}
	i := 1
	if projectID != nil {
		conds = append(conds, "t.project_id = $"+strconv.Itoa(i))
		args = append(args, projectID)
		i++
	}
	if status != nil {
		conds = append(conds, "t.status = $"+strconv.Itoa(i))
		args = append(args, status)
		i++
	}
	if q != nil {
		conds = append(conds, "t.name LIKE $"+strconv.Itoa(i))
		args = append(args, q)
		i++
	}
	if len(conds) > 0 {
		sqlText += " WHERE " + strings.Join(conds, " AND ")
	}
	sqlText += " ORDER BY t.due_date IS NULL, t.due_date, t.id"
	return db.QueryMaps(p.database, sqlText, args...)
}

func (p *Postgres) GetTask(id int64) (map[string]any, error) {
	return db.QueryMap(p.database, taskSelect+" WHERE t.id = $1", id)
}

func (p *Postgres) CreateTask(projectID int64, assigneeID any, name string, description string, startDate, dueDate any, status, priority string) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO tasks (project_id, assignee_id, name, description, start_date, due_date, status, priority) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		projectID, assigneeID, name, description, startDate, dueDate, status, priority).Scan(&id)
	return id, err
}

func (p *Postgres) UpdateTask(id int64, updates []FieldUpdate) error {
	args := []any{}
	sets := []string{}
	for i, u := range updates {
		sets = append(sets, u.Column+" = $"+strconv.Itoa(i+1))
		args = append(args, u.Value)
	}
	args = append(args, id)
	_, err := p.database.Exec(`UPDATE tasks SET `+strings.Join(sets, ", ")+`, updated_at = to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS') WHERE id = $`+strconv.Itoa(len(args)), args...)
	return err
}

func (p *Postgres) ProjectExists(id int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = $1`, id).Scan(&c)
	return c > 0, err
}

func (p *Postgres) UserExists(id int64) (bool, error) {
	var c int
	err := p.database.QueryRow(`SELECT COUNT(*) FROM users WHERE id = $1`, id).Scan(&c)
	return c > 0, err
}

func (p *Postgres) LogActivity(projectID any, actorID any, name, typ, text string) {
	p.database.Exec(`INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES ($1, $2, $3, $4, $5)`,
		projectID, actorID, name, typ, text)
}
