package project

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
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

type Project struct {
	ID                 string                          `json:"id" datastore:"-" gorm:"type:varchar(40);primary_key"`
	Name               string                          `json:"name" gorm:"type:varchar(250)"`
	DateCreated        time.Time                       `json:"date_created"`
	DateModified       time.Time                       `json:"date_modified"`
	Status             status                          `json:"status" gorm:"type:varchar(40)"`
	VideoSegments      []videosegment.VideoSegment     `json:"video_segments,omitempty" datastore:"-"`
	PDFSlideImages     []pdfslideimages.PDFSlideImages `json:"pdf_slide_images,omitempty" datastore:"-"`
	VideoOutputID      string                          `json:"video_output_id,omitempty" gorm:"type:varchar(40)"`
	ACLs               []acl.ACL                       `json:"acls" datastore:"-" gorm:"-"`
	SetRunningIdemKey  string                          `json:"-" gorm:"varchar(40)"`
	CompleteRecIdemKey string                          `json:"-" gorm:"varchar(40)"`
}

func New() Project {
	projectID, _ := uuid.NewV4()
	currentTime := time.Now()
	return Project{
		ID:           projectID.String(),
		Name:         "default",
		DateCreated:  currentTime,
		DateModified: currentTime,
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
