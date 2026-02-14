package workers

import (
	"context"
	"errors"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type Worker interface {
	Start(ctx context.Context) error
}

type QueueWorker struct {
	Logger        logger.Logger
	Queue         queue.Queue
	ProcessorFunc func(context.Context, []byte) error
	WorkerName    string
}

func (w *QueueWorker) Start(ctx context.Context) error {
	w.Logger.Infof("Starting worker: %s", w.WorkerName)
	for {
		select {
		case <-ctx.Done():
			w.Logger.Infof("Worker %s shutting down", w.WorkerName)
			return ctx.Err()
		default:
			msg, err := w.Queue.Pop(ctx)
			if err != nil {
				w.Logger.Errorf("[%s] Error popping from queue: %v", w.WorkerName, err)
				// Check if error is due to context cancellation
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
					// Don't sleep on cancellation, allow loop to observe ctx.Done()
					continue
				}
				time.Sleep(10 * time.Second)
				continue
			}

			if err := w.ProcessorFunc(ctx, msg); err != nil {
				w.Logger.Errorf("[%s] Error processing message: %v", w.WorkerName, err)
			}
		}
	}
}
