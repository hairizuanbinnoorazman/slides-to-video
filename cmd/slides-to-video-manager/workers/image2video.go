package workers

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/image-to-video/image2videoconverter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

func NewImage2VideoWorker(logger logger.Logger, q queue.Queue, processor image2videoconverter.Image2VideoConverter) Worker {
	processorFunc := func(msg []byte) error {
		job := image2videoconverter.JobDetails{}
		if err := json.Unmarshal(msg, &job); err != nil {
			return err
		}
		return processor.Process(context.Background(), job)
	}

	return &QueueWorker{
		Logger:        logger,
		Queue:         q,
		ProcessorFunc: processorFunc,
		WorkerName:    "image-to-video",
	}
}
