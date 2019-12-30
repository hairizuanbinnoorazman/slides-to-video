package jobs

import (
	"context"
	"time"
)

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
