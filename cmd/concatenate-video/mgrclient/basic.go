package mgrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type basic struct {
	client       *http.Client
	baseEndpoint string
	logger       logger.Logger
}

func NewBasic(l logger.Logger, endpoint string, cl *http.Client) basic {
	return basic{
		baseEndpoint: endpoint,
		logger:       l,
		client:       cl,
	}
}

func (b basic) UpdateRunning(ctx context.Context, projectID, idemKey string) error {
	endpoint := b.baseEndpoint + "/project/" + projectID
	type updateInput struct {
		Status            string `json:"status"`
		IdemKeySetRunning string `json:"idem_key_running"`
	}
	updateInputReq := updateInput{
		Status:            "running",
		IdemKeySetRunning: idemKey,
	}
	rawUpdateInputReq, err := json.Marshal(updateInputReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(rawUpdateInputReq))
	if err != nil {
		return err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	rawResp, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("issue with updating. %v", string(rawResp))
	}

	return nil
}

func (b basic) FailedTask(ctx context.Context, projectID, idemKey string) error {
	endpoint := b.baseEndpoint + "/project/" + projectID
	type updateInput struct {
		Status             string `json:"status"`
		IdemKeyCompleteRec string `json:"idem_key_complete_rec"`
	}
	updateInputReq := updateInput{
		Status:             "error",
		IdemKeyCompleteRec: idemKey,
	}
	rawUpdateInputReq, err := json.Marshal(updateInputReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(rawUpdateInputReq))
	if err != nil {
		return err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	rawResp, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("issue with updating. %v", string(rawResp))
	}

	return nil
}

func (b basic) CompleteTask(ctx context.Context, projectID, idemKey, videoOutputID string) error {
	endpoint := b.baseEndpoint + "/project/" + projectID
	type updateInput struct {
		Status             string `json:"status"`
		VideoOutputID      string `json:"video_output_id"`
		IdemKeyCompleteRec string `json:"idem_key_complete_rec"`
	}
	updateInputReq := updateInput{
		Status:             "completed",
		VideoOutputID:      videoOutputID,
		IdemKeyCompleteRec: idemKey,
	}
	rawUpdateInputReq, err := json.Marshal(updateInputReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(rawUpdateInputReq))
	if err != nil {
		return err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	rawResp, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("issue with updating. %v", string(rawResp))
	}

	return nil
}
