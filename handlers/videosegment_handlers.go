package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videogenerator"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type CreateVideoSegment struct {
	Logger            logger.Logger
	VideoSegmentStore videosegment.Store
}

func (h CreateVideoSegment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start CreateVideoSegment Handler")
	defer h.Logger.Info("End CreateVideoSegment Handler")

	projectID := mux.Vars(r)["project_id"]
	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(errMsg))
		return
	}

	type createVideoSegmentReq struct {
		ImageID string `json:"image_id"`
		Order   int    `json:"order"`
	}
	req := createVideoSegmentReq{}
	json.Unmarshal(rawReq, &req)

	item := videosegment.New(projectID, req.ImageID, req.Order)
	err = h.VideoSegmentStore.Create(context.Background(), item)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create video segment in datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusCreated)
	rawItem, _ := json.Marshal(item)
	w.Write(rawItem)
	return
}

type UpdateVideoSegment struct {
	Logger            logger.Logger
	VideoSegmentStore videosegment.Store
}

func (h UpdateVideoSegment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start CreateVideoSegment Handler")
	defer h.Logger.Info("End CreateVideoSegment Handler")

	projectID := mux.Vars(r)["project_id"]
	videoSegmentID := mux.Vars(r)["videosegment_id"]
	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(errMsg))
		return
	}

	type updateVideoSegmentReq struct {
		VideoFile string `json:"video_file"`
		Hidden    bool   `json:"hidden"`
		Status    string `json:"status"`
		IdemKey   string `json:"idem_key"`
	}
	req := updateVideoSegmentReq{}
	json.Unmarshal(rawReq, &req)

	var updaters []func(*videosegment.VideoSegment)
	updaters = append(updaters, videosegment.SetHidden(req.Hidden))
	updaters = append(updaters, videosegment.SetIdemKey(req.IdemKey))
	if req.VideoFile != "" {
		updaters = append(updaters, videosegment.SetVideoFile(req.VideoFile))
	}
	if req.Status != "" {
		updaters = append(updaters, videosegment.SetStatus(req.Status))
	}

	item, err := h.VideoSegmentStore.Update(context.Background(), projectID, videoSegmentID, updaters...)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create video segment in datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	rawItem, _ := json.Marshal(item)
	w.Write(rawItem)
	return
}

type GetVideoSegment struct {
	Logger            logger.Logger
	VideoSegmentStore videosegment.Store
}

func (h GetVideoSegment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start CreateVideoSegment Handler")
	defer h.Logger.Info("End CreateVideoSegment Handler")

	projectID := mux.Vars(r)["project_id"]
	videoSegmentID := mux.Vars(r)["videosegment_id"]

	videosegment, err := h.VideoSegmentStore.Get(context.Background(), projectID, videoSegmentID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get video segment in datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	rawItem, _ := json.Marshal(videosegment)
	w.Write(rawItem)
	return
}

type StartVideoSegmentGeneration struct {
	Logger            logger.Logger
	VideoSegmentStore videosegment.Store
	VideoGenerator    videogenerator.VideoGenerator
}

func (h StartVideoSegmentGeneration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start VideoSegmentGeneration API Handler")
	defer h.Logger.Info("End VideoSegmentGeneration API Handler")

	projectID := mux.Vars(r)["project_id"]
	videosegmentID := mux.Vars(r)["project_id"]

	videosegment, err := h.VideoSegmentStore.Get(context.Background(), projectID, videosegmentID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve the project entity. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	err = h.VideoGenerator.Start(context.Background(), videosegment)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to start async video generation. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	resp := map[string]string{
		"status": "successfully sent",
	}
	rawResp, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}
