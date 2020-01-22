package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type UpdateJobStatus struct {
	Logger           logger.Logger
	JobStore         jobs.JobStore
	ProjectStore     project.ProjectStore
	VideoConcatQueue queue.Queue
}

func (h UpdateJobStatus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Patch Job API Handler")
	defer h.Logger.Info("End Patch Job API Handler")

	jobID := mux.Vars(r)["job_id"]
	type reqBody struct {
		JobType    string          `json:"job_type"`
		Status     string          `json:"status"`
		JobDetails json.RawMessage `json:"job_details"`
	}

	raw, err := ioutil.ReadAll(r.Body)
	var item reqBody
	err = json.Unmarshal(raw, &item)
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
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.PDFToImage, jobID, item.Status)
		err = h.handlePDFToImage(jobID, item.Status, item.JobDetails)
	case jobs.ImageToVideo:
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.ImageToVideo, jobID, item.Status)
		err = h.handleImageToVideo(jobID, item.Status, item.JobDetails)
	case jobs.VideoConcat:
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.VideoConcat, jobID, item.Status)
		err = h.handleVideoConcat(jobID, item.Status, item.JobDetails)
	}

	if err != nil {
		errMsg := fmt.Sprintf("Error - Issue with handling of job. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

type SlideDetail struct {
	ImageID string `json:"image"`
	SlideNo int    `json:"slide_no"`
}

func (h UpdateJobStatus) handlePDFToImage(jobID, jobStatus string, rawJobDetails []byte) error {
	switch jobStatus {
	case jobs.SuccessStatus:
		h.Logger.Info("Handling Successful PDF To Image")
		job, err := h.JobStore.GetJob(context.Background(), jobID)
		if err != nil {
			return fmt.Errorf("Unable to retrieve the job")
		}

		err = h.JobStore.UpdateJob(context.Background(), jobID, jobs.SetJobStatus(jobs.SuccessStatus))
		if err != nil {
			return fmt.Errorf("Unable to updata successful pdf to image")
		}

		type succcessfulJobDetails struct {
			SlideDetails []SlideDetail `json:"slide_details"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		var setters []func(*project.Project)
		for _, slideDetail := range jobDetails.SlideDetails {
			setters = append(setters, project.SetImage(slideDetail.ImageID, slideDetail.SlideNo))
		}
		err = h.ProjectStore.UpdateProject(context.Background(), job.RefID, setters...)
		if err != nil {
			return fmt.Errorf("Unable to update project successfully")
		}

	default:
		h.Logger.Info("Unknown status set for the job")
		return fmt.Errorf("Unknown status set for job")
	}
	return nil
}

func (h UpdateJobStatus) handleImageToVideo(jobID, jobStatus string, rawJobDetails []byte) error {
	switch jobStatus {
	case jobs.SuccessStatus:
		h.Logger.Info("Successful Image to Video")

		job, err := h.JobStore.GetJob(context.Background(), jobID)
		if err != nil {
			return fmt.Errorf("Unable to retrieve job information")
		}

		type succcessfulJobDetails struct {
			ID         string `json:"id"`
			OutputFile string `json:"output_file"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetVideoID(jobDetails.ID, jobDetails.OutputFile))

		successfulJobs, _ := h.JobStore.GetAllJobs(context.Background(), jobs.FilterRefID(job.RefID), jobs.FilterStatus(jobs.SuccessStatus))
		project, _ := h.ProjectStore.GetProject(context.Background(), job.RefID)

		if len(project.SlideAssets) == len(successfulJobs) {
			videoConcatJob := jobs.NewJob(job.RefID, jobs.VideoConcat, "")
			var videoList []string
			for _, slideAsset := range project.SlideAssets {
				videoList = append(videoList, slideAsset.VideoID)
			}
			videoConcatJobDetails := map[string]interface{}{"id": videoConcatJob.ID, "video_ids": videoList}
			rawVideoConcatJobDetails, _ := json.Marshal(videoConcatJobDetails)
			videoConcatJob.Message = string(rawVideoConcatJobDetails)
			h.JobStore.CreateJob(context.Background(), videoConcatJob)
			h.VideoConcatQueue.Add(context.Background(), rawVideoConcatJobDetails)
		}

	case jobs.FailureStatus:
		h.Logger.Info("Failed Image to Video")
	case jobs.RunningStatus:
		h.Logger.Info("Running Image to Video")
	}
	return nil
}

func (h UpdateJobStatus) handleVideoConcat(jobID, jobStatus string, rawJobDetails []byte) error {
	switch jobStatus {
	case jobs.SuccessStatus:
		h.Logger.Info("Successful Video Concat")

		job, err := h.JobStore.GetJob(context.Background(), jobID)
		if err != nil {
			return fmt.Errorf("Unable to retrieve job information")
		}

		type succcessfulJobDetails struct {
			OutputVideo string `json:"output_video"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		err = h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetVideoOutputID(jobDetails.OutputVideo), project.SetStatus(project.SuccessfulStatus))
		if err != nil {
			return fmt.Errorf("Unable to update project")
		}

		err = h.JobStore.DeleteJobs(context.Background(), jobs.FilterRefID(job.RefID))
		if err != nil {
			return fmt.Errorf("Issue with removing data")
		}

	case jobs.FailureStatus:
		h.Logger.Info("Failed Video Concat")
	case jobs.RunningStatus:
		h.Logger.Info("Running Video Concat")
	}
	return nil
}
