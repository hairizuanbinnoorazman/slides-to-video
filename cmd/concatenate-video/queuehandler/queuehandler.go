package queuehandler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type basic struct {
	queue         queue.Queue
	logger        logger.Logger
	videoConcater videoconcater.VideoConcater
}

func NewBasic(logger logger.Logger, queue queue.Queue, concater videoconcater.VideoConcater) basic {
	return basic{
		logger:        logger,
		queue:         queue,
		videoConcater: concater,
	}
}

func (h basic) HandleMessages() {
	h.logger.Infof("Queue Handler started")
	for {
		msg, err := h.queue.Pop(context.TODO())
		if err != nil {
			h.logger.Errorf("Unable to receive message for queue system. Err: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		h.logger.Infof("Received the following message. Msg: %v", string(msg))

		job := videoconcater.JobDetails{}
		err = json.Unmarshal(msg, &job)

		if err != nil {
			h.logger.Errorf("Unable to marshal message for queue system. Err: %v", err)
			continue
		}

		err = h.videoConcater.Process(context.TODO(), job)
		if err != nil {
			h.logger.Errorf("Error in processing job. Err: %v", err)
			continue
		}
	}

}
