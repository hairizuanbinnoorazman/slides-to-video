package mgrclient

import "context"

type Client interface {
	UpdateRunning(ctx context.Context, authToken, projectID, idemKey string) error
	FailedTask(ctx context.Context, authToken, projectID, idemKey string) error
	CompleteTask(ctx context.Context, authToken, projectID, idemKey, videoOutputID string) error
}
