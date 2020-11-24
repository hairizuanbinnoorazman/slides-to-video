package videosegment

import (
	"time"

	"github.com/gofrs/uuid"
)

type status string

var (
	created     status = "created"
	running     status = "running"
	errorStatus status = "error"
	completed   status = "completed"
	unset       status = "unset"
)

type VideoSegment struct {
	ID                 string    `json:"id" datastore:"-" gorm:"type:varchar(40);primary_key"`
	ProjectID          string    `json:"project_id" datastore:"-" gorm:"type:varchar(40)"`
	VideoFile          string    `json:"video_file" gorm:"type:varchar(100)"`
	DateCreated        time.Time `json:"date_created"`
	DateModified       time.Time `json:"date_modified"`
	Order              int       `json:"order" gorm:"type:int"`
	Hidden             bool      `json:"hidden" gorm:"type:bool"`
	Status             status    `json:"status" gorm:"type:varchar(20)"`
	SetRunningIdemKey  string    `json:"-" gorm:"type:varchar(40)"`
	CompleteRecIdemKey string    `json:"-" gorm:"type:varchar(40)"`
	// Image Source
	ImageID string `json:"image_id" gorm:"type:varchar(40)"`
	Script  string `json:"script" gorm:"type:text"`
	// Audio Source
	AudioID string `json:"audio_id" gorm:"type:varchar(40)"`
	// Video Source
	VideoSrcID string `json:"video_src_id" gorm:"type:varchar(40)"`
}

func (v *VideoSegment) IsReady() bool {
	if v.Status == completed {
		return true
	}
	return false
}

func New(projectID, imageID string, order int) VideoSegment {
	videoSegmentID, _ := uuid.NewV4()
	return VideoSegment{
		ID:           videoSegmentID.String(),
		ProjectID:    projectID,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Status:       created,
		ImageID:      imageID,
		Order:        order,
	}
}

type ByOrder []VideoSegment

func (s ByOrder) Len() int { return len(s) }

func (s ByOrder) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByOrder) Less(i, j int) bool { return s[i].Order < s[j].Order }
