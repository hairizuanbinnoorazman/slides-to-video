// Package mgrclient contains functionality that wraps capability to contact manager node
// This is to update status and pass info back to be persisted
package mgrclient

import "context"

type Client interface {
	UpdateRunning(ctx context.Context, projectID, videoSegmentID, idemKey string) error
	FailedTask(ctx context.Context, projectID, videoSegmentID, idemKey string) error
	CompleteTask(ctx context.Context, projectID, videoSegmentID, idemKey, videoFile string) error
}
