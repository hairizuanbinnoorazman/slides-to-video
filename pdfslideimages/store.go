package pdfslideimages

import "context"

type Store interface {
	Create(ctx context.Context, e PDFSlideImages) error
	Get(ctx context.Context, projectID, ID string) (PDFSlideImages, error)
	GetAll(ctx context.Context, ProjectID string, Limit, After int) ([]PDFSlideImages, error)
	Update(ctx context.Context, projectID, ID string, setters ...func(*PDFSlideImages)) (PDFSlideImages, error)
	Delete(ctx context.Context, projectID, ID string) error
}

func SetStatus(status string) func(*PDFSlideImages) {
	return func(a *PDFSlideImages) {
		if status == "running" {
			a.Status = running
		} else if status == "completed" {
			a.Status = completed
		} else if status == "error" {
			a.Status = errorStatus
		}
	}
}

func SetSlideAssets(assets []SlideAsset) func(*PDFSlideImages) {
	return func(a *PDFSlideImages) {
		if len(assets) == 0 {
			return
		}
		a.SlideAssets = assets
	}
}

func SetIdemKey(idemKey string) func(*PDFSlideImages) {
	return func(a *PDFSlideImages) {
		a.IdemKey = idemKey
	}
}
