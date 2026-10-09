package reports

import (
	"database/sql"

	"vecai/internal/db"
)

// Postgres is the PostgreSQL-backed reports repository.
type Postgres struct {
	database *sql.DB
}

func NewPostgres(database *sql.DB) *Postgres { return &Postgres{database: database} }

func (p *Postgres) Query(q string, args ...any) ([]map[string]any, error) {
	return db.QueryMaps(p.database, q, args...)
}

func (p *Postgres) QueryMap(q string, args ...any) (map[string]any, error) {
	return db.QueryMap(p.database, q, args...)
}

func (p *Postgres) CreateDailyReport(projectID int64, submittedBy int64, date, weather, temperature string, workers int64, workCompleted, issues, notes string) (int64, error) {
	var id int64
	err := p.database.QueryRow(
		`INSERT INTO daily_reports (project_id, submitted_by, date, weather, temperature, workers_on_site, work_completed, issues, notes, photo_count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0) RETURNING id`,
		projectID, submittedBy, date, weather, temperature, workers, workCompleted, issues, notes).Scan(&id)
	return id, err
}

func (p *Postgres) CreateDailyReportMaterial(reportID int64, name string, quantity float64, unit string) error {
	_, err := p.database.Exec(`INSERT INTO daily_report_materials (daily_report_id, name, quantity, unit) VALUES ($1, $2, $3, $4)`, reportID, name, quantity, unit)
	return err
}

func (p *Postgres) ProjectName(projectID int64) (string, error) {
	var name string
	err := p.database.QueryRow(`SELECT name FROM projects WHERE id = $1`, projectID).Scan(&name)
	return name, err
}

func (p *Postgres) LogActivity(projectID any, actorID any, name, typ, text string) {
	p.database.Exec(`INSERT INTO activity_log (project_id, actor_id, actor_name, type, text) VALUES ($1, $2, $3, $4, $5)`,
		projectID, actorID, name, typ, text)
}
