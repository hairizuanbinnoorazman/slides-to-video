package queue

import "context"

type Queue interface {
	Add(ctx context.Context, message []byte) error
	Pop(ctx context.Context) ([]byte, error)
}
