package boq

import (
	"errors"
	"time"
)

var (
	ErrItemNotFound = errors.New("BOQ item not found")
	ErrInvalidInput = errors.New("invalid BOQ input")
)

// Item mirrors the boq_items table from docs/database.md.
type Item struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	ItemName      string    `json:"item_name"`
	Quantity      float64   `json:"quantity"`
	Unit          string    `json:"unit"`
	EstimatedCost float64   `json:"estimated_cost"`
	ActualCost    float64   `json:"actual_cost"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Summary aggregates estimated vs actual cost for the dashboard.
type Summary struct {
	TotalItems     int     `json:"total_items"`
	TotalEstimated float64 `json:"total_estimated_cost"`
	TotalActual    float64 `json:"total_actual_cost"`
	Variance       float64 `json:"variance"`
}
