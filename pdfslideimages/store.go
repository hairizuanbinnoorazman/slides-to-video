package pdfslideimages

import (
	"context"
	"fmt"
)

type Store interface {
	Create(ctx context.Context, e PDFSlideImages) error
	Get(ctx context.Context, projectID, ID string) (PDFSlideImages, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]PDFSlideImages, error)
	Update(ctx context.Context, projectID, ID string, setters ...func(*PDFSlideImages) error) (PDFSlideImages, error)
	Delete(ctx context.Context, projectID, ID string) error
}

func GetUpdaters(runningIdemKey, completeRecIdemKey, state string, assets []SlideAsset) ([]func(*PDFSlideImages) error, error) {
	var s status
	switch state {
	case "running":
		s = running
	case "completed":
		s = completed
	case "error":
		s = errorStatus
	default:
		return []func(*PDFSlideImages) error{}, fmt.Errorf("Bad status is passed into it")
	}
	var setters []func(*PDFSlideImages) error
	if s == running && runningIdemKey == "" {
		return setters, fmt.Errorf("No IdemKey passed to change the status to running state")
	}
	if s == errorStatus && completeRecIdemKey == "" || s == completed && completeRecIdemKey == "" {
		return setters, fmt.Errorf("No CompleteRec IdemKey passed to change status to error/completed")
	}
	if s == completed && len(assets) == 0 {
		return setters, fmt.Errorf("Attempt to set complete but no assets found for pdf slides")
	}
	if s == running && runningIdemKey != "" {
		setters = append(setters, setStatus(s), clearSetRunningIdemKey(runningIdemKey))
		return setters, nil
	}
	if s == errorStatus && completeRecIdemKey != "" {
		setters = append(setters, setStatus(s), clearCompleteRecIdemKey(completeRecIdemKey))
		return setters, nil
	}
	if s == completed && completeRecIdemKey != "" && len(assets) > 0 {
		setters = append(setters, setStatus(s), clearCompleteRecIdemKey(completeRecIdemKey))
		return setters, nil
	}
	return setters, fmt.Errorf("Unexpected issue found")
}

func setStatus(s status) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		a.Status = s
		return nil
	}
}

func setSlideAssets(assets []SlideAsset) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		a.SlideAssets = assets
		return nil
	}
}

func clearSetRunningIdemKey(idemKey string) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if a.SetRunningIdemKey == idemKey {
			a.SetRunningIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func clearCompleteRecIdemKey(idemKey string) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if a.CompleteRecIdemKey == idemKey {
			a.CompleteRecIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}
