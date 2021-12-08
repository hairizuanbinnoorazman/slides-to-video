package acl

import "context"

type Store interface {
	Create(ctx context.Context, e ACL) error
	Get(ctx context.Context, ProjectID, UserID string) (ACL, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]ACL, error)
	Update(ctx context.Context, ProjectID, UserID string, setters ...func(*ACL) error) (ACL, error)
	Delete(ctx context.Context, ProjectID, UserID string) error
}
