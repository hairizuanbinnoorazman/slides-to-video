package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type UpdateJobStatus struct {
	Logger   logger.Logger
	JobStore jobs.JobStore
}

func (h UpdateJobStatus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Patch Job API Handler")
	defer h.Logger.Info("End Patch Job API Handler")

	jobID := mux.Vars(r)["job_id"]
	type reqBody struct {
		JobType string `json:"job_type"`
		Status  string `json:"status"`
	}

	raw, err := ioutil.ReadAll(r.Body)
	var item reqBody
	err = json.Unmarshal(raw, item)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse request body correctly. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	updateJobStatus := jobs.SetJobStatus(item.Status)
	h.JobStore.UpdateJob(context.Background(), jobID, updateJobStatus)

	switch item.JobType {
	case jobs.PDFToImage:
		h.Logger.Info(jobs.PDFToImage)
	case jobs.ImageToVideo:
		h.Logger.Info(jobs.ImageToVideo)
	case jobs.VideoConcat:
		h.Logger.Info(jobs.VideoConcat)
	}

	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
