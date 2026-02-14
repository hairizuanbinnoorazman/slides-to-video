package workers

import (
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/pdfsplitter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

func NewPDFSplitterWorker(logger logger.Logger, q queue.Queue, processor pdfsplitter.PDFSplitter) Worker {
	processorFunc := func(msg []byte) error {
		job := pdfsplitter.PdfSplitJob{}
		if err := json.Unmarshal(msg, &job); err != nil {
			return err
		}
		return processor.Process(job)
	}

	return &QueueWorker{
		Logger:        logger,
		Queue:         q,
		ProcessorFunc: processorFunc,
		WorkerName:    "pdf-splitter",
	}
}
