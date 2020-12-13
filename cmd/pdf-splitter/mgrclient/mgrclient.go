package mgrclient

import "context"

type SlideAsset struct {
	ImageID string `json:"image_id"`
	Order   int    `json:"order"`
}

type Client interface {
	UpdateRunning(ctx context.Context, projectID, pdfslideimagesID, idemKey string) error
	FailedTask(ctx context.Context, projectID, pdfslideimagesID, idemKey string) error
	CompleteTask(ctx context.Context, projectID, pdfslideimagesID, idemKey string, slideAssets []SlideAsset) error
}
