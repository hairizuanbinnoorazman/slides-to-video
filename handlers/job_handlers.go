package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type CreateParentJob struct {
	Logger           logger.Logger
	Blobstorage      blobstorage.BlobStorage
	ParentStore      jobs.ParentJobStore
	PDFToImageStore  jobs.PDFToImageStore
	PDFToImageQueue  queue.Queue
	BucketFolderName string
}

func (h CreateParentJob) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Example Handler")
	defer h.Logger.Info("End Example Handler")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve parse multipart form data. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	script := r.FormValue("script")
	h.Logger.Info(script)
	file, handler, err := r.FormFile("myfile")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve form data. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	defer file.Close()
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	io.Copy(bw, file)

	parentJob, err := h.createParentJob(h.ParentStore, handler.Filename, script)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save parent job to datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	h.Blobstorage.Save(context.Background(), h.BucketFolderName+"/"+parentJob.Filename, b.Bytes())

	job, err := h.createPDFSplitJob(h.PDFToImageStore, parentJob.ID, parentJob.Filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save pdf split job. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	values := map[string]string{"id": job.ID, "pdfFileName": job.Filename}
	jsonValue, _ := json.Marshal(values)
	err = h.PDFToImageQueue.Add(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v successfully uploaded. Go to the /jobs page for viewing progress", handler.Filename)))
	return
}

func (h CreateParentJob) createParentJob(store jobs.ParentJobStore, filename, script string) (jobs.ParentJob, error) {
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	parentJob := jobs.ParentJob{
		ID:               jobID,
		OriginalFilename: filename,
		Filename:         jobID + ".pdf",
		Script:           script,
		Status:           "created",
	}
	err := store.StoreParentJob(context.Background(), parentJob)
	if err != nil {
		return jobs.ParentJob{}, err
	}
	return parentJob, nil
}

func (h CreateParentJob) createPDFSplitJob(store jobs.PDFToImageStore, parentJobID, filename string) (jobs.PDFToImageJob, error) {
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	pdfToImageJob := jobs.PDFToImageJob{
		ID:          jobID,
		ParentJobID: parentJobID,
		Filename:    filename,
		Status:      "created",
	}
	err := store.StorePDFToImageJob(context.Background(), pdfToImageJob)
	if err != nil {
		return jobs.PDFToImageJob{}, err
	}
	return pdfToImageJob, nil
}

type ViewAllParentJobsAPI struct {
	Logger      logger.Logger
	ParentStore jobs.ParentJobStore
}

func (h ViewAllParentJobsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start View All Parent Jobs API Handler")
	defer h.Logger.Info("End View All Parent Jobs API Handler")

	parentJobs, err := h.ParentStore.GetAllParentJobs(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	type jobsResponse struct {
		Jobs []jobs.ParentJob `json:"jobs"`
	}

	rawParentJobs, err := json.Marshal(jobsResponse{Jobs: parentJobs})
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawParentJobs)
	return
}
