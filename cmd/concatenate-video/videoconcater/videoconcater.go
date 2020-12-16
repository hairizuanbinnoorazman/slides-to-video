package videoconcater

import "context"

type JobDetails struct {
	ID                 string   `json:"id" validate:"required"`
	VideoIDs           []string `json:"video_segments" validate:"required"`
	RunningIdemKey     string   `json:"idem_key_running" validate:"required"`
	CompleteRecIdemKey string   `json:"idem_key_complete_rec" validate:"required"`
}

type VideoConcater interface {
	Process(ctx context.Context, job JobDetails) error
}
