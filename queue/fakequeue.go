package queue

import (
	"context"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Fake struct {
	Logger logger.Logger
}

func (f Fake) Add(ctx context.Context, message []byte) error {
	f.Logger.Infof("Received following message: %v", string(message))
	return nil
}

func NewFake(l logger.Logger) Fake {
	return Fake{Logger: l}
}
