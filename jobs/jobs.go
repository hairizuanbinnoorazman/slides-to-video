package jobs

import "time"

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
