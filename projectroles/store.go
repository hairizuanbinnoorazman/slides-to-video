package projectroles

import "context"

type Store interface {
	Create(ctx context.Context, e ProjectRole) error
	Get(ctx context.Context, ProjectID, EntityID string) (ProjectRole, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]ProjectRole, error)
	Update(ctx context.Context, ProjectID, EntityID string, setters ...func(*ProjectRole) error) (ProjectRole, error)
	Delete(ctx context.Context, ProjectID, EntityID string) error
}
