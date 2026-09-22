package boq

import "context"

type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, projectID, itemID string) (*Item, error)
	ListByProject(ctx context.Context, projectID string) ([]*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, projectID, itemID string) error
}
