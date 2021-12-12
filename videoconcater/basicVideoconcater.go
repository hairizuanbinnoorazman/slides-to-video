package videoconcater

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
)

type basic struct {
	queue        queue.Queue
	projectStore project.Store
	authStore    services.Auth
}

func NewBasic(q queue.Queue, s project.Store, a services.Auth) basic {
	return basic{
		queue:        q,
		projectStore: s,
		authStore:    a,
	}
}

func (b basic) Start(ctx context.Context, projectID string, userID string, authToken string, videoSegmentList []string) error {
	if len(videoSegmentList) == 0 {
		return fmt.Errorf("No video segments to combine to single output video")
	}

	updaters, _ := project.RegenerateIdemKeys()
	newProject, err := b.projectStore.Update(ctx, projectID, userID, updaters...)
	if err != nil {
		return err
	}

	token, err := services.NewToken(userID, b.authStore.ExpiryTime, b.authStore.Secret, b.authStore.Issuer)
	if err != nil {
		return err
	}

	values := map[string]interface{}{
		"id":                    projectID,
		"auth_token":            "Bearer " + token,
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
