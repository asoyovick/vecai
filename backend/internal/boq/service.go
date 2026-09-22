package boq

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a Bill of Quantities line item to a project.
func (s *Service) Create(ctx context.Context, projectID string, req ItemRequest) (*Item, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	item := &Item{
		ID:            uuid.NewString(),
		ProjectID:     projectID,
		ItemName:      strings.TrimSpace(req.ItemName),
		Quantity:      req.Quantity,
		Unit:          strings.TrimSpace(req.Unit),
		EstimatedCost: req.EstimatedCost,
		ActualCost:    req.ActualCost,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create BOQ item: %w", err)
	}
	return item, nil
}

func (s *Service) Get(ctx context.Context, projectID, itemID string) (*Item, error) {
	return s.repo.GetByID(ctx, projectID, itemID)
}

func (s *Service) List(ctx context.Context, projectID string) ([]*Item, error) {
	return s.repo.ListByProject(ctx, projectID)
}

// Update replaces the editable fields of a BOQ item.
func (s *Service) Update(ctx context.Context, projectID, itemID string, req ItemRequest) (*Item, error) {
	existing, err := s.repo.GetByID(ctx, projectID, itemID)
	if err != nil {
		return nil, err
	}

	if err := validate(req); err != nil {
		return nil, err
	}

	existing.ItemName = strings.TrimSpace(req.ItemName)
	existing.Quantity = req.Quantity
	existing.Unit = strings.TrimSpace(req.Unit)
	existing.EstimatedCost = req.EstimatedCost
	existing.ActualCost = req.ActualCost

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, projectID, itemID string) error {
	return s.repo.Delete(ctx, projectID, itemID)
}

// Compare returns estimated vs actual totals so cost drift is visible early.
func (s *Service) Compare(items []*Item) Summary {
	summary := Summary{TotalItems: len(items)}
	for _, item := range items {
		summary.TotalEstimated += item.EstimatedCost
		summary.TotalActual += item.ActualCost
	}
	summary.Variance = summary.TotalActual - summary.TotalEstimated
	return summary
}

func validate(req ItemRequest) error {
	if strings.TrimSpace(req.ItemName) == "" {
		return fmt.Errorf("%w: item_name is required", ErrInvalidInput)
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}
	if req.EstimatedCost < 0 || req.ActualCost < 0 {
		return fmt.Errorf("%w: costs cannot be negative", ErrInvalidInput)
	}
	return nil
}
