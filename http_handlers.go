package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"
)

type scriptParse struct {
	Script []string `json:"script"`
}

type startVideoGeneration struct {
	Logger            logger.Logger
	ParentStore       jobs.ParentJobStore
	PdfToImageStore   jobs.PDFToImageStore
	ImageToVideoStore jobs.ImageToVideoStore
	ImageToVideoQueue queue.Queue
}

func (h startVideoGeneration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parentJobID := mux.Vars(r)["parent_job_id"]

	parentJob, _ := h.ParentStore.GetParentJob(context.Background(), parentJobID)
	parentJob.Status = "running"
	h.ParentStore.StoreParentJob(context.Background(), parentJob)

	var scripts scriptParse
	rawData, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(rawData, &scripts)

	imageToVideoJobs, _ := h.ImageToVideoStore.GetAllImageToVideoJobs(context.Background(), parentJobID)

	for i, imageToVideoJob := range imageToVideoJobs {
		imageToVideoJob.Text = scripts.Script[i]
		imageToVideoJob.DateModified = time.Now()
		h.ImageToVideoStore.StoreImageToVideoJob(context.Background(), imageToVideoJob)

		values := map[string]string{"id": imageToVideoJob.ID, "image_id": imageToVideoJob.ImageID, "text": imageToVideoJob.Text}
		jsonValue, _ := json.Marshal(values)
		err := h.ImageToVideoQueue.Add(context.Background(), jsonValue)
		if err != nil {
			errMsg := fmt.Sprintf("Error - unable to send image to video job. Error: %v", err)
			h.Logger.Error(errMsg)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}
