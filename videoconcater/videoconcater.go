package videoconcater

import "context"

type VideoConcater interface {
	Start(ctx context.Context, projectID, userID, authToken string, videoSegmentList []string) error
}
