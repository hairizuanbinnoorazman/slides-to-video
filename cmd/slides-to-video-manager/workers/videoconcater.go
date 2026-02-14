package workers

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

func NewVideoConcaterWorker(logger logger.Logger, q queue.Queue, processor videoconcater.VideoConcater) Worker {
	processorFunc := func(ctx context.Context, msg []byte) error {
		job := videoconcater.JobDetails{}
		if err := json.Unmarshal(msg, &job); err != nil {
			return err
		}
		return processor.Process(ctx, job)
	}

	return &QueueWorker{
		Logger:        logger,
		Queue:         q,
		ProcessorFunc: processorFunc,
		WorkerName:    "concatenate-video",
	}
}
