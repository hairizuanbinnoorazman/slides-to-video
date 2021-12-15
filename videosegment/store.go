package videosegment

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
)

type Store interface {
	Create(ctx context.Context, e VideoSegment) error
	Get(ctx context.Context, projectID, ID string) (VideoSegment, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]VideoSegment, error)
	Update(ctx context.Context, projectID, ID string, setters ...func(*VideoSegment) error) (VideoSegment, error)
	Delete(ctx context.Context, projectID, ID string) error
}

func GetUpdaters(runningIdemKey, completeRecIdemKey, state, videoFile, script string, hidden *bool) ([]func(*VideoSegment) error, error) {
	var s status
	switch state {
	case "running":
		s = running
	case "completed":
		s = completed
	case "error":
		s = errorStatus
	default:
		s = unset
	}
	var setters []func(*VideoSegment) error
	if s == unset && script != "" {
		setters = append(setters, setScript(script))
		return setters, nil
	}
	if s == running && runningIdemKey == "" {
		return setters, fmt.Errorf("No IdemKey passed to change the status to running state")
	}
	if s == errorStatus && completeRecIdemKey == "" || s == completed && completeRecIdemKey == "" {
		return setters, fmt.Errorf("No CompleteRec IdemKey passed to change status to error/completed")
	}
	if s == completed && videoFile == "" || s == completed && !strings.Contains(videoFile, ".mp4") {
		return setters, fmt.Errorf("Missing/invalid videofile")
	}
	if s == running && runningIdemKey != "" {
		setters = append(setters, setStatus(s), clearSetRunningIdemKey(runningIdemKey))
		return setters, nil
	}
	if s == errorStatus && completeRecIdemKey != "" {
		setters = append(setters, setStatus(s), clearCompleteRecIdemKey(completeRecIdemKey))
		return setters, nil
	}
	if s == completed && completeRecIdemKey != "" {
		setters = append(setters, setStatus(s), clearCompleteRecIdemKey(completeRecIdemKey), setVideoFile(videoFile))
		return setters, nil
	}
	if hidden != nil {
		setters = append(setters, setHidden(*hidden))
	}
	return setters, fmt.Errorf("Unexpected issue found")
}

func RegenerateIdemKeys() ([]func(*VideoSegment) error, error) {
	var setters []func(*VideoSegment) error
	setters = append(setters, recreateIdemKeys())
	return setters, nil
}

func ResetStatus() ([]func(*VideoSegment) error, error) {
	var setters []func(*VideoSegment) error
	setters = append(setters, setStatus(unset))
	setters = append(setters, setVideoFile(""))
	return setters, nil
}

func recreateIdemKeys() func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		idemKey1, _ := uuid.NewV4()
		idemKey2, _ := uuid.NewV4()
		a.SetRunningIdemKey = idemKey1.String()
		a.CompleteRecIdemKey = idemKey2.String()
		return nil
	}
}

func setStatus(s status) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.Status = s
		return nil
	}
}

func setHidden(hide bool) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.Hidden = hide
		return nil
	}
}

func setVideoFile(videoFile string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.VideoFile = videoFile
		return nil
	}
}

func setScript(script string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		a.Script = script
		return nil
	}
}

func clearSetRunningIdemKey(idemKey string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		if a.SetRunningIdemKey == idemKey {
			a.SetRunningIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func clearCompleteRecIdemKey(idemKey string) func(*VideoSegment) error {
	return func(a *VideoSegment) error {
		if a.CompleteRecIdemKey == idemKey {
			a.CompleteRecIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}
