package videosegment

import (
	"context"
	"fmt"
)

type Store interface {
	Create(ctx context.Context, e VideoSegment) error
	Get(ctx context.Context, projectID, ID string) (VideoSegment, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]VideoSegment, error)
	Update(ctx context.Context, projectID, ID string, setters ...func(*VideoSegment) error) (VideoSegment, error)
	Delete(ctx context.Context, projectID, ID string) error
}

func SetStatus(status string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		} else if status == "error" {
			a.Status = errored
		}
		return nil
	}
}

func SetHidden(hide bool) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.Hidden = hide
		return nil
	}
}

func SetVideoFile(videoFile string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.VideoFile = videoFile
		return nil
	}
}

func ClearSetRunningIdemKey(idemKey string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		if a.SetRunningIdemKey == idemKey {
			a.SetRunningIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func ClearCompleteRecIdemKey(idemKey string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		if a.CompleteRecIdemKey == idemKey {
			a.CompleteRecIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}
