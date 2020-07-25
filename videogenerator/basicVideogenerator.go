package videogenerator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type basic struct {
	queue             queue.Queue
	videosegmentStore videosegment.Store
}

func NewBasic(q queue.Queue, store videosegment.Store) basic {
	return basic{
		queue:             q,
		videosegmentStore: store,
	}
}

func (b basic) Start(ctx context.Context, v videosegment.VideoSegment) error {
	updaters, _ := videosegment.RegenerateIdemKeys()
	newV, err := b.videosegmentStore.Update(ctx, v.ProjectID, v.ID, updaters...)
	if err != nil {
		return fmt.Errorf("unable to generate idem keys for video segment creation. %v %v", v.ProjectID, v.ID)
	}

	values := map[string]string{
		"id":                    newV.ID,
		"project_id":            newV.ProjectID,
		"script":                newV.Script,
		"image_id":              newV.ImageID,
		"idem_key_running":      newV.SetRunningIdemKey,
		"idem_key_complete_rec": newV.CompleteRecIdemKey,
	}
	jsonValue, _ := json.Marshal(values)

	err = b.queue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}

	return nil
}
