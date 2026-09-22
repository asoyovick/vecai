package materials

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create records a material delivery against a project.
func (s *Service) Create(ctx context.Context, projectID, userID string, req CreateRequest) (*Material, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	m := &Material{
		ID:           uuid.NewString(),
		ProjectID:    projectID,
		MaterialName: strings.TrimSpace(req.MaterialName),
		Quantity:     req.Quantity,
		Unit:         strings.TrimSpace(req.Unit),
		UnitCost:     req.UnitCost,
		Supplier:     strings.TrimSpace(req.Supplier),
		DeliveryDate: req.DeliveryDate,
		RecordedBy:   userID,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("failed to record material: %w", err)
	}
	return m, nil
}

func (s *Service) Get(ctx context.Context, projectID, materialID string) (*Material, error) {
	return s.repo.GetByID(ctx, projectID, materialID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]*Material, error) {
	return s.repo.ListByProject(ctx, projectID)
}

// Update applies a partial update to a material record.
func (s *Service) Update(ctx context.Context, projectID, materialID string, req UpdateRequest) (*Material, error) {
	existing, err := s.repo.GetByID(ctx, projectID, materialID)
	if err != nil {
		return nil, err
	}

	if req.MaterialName != nil {
		if strings.TrimSpace(*req.MaterialName) == "" {
			return nil, fmt.Errorf("%w: material_name cannot be empty", ErrInvalidInput)
		}
		existing.MaterialName = strings.TrimSpace(*req.MaterialName)
	}
	if req.Quantity != nil {
		if *req.Quantity < 0 {
			return nil, fmt.Errorf("%w: quantity cannot be negative", ErrInvalidInput)
		}
		existing.Quantity = *req.Quantity
	}
	if req.Unit != nil {
		existing.Unit = strings.TrimSpace(*req.Unit)
	}
	if req.UnitCost != nil {
		if *req.UnitCost < 0 {
			return nil, fmt.Errorf("%w: unit_cost cannot be negative", ErrInvalidInput)
		}
		existing.UnitCost = *req.UnitCost
	}
	if req.Supplier != nil {
		existing.Supplier = strings.TrimSpace(*req.Supplier)
	}
	if req.DeliveryDate != nil {
		if *req.DeliveryDate != "" && !isValidDate(*req.DeliveryDate) {
			return nil, fmt.Errorf("%w: delivery_date must be YYYY-MM-DD", ErrInvalidInput)
		}
		existing.DeliveryDate = *req.DeliveryDate
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, projectID, materialID string) error {
	return s.repo.Delete(ctx, projectID, materialID)
}

// Summary returns aggregate inventory figures for a project dashboard.
type Summary struct {
	TotalRecords  int     `json:"total_records"`
	TotalQuantity float64 `json:"total_quantity"`
	TotalCost     float64 `json:"total_cost"`
}

func (s *Service) Summarize(items []*Material) Summary {
	summary := Summary{TotalRecords: len(items)}
	for _, m := range items {
		summary.TotalQuantity += m.Quantity
		summary.TotalCost += m.Quantity * m.UnitCost
	}
	return summary
}

func validate(req CreateRequest) error {
	if strings.TrimSpace(req.MaterialName) == "" {
		return fmt.Errorf("%w: material_name is required", ErrInvalidInput)
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(req.Unit) == "" {
		return fmt.Errorf("%w: unit is required", ErrInvalidInput)
	}
	if req.UnitCost < 0 {
		return fmt.Errorf("%w: unit_cost cannot be negative", ErrInvalidInput)
	}
	if req.DeliveryDate != "" && !isValidDate(req.DeliveryDate) {
		return fmt.Errorf("%w: delivery_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	return nil
}

func isValidDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
