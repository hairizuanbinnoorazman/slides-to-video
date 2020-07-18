package videogenerator

import (
	"context"
	"encoding/json"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type basic struct {
	queue queue.Queue
}

func NewBasic(q queue.Queue) basic {
	return basic{
		queue: q,
	}
}

func (b basic) Start(ctx context.Context, v videosegment.VideoSegment) error {
	values := map[string]string{
		"id":       v.ID,
		"idem_key": v.IdemKey,
		"script":   v.Script,
		"image_id": v.ImageID,
	}
	jsonValue, _ := json.Marshal(values)

	err := b.queue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}

	return nil
}
