package workers

import (
	"context"
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
	ProcessorFunc func([]byte) error
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
				time.Sleep(10 * time.Second)
				continue
			}

			if err := w.ProcessorFunc(msg); err != nil {
				w.Logger.Errorf("[%s] Error processing message: %v", w.WorkerName, err)
			}
		}
	}
}
