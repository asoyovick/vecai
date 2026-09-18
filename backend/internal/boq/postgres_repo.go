package boq

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
	id, project_id, item_name, quantity, COALESCE(unit, ''), estimated_cost, actual_cost, created_at, updated_at
`

func scan(row interface{ Scan(dest ...any) error }) (*Item, error) {
	i := &Item{}
	err := row.Scan(
		&i.ID, &i.ProjectID, &i.ItemName, &i.Quantity, &i.Unit,
		&i.EstimatedCost, &i.ActualCost, &i.CreatedAt, &i.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrItemNotFound
	}
	return i, err
}

func (r *PostgresRepository) Create(ctx context.Context, item *Item) error {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	item.UpdatedAt = item.CreatedAt

	query := `
		INSERT INTO boq_items (id, project_id, item_name, quantity, unit, estimated_cost, actual_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ProjectID, item.ItemName, item.Quantity, item.Unit,
		item.EstimatedCost, item.ActualCost, item.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, projectID, itemID string) (*Item, error) {
	query := `SELECT ` + columns + ` FROM boq_items WHERE project_id = $1 AND id = $2`
	return scan(r.db.QueryRowContext(ctx, query, projectID, itemID))
}

func (r *PostgresRepository) ListByProject(ctx context.Context, projectID string) ([]*Item, error) {
	query := `SELECT ` + columns + ` FROM boq_items WHERE project_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*Item, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, item *Item) error {
	item.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE boq_items
		SET item_name = $3, quantity = $4, unit = $5, estimated_cost = $6, actual_cost = $7, updated_at = $8
		WHERE project_id = $1 AND id = $2
	`
	result, err := r.db.ExecContext(ctx, query,
		item.ProjectID, item.ID, item.ItemName, item.Quantity, item.Unit,
		item.EstimatedCost, item.ActualCost, item.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, projectID, itemID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM boq_items WHERE project_id = $1 AND id = $2`, projectID, itemID)
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err == nil && rows == 0 {
		return ErrItemNotFound
	}
	return nil
}
