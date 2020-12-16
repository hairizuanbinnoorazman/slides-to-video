package queuehandler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/image-to-video/image2videoconverter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type basic struct {
	queue                queue.Queue
	logger               logger.Logger
	image2videoConverter image2videoconverter.Image2VideoConverter
}

func NewBasic(logger logger.Logger, queue queue.Queue, converter image2videoconverter.Image2VideoConverter) basic {
	return basic{
		logger:               logger,
		queue:                queue,
		image2videoConverter: converter,
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

		job := image2videoconverter.JobDetails{}
		err = json.Unmarshal(msg, &job)

		if err != nil {
			h.logger.Errorf("Unable to marshal message for queue system. Err: %v", err)
			continue
		}

		err = h.image2videoConverter.Process(context.TODO(), job)
		if err != nil {
			h.logger.Errorf("Error in processing job. Err: %v", err)
			continue
		}
	}

}
