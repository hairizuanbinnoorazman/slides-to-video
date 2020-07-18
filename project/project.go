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
	created   status = "created"
	running   status = "running"
	completed status = "completed"

	owner  status = "owner"
	editor status = "editor"
	reader status = "reader"
)

type ACL struct {
	UserID      string
	Permissions string
}

type Project struct {
	ID             string                          `json:"id" datastore:"-"`
	DateCreated    time.Time                       `json:"date_created"`
	DateModified   time.Time                       `json:"date_modified"`
	Status         status                          `json:"status"`
	VideoSegments  []videosegment.VideoSegment     `json:"video_segments,omitempty" datastore:"-"`
	PDFSlideImages []pdfslideimages.PDFSlideImages `json:"pdf_slide_images,omitempty" datastore:"-"`
	VideoOutputID  string                          `json:"video_output_id,omitempty"`
	ACLs           []ACL                           `json:"acls" datastore:"-"`
	IdemKey        string                          `json:"idem_key"`
}

// ValidateForConcat is for checking if project item contains the necessary information to
// do concatenation
func (p *Project) ValidateForConcat() error {
	if len(p.VideoSegments) == 0 {
		return fmt.Errorf("no video segments to concatenate together")
	}
	for _, v := range p.VideoSegments {
		if !v.IsReady() {
			return fmt.Errorf("video segments are being processed, or may have errors")
		}
	}
	return nil
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
