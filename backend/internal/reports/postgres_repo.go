package reports

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

const columns = `
	id, project_id, user_id, COALESCE(TO_CHAR(report_date, 'YYYY-MM-DD'), ''),
	COALESCE(completed_work, ''), workers_present, COALESCE(remarks, ''), created_at
`

func scan(row interface{ Scan(dest ...any) error }) (*Report, error) {
	r := &Report{}
	err := row.Scan(
		&r.ID, &r.ProjectID, &r.UserID, &r.ReportDate,
		&r.CompletedWork, &r.WorkersPresent, &r.Remarks, &r.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReportNotFound
	}
	return r, err
}

func (r *PostgresRepository) Create(ctx context.Context, report *Report) error {
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO daily_reports (id, project_id, user_id, report_date, completed_work, workers_present, remarks, created_at)
		VALUES ($1, $2, $3, $4::DATE, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.ProjectID, report.UserID, report.ReportDate,
		report.CompletedWork, report.WorkersPresent, report.Remarks, report.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, projectID, reportID string) (*Report, error) {
	query := `SELECT ` + columns + ` FROM daily_reports WHERE project_id = $1 AND id = $2`
	return scan(r.db.QueryRowContext(ctx, query, projectID, reportID))
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]*Report, error) {
	query := `SELECT ` + columns + ` FROM daily_reports WHERE project_id = $1 ORDER BY report_date DESC, created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*Report, 0)
	for rows.Next() {
		report, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, report)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Delete(ctx context.Context, projectID, reportID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM daily_reports WHERE project_id = $1 AND id = $2`, projectID, reportID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrReportNotFound
	}
	return nil
}
