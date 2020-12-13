package queuehandler

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/pdfsplitter"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type basic struct {
	queue       queue.Queue
	logger      logger.Logger
	pdfsplitter pdfsplitter.PDFSplitter
}

func NewBasic(logger logger.Logger, queue queue.Queue, pdfsplitter pdfsplitter.PDFSplitter) basic {
	return basic{
		logger:      logger,
		queue:       queue,
		pdfsplitter: pdfsplitter,
	}
}

func (h basic) HandleMessages() {
	for {
		msg, err := h.queue.Pop(context.TODO())
		if err != nil {
			h.logger.Errorf("Unable to receive message for queue system. Err: %v", err)
			continue
		}

		h.logger.Infof("Received the following message. Msg: %v", msg)

		job := pdfsplitter.PdfSplitJob{}
		err = json.Unmarshal(msg, &job)

		if err != nil {
			h.logger.Errorf("Unable to marshal message for queue system. Err: %v", err)
			continue
		}

		err = h.pdfsplitter.Process(job)
		if err != nil {
			h.logger.Errorf("Error in processing job. Err: %v", err)
			continue
		}
	}

}
