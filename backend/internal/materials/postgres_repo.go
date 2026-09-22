package materials

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
	id, project_id, material_name, quantity, unit, unit_cost,
	COALESCE(supplier, ''), COALESCE(TO_CHAR(delivery_date, 'YYYY-MM-DD'), ''), recorded_by, created_at
`

func scan(row interface{ Scan(dest ...any) error }) (*Material, error) {
	m := &Material{}
	err := row.Scan(
		&m.ID, &m.ProjectID, &m.MaterialName, &m.Quantity, &m.Unit, &m.UnitCost,
		&m.Supplier, &m.DeliveryDate, &m.RecordedBy, &m.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMaterialNotFound
	}
	return m, err
}

func (r *PostgresRepository) Create(ctx context.Context, m *Material) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO materials (id, project_id, material_name, quantity, unit, unit_cost, supplier, delivery_date, recorded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, '')::DATE, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.ProjectID, m.MaterialName, m.Quantity, m.Unit, m.UnitCost,
		m.Supplier, m.DeliveryDate, m.RecordedBy, m.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, projectID, materialID string) (*Material, error) {
	query := `SELECT ` + columns + ` FROM materials WHERE project_id = $1 AND id = $2`
	return scan(r.db.QueryRowContext(ctx, query, projectID, materialID))
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]*Material, error) {
	query := `SELECT ` + columns + ` FROM materials WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*Material, 0)
	for rows.Next() {
		m, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, m *Material) error {
	query := `
		UPDATE materials
		SET material_name = $3, quantity = $4, unit = $5, unit_cost = $6,
		    supplier = NULLIF($7, ''), delivery_date = NULLIF($8, '')::DATE
		WHERE project_id = $1 AND id = $2
	`
	result, err := r.db.ExecContext(ctx, query,
		m.ProjectID, m.ID, m.MaterialName, m.Quantity, m.Unit, m.UnitCost,
		m.Supplier, m.DeliveryDate,
	)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrMaterialNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, projectID, materialID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM materials WHERE project_id = $1 AND id = $2`, projectID, materialID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrMaterialNotFound
	}
	return nil
}
