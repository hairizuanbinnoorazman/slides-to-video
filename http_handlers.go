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
	"strconv"
	"strings"

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
	w.Write([]byte(fmt.Sprintf("%v successfully uploaded. Go to the /jobs page for viewing progress", handler.Filename)))
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

	if job.Status == "completed" {
		h.logger.Infof("Detected completed PDF job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
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

	parentJob.Status = "running"
	parentStore.StoreParentJob(context.Background(), parentJob)

	for i, slideDetail := range req.SlideDetails {
		store := NewStore(h.datastoreClient, h.nextTableName)
		rawID, _ := uuid.NewV4()
		jobID := rawID.String()
		splitFileName := strings.Split(slideDetail.ImageID, "-")
		slideNoAndFileFormat := strings.Split(splitFileName[len(splitFileName)-1], ".")
		num, _ := strconv.Atoi(slideNoAndFileFormat[0])

		image2videoJob := ImageToVideoJob{
			ID:          jobID,
			ParentJobID: job.ParentJobID,
			ImageID:     slideDetail.ImageID,
			SlideID:     num,
			Text:        scripts.Script[i],
			Status:      "created",
		}
		store.StoreImageToVideoJob(context.Background(), image2videoJob)

		values := map[string]string{"id": image2videoJob.ID, "image_id": image2videoJob.ImageID, "text": image2videoJob.Text}
		jsonValue, _ := json.Marshal(values)
		pubsub := Pubsub{h.logger, h.pubsubClient, h.nextTopicName}
		err = pubsub.publish(context.Background(), jsonValue)
		if err != nil {
			errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
			h.logger.Error(errMsg)
		}
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

	if job.Status == "completed" {
		h.logger.Infof("Detected completed Image to Video job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
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

	nextStore := NewStore(h.datastoreClient, h.nextTableName)
	rawID, _ := uuid.NewV4()
	concatJob := VideoConcatJob{
		ID:          rawID.String(),
		ParentJobID: job.ParentJobID,
		Videos:      videoList,
		Status:      "created",
	}
	nextStore.StoreVideoConcatJob(context.Background(), concatJob)

	values := map[string]interface{}{"id": concatJob.ID, "video_ids": concatJob.Videos}
	jsonValue, _ := json.Marshal(values)
	pubsub := Pubsub{h.logger, h.pubsubClient, h.nextTopicName}
	err = pubsub.publish(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
	}

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

	if parentJob.Status == "completed" {
		h.logger.Infof("Detected parent job status is already completed but another request comes in to reset it. Reject it")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputVideo

	store.StoreVideoConcatJob(context.Background(), job)

	parentJob.Status = req.Status
	parentJob.VideoFile = req.OutputVideo

	parentStore.StoreParentJob(context.Background(), parentJob)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}

type viewAllParentJobs struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
}

func (h viewAllParentJobs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start View All Parent Jobs Handler")
	defer h.logger.Info("End View All Parent Jobs Handler")

	store := NewStore(h.datastoreClient, h.tableName)
	parentJobs, err := store.GetAllParentJobs(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	t, err := template.ParseFiles("parentJobs.html")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	varmap := map[string]interface{}{
		"parentJobs": parentJobs,
	}

	err = t.Execute(w, varmap)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}
	return
}

type viewAllParentJobsAPI struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
}

func (h viewAllParentJobsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start View All Parent Jobs API Handler")
	defer h.logger.Info("End View All Parent Jobs API Handler")

	store := NewStore(h.datastoreClient, h.tableName)
	parentJobs, err := store.GetAllParentJobs(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	rawParentJobs, err := json.Marshal(parentJobs)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawParentJobs)
	return
}

type downloadJob struct {
	logger        Logger
	storageClient *storage.Client
	bucketName    string
}

func (h downloadJob) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Download Handler")
	defer h.logger.Info("End Download Handler")

	filename, ok := r.URL.Query()["filename"]
	if !ok {
		errMsg := fmt.Sprintf("Error - Missing filename from url param.")
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	bs := BlobStorage{h.logger, h.storageClient, h.bucketName}
	content, err := bs.Load(context.Background(), "videos/"+filename[0])
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
