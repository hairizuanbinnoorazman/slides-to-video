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
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
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
		DedupID    string          `json:"dedup_id"`
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

	// Run the dedup message check - drop duplicated messages
	checkJob, err := h.JobStore.GetJob(context.Background(), jobID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to retrive job to check deduplication id. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}
	if checkJob.DedupID == item.DedupID {
		w.WriteHeader(200)
		return
	}

	updateJobStatus := jobs.SetJobStatus(item.Status)
	updateDedupID := jobs.SetDedup(item.DedupID)
	job, err := h.JobStore.UpdateJob(context.Background(), jobID, updateJobStatus, updateDedupID)

	switch item.JobType {
	case jobs.PDFToImage:
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.PDFToImage, jobID, item.Status)
		err = h.handlePDFToImage(job, item.JobDetails)
	case jobs.ImageToVideo:
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.ImageToVideo, jobID, item.Status)
		err = h.handleImageToVideo(job, item.JobDetails)
	case jobs.VideoConcat:
		h.Logger.Infof("JobType: %v JobID: %v JobStatus: %v", jobs.VideoConcat, jobID, item.Status)
		err = h.handleVideoConcat(job, item.JobDetails)
	default:
		h.Logger.Errorf("Invalid JobType - JobType: %v JobID: %v JobStatus: %v", jobs.VideoConcat, jobID, item.Status)
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

func (h UpdateJobStatus) handlePDFToImage(job jobs.Job, rawJobDetails []byte) error {
	switch job.Status {
	case jobs.SuccessStatus:
		h.Logger.Info("Handling Successful PDF To Image")

		type succcessfulJobDetails struct {
			SlideDetails []SlideDetail `json:"slide_details"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		var setters []func(*project.Project)
		for _, slideDetail := range jobDetails.SlideDetails {
			setters = append(setters, project.SetImage(slideDetail.ImageID, slideDetail.SlideNo))
		}
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, setters...)
		if err != nil {
			return fmt.Errorf("Unable to update project successfully")
		}

	case jobs.FailureStatus:
		h.Logger.Info("Failure PDF Split")
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Unable to update project successfully")
		}
		err = h.JobStore.DeleteJob(context.Background(), job.ID)
		if err != nil {
			return fmt.Errorf("Unable to delete job successfully")
		}

	case jobs.RunningStatus:
		h.Logger.Info("Running PDF Split")
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Unable to update project successfully")
		}

	default:
		h.Logger.Info("Unknown status set for the job")
		return fmt.Errorf("Unknown status set for job")
	}
	return nil
}

func (h UpdateJobStatus) handleImageToVideo(job jobs.Job, rawJobDetails []byte) error {
	switch job.Status {
	case jobs.SuccessStatus:
		h.Logger.Info("Successful Image to Video")
		type succcessfulJobDetails struct {
			ImageID    string `json:"image_id"`
			OutputFile string `json:"output_file"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		updatedProject, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetVideoID(jobDetails.ImageID, jobDetails.OutputFile))
		if err != nil {
			return fmt.Errorf("Unable to update project entity. err: %v", err)
		}
		successfulJobs, _ := h.JobStore.GetAllJobs(context.Background(), jobs.FilterRefID(job.RefID), jobs.FilterStatus(jobs.SuccessStatus))

		if len(updatedProject.SlideAssets) == len(successfulJobs) {
			videoConcatJob := jobs.NewJob(job.RefID, jobs.VideoConcat, "")
			var videoList []string
			for _, slideAsset := range updatedProject.SlideAssets {
				videoList = append(videoList, slideAsset.VideoID)
			}
			videoConcatJobDetails := map[string]interface{}{"id": videoConcatJob.ID, "video_ids": videoList}
			rawVideoConcatJobDetails, _ := json.Marshal(videoConcatJobDetails)
			videoConcatJob.Message = string(rawVideoConcatJobDetails)
			h.JobStore.CreateJob(context.Background(), videoConcatJob)
			h.VideoConcatQueue.Add(context.Background(), rawVideoConcatJobDetails)
			_, err = h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
			if err != nil {
				return fmt.Errorf("Error with trying to update project status of project on project store")
			}
		}

	case jobs.FailureStatus:
		h.Logger.Info("Failed Image to Video")
		// Even if one of the image to video creation fails, the whole project is on a failed state as all of the videos needs to be rendered regardless
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Error with trying to update project status of project on project store")
		}
		err = h.JobStore.DeleteJobs(context.Background(), jobs.FilterRefID(job.RefID))
		if err != nil {
			return fmt.Errorf("Issue with removing data")
		}

	case jobs.RunningStatus:
		h.Logger.Info("Running Image to Video")
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Error with trying to update project status of project on project store")
		}
	default:
		h.Logger.Info("Unknown status set for the job")
		return fmt.Errorf("Unknown status set for job")
	}
	return nil
}

func (h UpdateJobStatus) handleVideoConcat(job jobs.Job, rawJobDetails []byte) error {
	switch job.Status {
	case jobs.SuccessStatus:
		h.Logger.Info("Successful Video Concat")
		type succcessfulJobDetails struct {
			OutputVideo string `json:"output_video"`
		}

		var jobDetails succcessfulJobDetails
		json.Unmarshal(rawJobDetails, &jobDetails)

		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetVideoOutputID(jobDetails.OutputVideo), project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Unable to update project")
		}

		err = h.JobStore.DeleteJobs(context.Background(), jobs.FilterRefID(job.RefID))
		if err != nil {
			return fmt.Errorf("Issue with removing data")
		}

	case jobs.FailureStatus:
		h.Logger.Info("Failed Video Concat")
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Error with trying to update project status of project on project store")
		}
		err = h.JobStore.DeleteJob(context.Background(), job.ID)
		if err != nil {
			return fmt.Errorf("Error with trying to delete job")
		}
	case jobs.RunningStatus:
		h.Logger.Info("Running Video Concat")
		_, err := h.ProjectStore.UpdateProject(context.Background(), job.RefID, project.SetStatus(project.ProjectStatus(job.Status, job.Type)))
		if err != nil {
			return fmt.Errorf("Error with trying to update project status of project on project store")
		}
	default:
		h.Logger.Info("Unknown status set for the job")
		return fmt.Errorf("Unknown status set for job")
	}

	return nil
}
