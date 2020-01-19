package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"

	"github.com/gofrs/uuid"

	"cloud.google.com/go/storage"
)

type scriptParse struct {
	Script []string `json:"script"`
}

type reportPDFSplit struct {
	Logger            logger.Logger
	ParentStore       jobs.ParentJobStore
	PdfToImageStore   jobs.PDFToImageStore
	ImageToVideoStore jobs.ImageToVideoStore
}

type SlideDetail struct {
	ImageID string `json:"image"`
	SlideNo int    `json:"slide_no"`
}

func (h reportPDFSplit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Report PDF Split Handler")
	defer h.Logger.Info("End Report PDF Split Handler")

	type request struct {
		ID           string        `json:"id"`
		SlideDetails []SlideDetail `json:"slide_details"`
		Status       string        `json:"status"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}

	job, err := h.PdfToImageStore.GetPDFToImageJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	if job.Status == "completed" {
		h.Logger.Infof("Detected completed PDF job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	err = h.PdfToImageStore.StorePDFToImageJob(context.Background(), job)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save and store pdf status. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	for _, slideDetail := range req.SlideDetails {
		rawID, _ := uuid.NewV4()
		jobID := rawID.String()
		splitFileName := strings.Split(slideDetail.ImageID, "-")
		slideNoAndFileFormat := strings.Split(splitFileName[len(splitFileName)-1], ".")
		num, _ := strconv.Atoi(slideNoAndFileFormat[0])

		image2videoJob := jobs.ImageToVideoJob{
			ID:           jobID,
			ParentJobID:  job.ParentJobID,
			ImageID:      slideDetail.ImageID,
			SlideID:      num,
			Text:         "",
			Status:       "created",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		}
		h.ImageToVideoStore.StoreImageToVideoJob(context.Background(), image2videoJob)

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
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

type reportImageToVideo struct {
	Logger            logger.Logger
	ImageToVideoStore jobs.ImageToVideoStore
	VideoConcatStore  jobs.VideoConcatStore
	VideoConcatQueue  queue.Queue
}

func (h reportImageToVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Report Image To Video Handler")
	defer h.Logger.Info("End Report Image to Video Handler")

	type request struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		OutputFile string `json:"output_file"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	job, err := h.ImageToVideoStore.GetImageToVideoJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image to video details. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	if job.Status == "completed" {
		h.Logger.Infof("Detected completed Image to Video job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputFile
	h.ImageToVideoStore.StoreImageToVideoJob(context.Background(), job)

	items, err := h.ImageToVideoStore.GetAllImageToVideoJobs(context.Background(), job.ParentJobID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve list of job ids based on parent. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	completedJobs := 0
	videoList := []string{}
	for _, item := range items {
		if item.Status == "completed" {
			completedJobs = completedJobs + 1
			videoList = append(videoList, item.OutputFile)
		}
	}
	if len(items) != completedJobs {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Not all jobs completed, will not issue new one"))
		return
	}

	videoList = []string{}
	for i, _ := range items {
		for _, item := range items {
			if item.SlideID == i {
				videoList = append(videoList, item.OutputFile)
			}
		}
	}

	rawID, _ := uuid.NewV4()
	concatJob := jobs.VideoConcatJob{
		ID:          rawID.String(),
		ParentJobID: job.ParentJobID,
		Videos:      videoList,
		Status:      "created",
	}
	h.VideoConcatStore.StoreVideoConcatJob(context.Background(), concatJob)

	values := map[string]interface{}{"id": concatJob.ID, "video_ids": concatJob.Videos}
	jsonValue, _ := json.Marshal(values)
	err = h.VideoConcatQueue.Add(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.Logger.Error(errMsg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stored New value"))
}

type reportVideoConcat struct {
	Logger           logger.Logger
	ParentStore      jobs.ParentJobStore
	VideoConcatStore jobs.VideoConcatStore
}

func (h reportVideoConcat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Report Video Concat Handler")
	defer h.Logger.Info("End Report Video Concat Handler")

	type request struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		OutputVideo string `json:"output_video"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	job, err := h.VideoConcatStore.GetVideoConcatJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	parentJob, _ := h.ParentStore.GetParentJob(context.Background(), job.ParentJobID)

	if parentJob.Status == "completed" {
		h.Logger.Infof("Detected parent job status is already completed but another request comes in to reset it. Reject it")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputVideo

	h.VideoConcatStore.StoreVideoConcatJob(context.Background(), job)

	parentJob.Status = req.Status
	parentJob.VideoFile = req.OutputVideo

	h.ParentStore.StoreParentJob(context.Background(), parentJob)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}

type downloadVideo struct {
	logger        logger.Logger
	storageClient *storage.Client
	bucketName    string
}

func (h downloadVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Download Handler")
	defer h.logger.Info("End Download Handler")

	filename := mux.Vars(r)["video_id"]
	if filename == "" {
		errMsg := fmt.Sprintf("Missing video id field")
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	bs := BlobStorage{h.logger, h.storageClient, h.bucketName}
	content, err := bs.Load(context.Background(), "videos/"+filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to download file from blob storage. Err: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(200)
	w.Write(content)
}

type downloadImage struct {
	logger        logger.Logger
	storageClient *storage.Client
	bucketName    string
}

func (h downloadImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Download Handler")
	defer h.logger.Info("End Download Handler")

	filename := mux.Vars(r)["image_id"]
	if filename == "" {
		errMsg := fmt.Sprintf("Missing image id field")
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	bs := BlobStorage{h.logger, h.storageClient, h.bucketName}
	content, err := bs.Load(context.Background(), "images/"+filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to download file from blob storage. Err: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(200)
	w.Write(content)
}
