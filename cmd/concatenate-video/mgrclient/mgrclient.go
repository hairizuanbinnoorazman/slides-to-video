package mgrclient

import "context"

type Client interface {
	UpdateRunning(ctx context.Context, projectID, idemKey string) error
	FailedTask(ctx context.Context, projectID, idemKey string) error
	CompleteTask(ctx context.Context, projectID, idemKey, videoOutputID string) error
}
