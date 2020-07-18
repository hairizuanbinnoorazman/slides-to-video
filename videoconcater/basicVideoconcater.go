package videoconcater

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type basic struct {
	queue queue.Queue
}

func NewBasic(q queue.Queue) basic {
	return basic{
		queue: q,
	}
}

func (b basic) Start(ctx context.Context, projectID string, videoSegmentList []string) error {
	values := map[string]interface{}{
		"project_id":     projectID,
		"video_segments": videoSegmentList,
	}
	jsonValue, _ := json.Marshal(values)

	err := b.queue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}

	return nil
}
