package documents

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
	id, project_id, uploaded_by, filename, COALESCE(file_type, ''), file_size, storage_path, COALESCE(category, ''), uploaded_at
`

func scan(row interface{ Scan(dest ...any) error }) (*Document, error) {
	d := &Document{}
	err := row.Scan(
		&d.ID, &d.ProjectID, &d.UploadedBy, &d.Filename, &d.FileType,
		&d.FileSize, &d.StoragePath, &d.Category, &d.UploadedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDocumentNotFound
	}
	return d, err
}

func (r *PostgresRepository) Create(ctx context.Context, d *Document) error {
	if d.UploadedAt.IsZero() {
		d.UploadedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO project_documents (id, project_id, uploaded_by, filename, file_type, file_size, storage_path, category, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.ProjectID, d.UploadedBy, d.Filename, d.FileType,
		d.FileSize, d.StoragePath, d.Category, d.UploadedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, projectID, documentID string) (*Document, error) {
	query := `SELECT ` + columns + ` FROM project_documents WHERE project_id = $1 AND id = $2`
	return scan(r.db.QueryRowContext(ctx, query, projectID, documentID))
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]*Document, error) {
	query := `SELECT ` + columns + ` FROM project_documents WHERE project_id = $1 ORDER BY uploaded_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*Document, 0)
	for rows.Next() {
		d, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Delete(ctx context.Context, projectID, documentID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM project_documents WHERE project_id = $1 AND id = $2`, projectID, documentID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrDocumentNotFound
	}
	return nil
}
