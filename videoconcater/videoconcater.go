package videoconcater

import "context"

type VideoConcater interface {
	Start(ctx context.Context, projectID, userID string, videoSegmentList []string) error
}
