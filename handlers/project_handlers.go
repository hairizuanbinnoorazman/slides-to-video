package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type CreateProject struct {
	Logger       logger.Logger
	ProjectStore project.Store
}

func (h CreateProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Create Parent Item Handler")
	defer h.Logger.Info("End Create Parent Item Handler")

	item := project.New()
	err := h.ProjectStore.Create(context.Background(), item)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create project in datastore. Error: %v", err)
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

type UpdateProject struct {
	Logger       logger.Logger
	ProjectStore project.Store
}

func (h UpdateProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Update Project API Handler")
	defer h.Logger.Info("End Update Project API Handler")

	projectID := mux.Vars(r)["project_id"]
	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	type updateProjectReq struct {
		Status             string `json:"status"`
		VideoOutputID      string `json:"video_output_id"`
		SetRunningIdemKey  string `json:"idem_key_running"`
		CompleteRecIdemKey string `json:"idem_key_complete_rec"`
	}

	req := updateProjectReq{}
	err = json.Unmarshal(rawReq, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to marshal value out to project item. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	updaters, err := project.GetUpdaters(req.SetRunningIdemKey, req.CompleteRecIdemKey, req.Status, req.VideoOutputID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create the required updaters to update project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	updatedProject, err := h.ProjectStore.Update(context.Background(), projectID, "user-id", updaters...)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to update project item. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	rawUpdateProject, _ := json.Marshal(updatedProject)
	w.WriteHeader(200)
	w.Write(rawUpdateProject)
}

type GetProject struct {
	Logger       logger.Logger
	ProjectStore project.Store
}

func (h GetProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Get Project API Handler")
	defer h.Logger.Info("End Get Project API Handler")

	projectID := mux.Vars(r)["project_id"]
	project, err := h.ProjectStore.Get(context.Background(), projectID, "user-id")
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
	ProjectStore project.Store
}

func (h GetAllProjects) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start View All Parent Jobs API Handler")
	defer h.Logger.Info("End View All Parent Jobs API Handler")

	projects, err := h.ProjectStore.GetAll(context.Background(), "user-id", 100, 0)
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

type StartVideoConcat struct {
	Logger            logger.Logger
	VideoSegmentStore videosegment.Store
	ProjectStore      project.Store
	VideoConcater     videoconcater.VideoConcater
}

func (h StartVideoConcat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start StartVideoConcat API Handler")
	defer h.Logger.Info("End StartVideoConcat API Handler")

	ctx := r.Context()
	projectID := mux.Vars(r)["project_id"]

	project, err := h.ProjectStore.Get(ctx, projectID, "")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve the project entity. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(errMsg))
		return
	}

	videoSegmentIDs, err := project.GetVideoSegmentList()
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve the project entity. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	err = h.VideoConcater.Start(context.Background(), projectID, videoSegmentIDs)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to start async video generation. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
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
