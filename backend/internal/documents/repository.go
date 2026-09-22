package documents

import "context"

type Repository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, projectID, documentID string) (*Document, error)
	ListByProject(ctx context.Context, projectID string) ([]*Document, error)
	Delete(ctx context.Context, projectID, documentID string) error
}
