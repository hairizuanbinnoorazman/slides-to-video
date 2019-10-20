package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gofrs/uuid"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

type scriptParse struct {
	Script []string `json:"script"`
}

type exampleHandler struct {
	logger           Logger
	client           *storage.Client
	datastoreClient  *datastore.Client
	pubsubClient     *pubsub.Client
	bucketName       string
	bucketFolderName string
	parentTableName  string
	tableName        string
	topicName        string
}

func (h exampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Example Handler")
	defer h.logger.Info("End Example Handler")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve parse multipart form data. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	script := r.FormValue("script")
	h.logger.Info(script)
	file, handler, err := r.FormFile("myfile")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve form data. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	defer file.Close()
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	io.Copy(bw, file)

	bs := BlobStorage{
		logger:     h.logger,
		client:     h.client,
		bucketName: h.bucketName,
	}
	parentJob, err := h.CreateParentJob(handler.Filename, script)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save parent job to datastore. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	bs.Save(context.Background(), h.bucketFolderName+"/"+parentJob.Filename, b.Bytes())

	job, err := h.CreatePDFSplitJob(parentJob.ID, parentJob.Filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	values := map[string]string{"id": job.ID, "pdfFileName": job.Filename}
	jsonValue, _ := json.Marshal(values)
	pubsub := Pubsub{h.logger, h.pubsubClient, h.topicName}
	err = pubsub.publish(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v successfully uploaded", handler.Filename)))
	return
}

func (h exampleHandler) CreateParentJob(filename, script string) (ParentJob, error) {
	store := NewStore(h.datastoreClient, h.parentTableName)
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	parentJob := ParentJob{
		ID:               jobID,
		OriginalFilename: filename,
		Filename:         jobID + ".pdf",
		Script:           script,
		Status:           "created",
	}
	err := store.StoreParentJob(context.Background(), parentJob)
	if err != nil {
		return ParentJob{}, err
	}
	return parentJob, nil
}

func (h exampleHandler) CreatePDFSplitJob(parentJobID, filename string) (PDFToImageJob, error) {
	store := NewStore(h.datastoreClient, h.tableName)
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	pdfToImageJob := PDFToImageJob{
		ID:          jobID,
		ParentJobID: parentJobID,
		Filename:    filename,
		Status:      "created",
	}
	err := store.StorePDFToImageJob(context.Background(), pdfToImageJob)
	if err != nil {
		return PDFToImageJob{}, err
	}
	return pdfToImageJob, nil
}

type mainPage struct {
	logger Logger
}

func (h mainPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Main Page Handler")
	defer h.logger.Info("End Main Page Handler")

	t, err := template.ParseFiles("main.html")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}

	err = t.Execute(w, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}
	return
}

type reportPDFSplit struct {
	logger          Logger
	datastoreClient *datastore.Client
	pubsubClient    *pubsub.Client
	parentTableName string
	tableName       string
	nextTableName   string
	nextTopicName   string
}

type SlideDetail struct {
	ImageID string `json:"image"`
	SlideNo int    `json:"slide_no"`
}

func (h reportPDFSplit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report PDF Split Handler")
	defer h.logger.Info("End Report PDF Split Handler")

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
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetPDFToImageJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v %v", err, h.tableName)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	job.Status = req.Status
	err = store.StorePDFToImageJob(context.Background(), job)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save and store pdf status. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	parentStore := NewStore(h.datastoreClient, h.parentTableName)
	parentJob, _ := parentStore.GetParentJob(context.Background(), job.ParentJobID)
	var scripts scriptParse
	json.Unmarshal([]byte(parentJob.Script), &scripts)

	for i, slideDetail := range req.SlideDetails {
		store := NewStore(h.datastoreClient, h.nextTableName)
		rawID, _ := uuid.NewV4()
		jobID := rawID.String()
		image2videoJob := ImageToVideoJob{
			ID:          jobID,
			ParentJobID: job.ParentJobID,
			ImageID:     slideDetail.ImageID,
			Text:        scripts.Script[i],
			Status:      "created",
		}
		store.StoreImageToVideoJob(context.Background(), image2videoJob)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}

type reportImageToVideo struct {
	logger          Logger
	datastoreClient *datastore.Client
	pubsubClient    *pubsub.Client
	tableName       string
	nextTableName   string
	nextTopicName   string
}

func (h reportImageToVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report Image To Video Handler")
	defer h.logger.Info("End Report Image to Video Handler")

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
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetImageToVideoJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image to video details. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputFile
	store.StoreImageToVideoJob(context.Background(), job)

	items, err := store.GetAllImageToVideoJobs(context.Background(), job.ParentJobID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve list of job ids based on parent. Error: %v", err)
		h.logger.Error(errMsg)
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
			h.logger.Errorf("OUTPUT FILE: %v", item.OutputFile)
		}
	}
	if len(items) != completedJobs {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Not all jobs completed, will not issue new one"))
		return
	}

	nextStore := NewStore(h.datastoreClient, h.nextTableName)
	rawID, _ := uuid.NewV4()
	concatJob := VideoConcatJob{
		ID:          rawID.String(),
		ParentJobID: job.ParentJobID,
		Videos:      videoList,
		Status:      "created",
	}
	nextStore.StoreVideoConcatJob(context.Background(), concatJob)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stored New value"))
}

type reportVideoConcat struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
	parentTableName string
}

func (h reportVideoConcat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report Video Concat Handler")
	defer h.logger.Info("End Report Video Concat Handler")

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
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetVideoConcatJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	parentStore := NewStore(h.datastoreClient, h.parentTableName)
	parentJob, _ := parentStore.GetParentJob(context.Background(), job.ParentJobID)

	job.Status = req.Status
	job.OutputFile = req.OutputVideo

	store.StoreVideoConcatJob(context.Background(), job)

	parentJob.Status = req.Status
	parentJob.VideoFile = req.OutputVideo

	parentStore.StoreParentJob(context.Background(), parentJob)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}
