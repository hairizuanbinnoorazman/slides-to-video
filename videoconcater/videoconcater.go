package videoconcater

import "context"

type VideoConcater interface {
	Start(ctx context.Context, projectID string, videoSegmentList []string) error
}
