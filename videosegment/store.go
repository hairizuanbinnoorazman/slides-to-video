package videosegment

import "context"

type Store interface {
	Create(ctx context.Context, e VideoSegment) error
	Get(ctx context.Context, projectID, ID string) (VideoSegment, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]VideoSegment, error)
	Update(ctx context.Context, projectID, ID string, setters ...func(*VideoSegment)) (VideoSegment, error)
	Delete(ctx context.Context, projectID, ID string) error
}

func SetStatus(status string) func(*VideoSegment) {
	return func(a *VideoSegment) {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		} else if status == "error" {
			a.Status = errored
		}
	}
}

func SetHidden(hide bool) func(*VideoSegment) {
	return func(a *VideoSegment) {
		a.Hidden = hide
	}
}

func SetVideoFile(videoFile string) func(*VideoSegment) {
	return func(a *VideoSegment) {
		a.VideoFile = videoFile
	}
}

func SetIdemKey(idemKey string) func(*VideoSegment) {
	return func(a *VideoSegment) {
		a.IdemKey = idemKey
	}
}
