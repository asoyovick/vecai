package materials

import "context"

type Repository interface {
	Create(ctx context.Context, material *Material) error
	GetByID(ctx context.Context, projectID, materialID string) (*Material, error)
	ListByProject(ctx context.Context, projectID string) ([]*Material, error)
	Update(ctx context.Context, material *Material) error
	Delete(ctx context.Context, projectID, materialID string) error
}
