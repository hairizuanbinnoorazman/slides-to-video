package jobs

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

var PDFToImage string = "pdf-split"
var ImageToVideo string = "image-to-video"
var VideoConcat string = "video-concat"

var SuccessStatus string = "success"
var RunningStatus string = "running"
var FailureStatus string = "failed"

type Job struct {
	ID string `json:"id" datastore:"-"`
	// Reference to task entity that would be linked up to this job
	// e.g. If linked to ParentJob - this would reflect on the various jobs that would need to be
	// run for that video project
	RefID string `json:"ref_id"`
	// Dedup id refers to a ID that is unique on per request basis.
	// The reason this ID is available is to provide capability to project the server against duplicated messages
	// From system that does not do guarantee of only one message guarantee: e.g. Google Pubsub
	DedupID      string    `json:"dedup_id"`
	Type         string    `json:"type"`
	Message      string    `json:"message"`
	Status       string    `json:"status"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
}

func NewJob(refID, Type, Msg string) Job {
	jobID, _ := uuid.NewV4()
	return Job{
		ID:           jobID.String(),
		RefID:        refID,
		Type:         Type,
		Message:      Msg,
		Status:       "created",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
}

type JobStore interface {
	CreateJob(ctx context.Context, e Job) error
	GetJob(ctx context.Context, ID string) (Job, error)
	GetAllJobs(ctx context.Context, filters ...filter) ([]Job, error)
	UpdateJob(ctx context.Context, ID string, setters ...func(*Job)) (Job, error)
	DeleteJob(ctx context.Context, ID string) error
	DeleteJobs(ctx context.Context, filters ...filter) error
}

type filter struct {
	Key      string
	Operator string
	Value    string
}

func SetJobStatus(status string) func(*Job) {
	return func(a *Job) {
		a.Status = status
	}
}

func SetDedup(dedupID string) func(*Job) {
	return func(a *Job) {
		a.DedupID = dedupID
	}
}

func FilterRefID(refID string) filter {
	return filter{
		Key:      "RefID",
		Operator: "=",
		Value:    refID,
	}
}

func FilterStatus(status string) filter {
	return filter{
		Key:      "Status",
		Operator: "=",
		Value:    status,
	}
}
