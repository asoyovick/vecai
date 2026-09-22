package materials

import (
	"errors"
	"time"
)

var (
	ErrMaterialNotFound = errors.New("material record not found")
	ErrInvalidInput     = errors.New("invalid material input")
)

// Material mirrors the materials table from docs/database.md.
type Material struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	MaterialName string    `json:"material_name"`
	Quantity     float64   `json:"quantity"`
	Unit         string    `json:"unit"`
	UnitCost     float64   `json:"unit_cost"`
	Supplier     string    `json:"supplier"`
	DeliveryDate string    `json:"delivery_date"`
	RecordedBy   string    `json:"recorded_by"`
	CreatedAt    time.Time `json:"created_at"`
}
