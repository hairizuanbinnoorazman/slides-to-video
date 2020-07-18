package pdfslideimages

import (
	"time"

	"github.com/gofrs/uuid"
)

// For job statuses
type status string

var (
	created     status = "created"
	running     status = "running"
	errorStatus status = "error"
	completed   status = "completed"
)

type SlideAsset struct {
	ImageID string `json:"image_id"`
	Order   int    `json:"order"`
}

type PDFSlideImages struct {
	ID                 string       `json:"id" datastore:"-"`
	ProjectID          string       `json:"project_id" datastore:"-"`
	PDFFile            string       `json:"pdf_file"`
	DateCreated        time.Time    `json:"date_created"`
	SlideAssets        []SlideAsset `json:"slide_assets"`
	Status             status       `json:"status"`
	SetRunningIdemKey  string       `json:"-"`
	CompleteRecIdemKey string       `json:"-"`
}

func (p *PDFSlideImages) IsComplete() bool {
	if p.Status == completed {
		return true
	}
	return false
}

func New(projectID string) PDFSlideImages {
	id, _ := uuid.NewV4()
	idemKey1, _ := uuid.NewV4()
	idemKey2, _ := uuid.NewV4()
	return PDFSlideImages{
		ID:                 id.String(),
		ProjectID:          projectID,
		PDFFile:            id.String() + ".pdf",
		DateCreated:        time.Now(),
		Status:             created,
		SetRunningIdemKey:  idemKey1.String(),
		CompleteRecIdemKey: idemKey2.String(),
	}
}
