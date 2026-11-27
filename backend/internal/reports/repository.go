package reports

import "context"

type Repository interface {
	Create(ctx context.Context, report *Report) error
	GetByID(ctx context.Context, projectID, reportID string) (*Report, error)
	ListByProject(ctx context.Context, projectID string) ([]*Report, error)
	Update(ctx context.Context, report *Report) error
	Delete(ctx context.Context, projectID, reportID string) error
}
