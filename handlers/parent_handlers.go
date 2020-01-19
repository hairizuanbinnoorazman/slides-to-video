package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
)

type CreateProject struct {
	Logger           logger.Logger
	Blobstorage      blobstorage.BlobStorage
	PDFToImageQueue  queue.Queue
	BucketFolderName string
	ProjectStore     project.ProjectStore
	JobStore         jobs.JobStore
}

func (h CreateProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Create Parent Item Handler")
	defer h.Logger.Info("End Create Parent Item Handler")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve parse multipart form data. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
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

	item, err := h.createProjectRecord()
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save parent job to datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	h.Blobstorage.Save(context.Background(), h.BucketFolderName+"/"+item.PDFFile, b.Bytes())

	err = h.createPDFSplitJob(item.ID, item.PDFFile)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create the pdf split job. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v successfully uploaded. Go to the /jobs page for viewing progress", handler.Filename)))
	return
}

func (h CreateProject) createProjectRecord() (project.Project, error) {
	item := project.NewProject()
	fileID, _ := uuid.NewV4()
	item.PDFFile = fileID.String()
	err := h.ProjectStore.CreateProject(context.Background(), item)
	if err != nil {
		return project.Project{}, err
	}
	return item, nil
}

func (h CreateProject) createPDFSplitJob(parentJobID, filename string) error {
	job := jobs.NewJob(parentJobID, "PDF_Split_Job", "")
	values := map[string]string{"id": job.ID, "pdfFileName": filename}
	jsonValue, _ := json.Marshal(values)
	job.Message = string(jsonValue)

	err := h.JobStore.CreateJob(context.Background(), job)
	if err != nil {
		return err
	}

	err = h.PDFToImageQueue.Add(context.Background(), jsonValue)
	if err != nil {
		return err
	}
	return nil
}

type GetProject struct {
	Logger       logger.Logger
	ProjectStore project.ProjectStore
}

func (h GetProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Get Parent Job API Handler")
	defer h.Logger.Info("End Get Parent Job API Handler")

	projectID := mux.Vars(r)["project_id"]
	project, err := h.ProjectStore.GetProject(context.Background(), projectID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	rawProject, _ := json.Marshal(project)

	w.WriteHeader(http.StatusOK)
	w.Write(rawProject)
	return
}

type GetAllProjects struct {
	Logger       logger.Logger
	ProjectStore project.ProjectStore
	ParentStore  jobs.ParentJobStore
}

func (h GetAllProjects) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start View All Parent Jobs API Handler")
	defer h.Logger.Info("End View All Parent Jobs API Handler")

	projects, err := h.ProjectStore.GetAllProjects(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	rawParentJobs, err := json.Marshal(projects)
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
