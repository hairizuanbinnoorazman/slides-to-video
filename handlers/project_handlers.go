package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/job"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videogenerator"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type CreateProject struct {
	Logger       logger.Logger
	ProjectStore project.Store
	ACLStore     acl.Store
}

func (h CreateProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Create Parent Item Handler")
	defer h.Logger.Info("End Create Parent Item Handler")

	ctx := r.Context()
	userID := ctx.Value(userIDKey).(string)

	item := project.New()
	err := h.ProjectStore.Create(context.Background(), item)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create project in datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	acl := acl.New(item.ID, userID)
	err = h.ACLStore.Create(context.Background(), acl)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create acl control in datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	rawItem, _ := json.Marshal(item)
	w.Write(rawItem)
}

type UpdateProject struct {
	Logger       logger.Logger
	ProjectStore project.Store
}

func (h UpdateProject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Update Project API Handler")
	defer h.Logger.Info("End Update Project API Handler")

	ctx := r.Context()
	userID := ctx.Value(userIDKey).(string)

	projectID := mux.Vars(r)["project_id"]
	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type updateProjectReq struct {
		Status             string `json:"status"`
		VideoOutputID      string `json:"video_output_id"`
		SetRunningIdemKey  string `json:"idem_key_running"`
		CompleteRecIdemKey string `json:"idem_key_complete_rec"`
		Name               string `json:"name"`
	}

	req := updateProjectReq{}
	err = json.Unmarshal(rawReq, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to marshal value out to project item. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	updaters, err := project.GetUpdaters(req.Name, req.SetRunningIdemKey, req.CompleteRecIdemKey, req.Status, req.VideoOutputID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create the required updaters to update project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	updatedProject, err := h.ProjectStore.Update(context.Background(), projectID, userID, updaters...)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to update project item. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
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

	ctx := r.Context()
	userID := ctx.Value(userIDKey).(string)

	projectID := mux.Vars(r)["project_id"]
	project, err := h.ProjectStore.Get(context.Background(), projectID, userID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	rawProject, _ := json.Marshal(project)

	w.WriteHeader(http.StatusOK)
	w.Write(rawProject)
}

type GetAllProjects struct {
	Logger       logger.Logger
	ProjectStore project.Store
}

func (h GetAllProjects) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start View All Parent Jobs API Handler")
	defer h.Logger.Info("End View All Parent Jobs API Handler")

	ctx := r.Context()
	userID := ctx.Value(userIDKey).(string)

	rawOffset := r.URL.Query().Get("offset")
	offset := 0
	if rawOffset != "" {
		offset, _ = strconv.Atoi(rawOffset)
	}

	rawLimit := r.URL.Query().Get("limit")
	limit := 10
	if rawLimit != "" {
		limit, _ = strconv.Atoi(rawLimit)
	}

	type getAllProjectsResp struct {
		Projects []project.Project `json:"projects"`
		Offset   int               `json:"offset"`
		Total    int               `json:"total"`
		Limit    int               `json:"limit"`
	}

	projects, err := h.ProjectStore.GetAll(ctx, userID, limit, offset)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	projectResp := getAllProjectsResp{
		Projects: projects,
		Offset:   offset,
		Limit:    limit,
		Total:    1000,
	}

	rawProjectResp, err := json.Marshal(projectResp)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawProjectResp)
}

type StartVideoConcat struct {
	Logger        logger.Logger
	ProjectStore  project.Store
	VideoConcater videoconcater.VideoConcater
}

func (h StartVideoConcat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start StartVideoConcat API Handler")
	defer h.Logger.Info("End StartVideoConcat API Handler")

	ctx := r.Context()
	projectID := mux.Vars(r)["project_id"]
	userID := ctx.Value(userIDKey).(string)

	project, err := h.ProjectStore.Get(ctx, projectID, userID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve the project entity. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	videoSegmentIDs, err := project.GetVideoSegmentList()
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve the project entity. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	err = h.VideoConcater.Start(context.Background(), projectID, userID, videoSegmentIDs)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to start async video generation. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	resp := map[string]string{
		"status": "successfully sent",
	}
	rawResp, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}

type StartProjectGenerateVideo struct {
	Logger             logger.Logger
	JobStore           job.Store
	ProjectStore       project.Store
	VideoSegmentsStore videosegment.Store
	VideoGenerator     videogenerator.VideoGenerator
}

func (h StartProjectGenerateVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start StartProjectGenerateVideo API Handler")
	defer h.Logger.Info("End StartProjectGenerateVideo API Handler")

	ctx := r.Context()
	projectID := mux.Vars(r)["project_id"]
	userID := ctx.Value(userIDKey).(string)

	singleProject, err := h.ProjectStore.Get(context.TODO(), projectID, userID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve project details. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	var updateVideoSegmentErr error
	for _, v := range singleProject.VideoSegments {
		updateVideoSegmentErr = nil
		updaters, _ := videosegment.ResetStatus()
		_, updateVideoSegmentErr = h.VideoSegmentsStore.Update(context.TODO(), projectID, v.ID, updaters...)
		if updateVideoSegmentErr != nil {
			errMsg := fmt.Sprintf("Error - unable to update video segment. ProjectID: %v :: VideoSegmentID: %v :: Error: %v", err)
			h.Logger.Error(errMsg)
		}
	}
	if updateVideoSegmentErr != nil {
		errMsg := fmt.Sprintf("Error - unable to reset video segment error. Error: %v", updateVideoSegmentErr)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	singleProject, err = h.ProjectStore.Get(context.TODO(), projectID, userID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve project details. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	var generateVideoErr error
	for _, v := range singleProject.VideoSegments {
		generateVideoErr = h.VideoGenerator.Start(context.TODO(), v)
		if generateVideoErr != nil {
			errMsg := fmt.Sprintf("Error - unable to generate video segment. ProjectID: %v :: VideoSegmentID: %v :: Error: %v", err)
			h.Logger.Error(errMsg)
		}
	}
	if generateVideoErr != nil {
		errMsg := fmt.Sprintf("Error - unable to reset video segment error. Error: %v", generateVideoErr)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	newJob := job.New(projectID, userID)
	err = h.JobStore.Create(context.TODO(), newJob)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to set job to generate video. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	resp := map[string]string{
		"status": "successfully set job",
	}
	rawResp, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}
