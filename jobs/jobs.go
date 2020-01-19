package jobs

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

var PDFToImage string = "pdf-split"
var ImageToVideo string = "image-to-video"
var VideoConcat string = "video-concat"

type ParentJob struct {
	ID               string    `json:"id" datastore:"-"`
	OriginalFilename string    `json:"original_filename"`
	Filename         string    `json:"filename"`
	Script           string    `json:"script"`
	Status           string    `json:"status"`
	VideoFile        string    `json:"video_file"`
	DateCreated      time.Time `json:"date_created"`
	DateModified     time.Time `json:"date_modified"`
	UserID           string    `json:"user_id"`
}

type PDFToImageJob struct {
	ID           string    `json:"id" datastore:"-"`
	ParentJobID  string    `json:"parent_job_id"`
	Filename     string    `json:"filename"`
	Status       string    `json:"status"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
	UserID       string    `json:"user_id"`
}

type ImageToVideoJob struct {
	ID           string    `json:"id" datastore:"-"`
	ParentJobID  string    `json:"parent_job_id"`
	ImageID      string    `json:"image_id"`
	SlideID      int       `json:"slide_id"`
	Text         string    `json:"text"`
	Status       string    `json:"status"`
	OutputFile   string    `json:"output_file"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
	UserID       string    `json:"user_id"`
}

type VideoConcatJob struct {
	ID           string    `json:"id" datastore:"-"`
	ParentJobID  string    `json:"parent_job_id"`
	Videos       []string  `json:"videos"`
	Status       string    `json:"status"`
	OutputFile   string    `json:"output_file"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
	UserID       string    `json:"user_id"`
}

type Job struct {
	ID string `json:"id" datastore:"-"`
	// Reference to task entity that would be linked up to this job
	// e.g. If linked to ParentJob - this would reflect on the various jobs that would need to be
	// run for that video project
	RefID        string    `json:"ref_id"`
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

type ParentJobStore interface {
	StoreParentJob(ctx context.Context, e ParentJob) error
	GetParentJob(ctx context.Context, ID string) (ParentJob, error)
	GetAllParentJobs(ctx context.Context) ([]ParentJob, error)
}

type PDFToImageStore interface {
	StorePDFToImageJob(ctx context.Context, e PDFToImageJob) error
	GetPDFToImageJob(ctx context.Context, ID string) (PDFToImageJob, error)
	GetAllPDFToImageJobs(ctx context.Context) ([]PDFToImageJob, error)
}

type ImageToVideoStore interface {
	StoreImageToVideoJob(ctx context.Context, e ImageToVideoJob) error
	GetImageToVideoJob(ctx context.Context, ID string) (ImageToVideoJob, error)
	GetAllImageToVideoJobs(ctx context.Context, filterByParentID string) ([]ImageToVideoJob, error)
}

type VideoConcatStore interface {
	StoreVideoConcatJob(ctx context.Context, e VideoConcatJob) error
	GetVideoConcatJob(ctx context.Context, ID string) (VideoConcatJob, error)
	GetAllVideoConcatJobs(ctx context.Context) ([]VideoConcatJob, error)
}

type JobStore interface {
	CreateJob(ctx context.Context, e Job) error
	GetJob(ctx context.Context, ID string) (Job, error)
	GetAllJobs(ctx context.Context) ([]Job, error)
	UpdateJob(ctx context.Context, ID string, setters ...func(*Job)) error
}

func SetJobStatus(status string) func(*Job) {
	return func(a *Job) {
		a.Status = status
	}
}
