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

func SetStatus(status string) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		} else if status == "error" {
			a.Status = errorStatus
		}
		return nil
	}
}

func SetSlideAssets(assets []SlideAsset) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if len(assets) == 0 {
			return fmt.Errorf("Empty slide assets passed in")
		}
		a.SlideAssets = assets
		return nil
	}
}

func ClearSetRunningIdemKey(idemKey string) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if a.SetRunningIdemKey == idemKey {
			a.SetRunningIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}

func ClearCompleteRecIdemKey(idemKey string) func(*PDFSlideImages) error {
	return func(a *PDFSlideImages) error {
		if a.CompleteRecIdemKey == idemKey {
			a.CompleteRecIdemKey = ""
			return nil
		}
		return fmt.Errorf("Idemkey set is not the same. Cannot clear idemkey values")
	}
}
