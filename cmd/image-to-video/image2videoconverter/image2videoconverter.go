package image2videoconverter

type JobDetails struct {
	ID                 string `json:"id" validate:"required"`
	ProjectID          string `json:"project_id" validate:"required"`
	ImageID            string `json:"image_id" validate:"required"`
	Text               string `json:"script" validate:"required"`
	RunningIdemKey     string `json:"idem_key_running" validate:"required"`
	CompleteRecIdemKey string `json:"idem_key_complete_rec" validate:"required"`
}

type Image2VideoConverter interface {
	Process() error
}

type TextToSpeechEngine interface {
	Generate(text string) ([]byte, error)
}
