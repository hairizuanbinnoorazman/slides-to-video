package videoconcater

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type basic struct {
	queue        queue.Queue
	projectStore project.Store
}

func NewBasic(q queue.Queue, s project.Store) basic {
	return basic{
		queue:        q,
		projectStore: s,
	}
}

func (b basic) Start(ctx context.Context, projectID string, videoSegmentList []string) error {
	if len(videoSegmentList) == 0 {
		return fmt.Errorf("No video segments to combine to single output video")
	}

	updaters, _ := project.RegenerateIdemKeys()
	newProject, err := b.projectStore.Update(ctx, projectID, "", updaters...)
	if err != nil {
		return err
	}

	values := map[string]interface{}{
		"project_id":            projectID,
		"video_segments":        videoSegmentList,
		"idem_key_running":      newProject.SetRunningIdemKey,
		"idem_key_complete_rec": newProject.CompleteRecIdemKey,
	}
	jsonValue, _ := json.Marshal(values)

	err = b.queue.Add(ctx, jsonValue)
	if err != nil {
		return err
	}

	return nil
}
