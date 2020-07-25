package project

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type status string
type permissions string

var (
	created     status = "created"
	running     status = "running"
	completed   status = "completed"
	errorStatus status = "error"

	owner  status = "owner"
	editor status = "editor"
	reader status = "reader"
)

type ACL struct {
	UserID      string
	Permissions string
}

type Project struct {
	ID                 string                          `json:"id" datastore:"-"`
	DateCreated        time.Time                       `json:"date_created"`
	DateModified       time.Time                       `json:"date_modified"`
	Status             status                          `json:"status"`
	VideoSegments      []videosegment.VideoSegment     `json:"video_segments,omitempty" datastore:"-"`
	PDFSlideImages     []pdfslideimages.PDFSlideImages `json:"pdf_slide_images,omitempty" datastore:"-"`
	VideoOutputID      string                          `json:"video_output_id,omitempty"`
	ACLs               []ACL                           `json:"acls" datastore:"-"`
	SetRunningIdemKey  string                          `json:"-"`
	CompleteRecIdemKey string                          `json:"-"`
}

func New() Project {
	projectID, _ := uuid.NewV4()
	return Project{
		ID:           projectID.String(),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Status:       created,
	}
}

func (p *Project) GetVideoSegmentList() ([]string, error) {
	items := []string{}
	for _, v := range p.VideoSegments {
		if v.VideoFile == "" {
			return []string{}, fmt.Errorf("unable to concatenate due to missing video file record")
		}
		items = append(items, v.VideoFile)
	}
	return items, nil
}
