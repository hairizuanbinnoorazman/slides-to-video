package job

import "context"

type Store interface {
	Create(ctx context.Context, e Job) error
	Get(ctx context.Context, ID string) (Job, error)
	GetAll(ctx context.Context, Limit, After int) ([]Job, error)
	Delete(ctx context.Context, ID string) error
}
