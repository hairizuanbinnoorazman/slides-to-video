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
	ID                 string    `json:"id" datastore:"-"`
	ProjectID          string    `json:"project_id" datastore:"-"`
	VideoFile          string    `json:"video_file"`
	DateCreated        time.Time `json:"date_created"`
	DateModified       time.Time `json:"date_modified"`
	Order              int       `json:"order"`
	Hidden             bool      `json:"hidden"`
	Status             status    `json:"status"`
	SetRunningIdemKey  string    `json:"-"`
	CompleteRecIdemKey string    `json:"-"`
	// Image Source
	ImageID string `json:"image_id"`
	Script  string `json:"script"`
	// Audio Source
	AudioID string `json:"audio_id"`
	// Video Source
	VideoID string `json:"video_id"`
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
