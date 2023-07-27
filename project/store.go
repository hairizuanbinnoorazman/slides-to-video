package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
)

type Store interface {
	Create(ctx context.Context, e Project) error
	Get(ctx context.Context, ID string) (Project, error)
	GetAll(ctx context.Context, UserID string, Limit, After int) ([]Project, error)
	Count(ctx context.Context, UserID string) (int, error)
	Update(ctx context.Context, ID string, setters ...func(*Project) error) (Project, error)
	Delete(ctx context.Context, ID string) error
}

func GetUpdaters(name, runningIdemKey, completeRecIdemKey, state, videoOutputID string) ([]func(*Project) error, error) {
	var s status
	switch state {
	case "running":
		s = running
	case "completed":
		s = completed
	case "error":
		s = errorStatus
	}
	var setters []func(*Project) error
	if name != "" {
		setters = append(setters, setName(name))
	}
	if s == running && runningIdemKey == "" {
		return setters, fmt.Errorf("no IdemKey passed to change the status to running state")
	}
	if s == errorStatus && completeRecIdemKey == "" || s == completed && completeRecIdemKey == "" {
		return setters, fmt.Errorf("no CompleteRec IdemKey passed to change status to error/completed")
	}
	if s == completed {
		if videoOutputID == "" || !strings.Contains(videoOutputID, ".mp4") {
			return setters, fmt.Errorf("empty video output id/invalid video output id")
		}
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
		setters = append(setters, setStatus(s), clearCompleteRecIdemKey(completeRecIdemKey), setVideoOutputID(videoOutputID))
		return setters, nil
	}
	if len(setters) > 0 {
		return setters, nil
	}
	return setters, fmt.Errorf("unexpected issue found")
}

func setName(name string) func(*Project) error {
	return func(a *Project) error {
		a.Name = name
		return nil
	}
}

func setVideoOutputID(videoOutputID string) func(*Project) error {
	return func(a *Project) error {
		a.VideoOutputID = videoOutputID
		return nil
	}
}

func setStatus(s status) func(*Project) error {
	return func(a *Project) error {
		a.Status = s
		return nil
	}
}

func clearSetRunningIdemKey(idemKey string) func(*Project) error {
	return func(a *Project) error {
		if a.SetRunningIdemKey == idemKey {
			a.SetRunningIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func clearCompleteRecIdemKey(idemKey string) func(*Project) error {
	return func(a *Project) error {
		if a.CompleteRecIdemKey == idemKey {
			a.CompleteRecIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func RegenerateIdemKeys() ([]func(*Project) error, error) {
	var setters []func(*Project) error
	setters = append(setters, recreateIdemKeys())
	return setters, nil
}

func recreateIdemKeys() func(*Project) error {
	return func(a *Project) error {
		idemKey1, _ := uuid.NewV4()
		idemKey2, _ := uuid.NewV4()
		a.SetRunningIdemKey = idemKey1.String()
		a.CompleteRecIdemKey = idemKey2.String()
		return nil
	}
}

type ByDateCreated []Project

func (s ByDateCreated) Len() int { return len(s) }

func (s ByDateCreated) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByDateCreated) Less(i, j int) bool { return s[i].DateCreated.Before(s[j].DateCreated) }
