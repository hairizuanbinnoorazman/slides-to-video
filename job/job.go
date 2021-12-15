// package job is meant to deal to handle and manage the processing of long term jobs
package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videoconcater"
)

type jobtype string

var (
	TriggerVideoConcat jobtype = "trigger video concat"
)

type Job struct {
	ID         string    `json:"id" gorm:"type:varchar(40);primary_key"`
	ProjectID  string    `json:"project_id" gorm:"unique"`
	UserID     string    `json:"user_id" gorm:"type:varchar(40)"`
	JobType    jobtype   `json:"job_type" gorm:"type:varchar(200)"`
	StartTime  time.Time `json:"start_time"`
	ExpiryTime time.Time `json:"expiry_time"`
}

func New(projectID, userID string) Job {
	jobID, _ := uuid.NewV4()
	return Job{
		ID:         jobID.String(),
		ProjectID:  projectID,
		UserID:     userID,
		JobType:    TriggerVideoConcat,
		StartTime:  time.Now(),
		ExpiryTime: time.Now().Add(1 * time.Hour),
	}
}

func NewProcessor(logger logger.Logger, jobStore Store, projectStore project.Store, videoconcater videoconcater.VideoConcater) (Processor, error) {
	if logger == nil || jobStore == nil || projectStore == nil || videoconcater == nil {
		return Processor{}, fmt.Errorf("cannot start processor as one of the inputs to processor is nil")
	}

	return Processor{
		logger:        logger,
		jobsStore:     jobStore,
		projectStore:  projectStore,
		videoconcater: videoconcater,
	}, nil
}

type Processor struct {
	logger        logger.Logger
	jobsStore     Store
	projectStore  project.Store
	videoconcater videoconcater.VideoConcater
}

func (p Processor) Start() {
	p.logger.Info("Start processor")
	for {
		var wg sync.WaitGroup
		maxWorkerCount := 5
		workerCount := maxWorkerCount

		p.logger.Info("Begin long running job")
		jobs, err := p.jobsStore.GetAll(context.TODO(), workerCount, 0)
		if err != nil {
			p.logger.Errorf("No jobs obtained from db :: Err %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if len(jobs) < maxWorkerCount {
			workerCount = len(jobs)
		}

		for i := 0; i < workerCount; i++ {
			wg.Add(1)

			singleJob := jobs[i]

			go func() {
				defer wg.Done()
				if singleJob.JobType == TriggerVideoConcat {
					go p.processTriggerVideoConcat(singleJob)
				}
			}()
		}

		wg.Wait()

		time.Sleep(10 * time.Second)
	}

}

// processTriggerVideoConcat monitors state of video segments and once
// the video segments have completed running, will trigger video concat
func (p Processor) processTriggerVideoConcat(j Job) {
	if time.Now().After(j.ExpiryTime) {
		p.logger.Errorf("job expired. will delete job. ProjectID - %v")
		p.jobsStore.Delete(context.TODO(), j.ID)
	}

	project, err := p.projectStore.Get(context.TODO(), j.ProjectID, j.UserID)
	if err != nil {
		p.logger.Errorf("unable to get project details. will retry. ProjectID - %v :: Error - %v", j.ProjectID, err)
		return
	}

	if len(project.VideoSegments) == 0 {
		p.logger.Errorf("no video segments created - job will never need to trigger video concatenation")
		p.jobsStore.Delete(context.TODO(), j.ID)
		return
	}

	completedStatusCount := 0
	for _, v := range project.VideoSegments {
		if v.Status == "completed" {
			completedStatusCount = completedStatusCount + 1
		}
	}

	if completedStatusCount < len(project.VideoSegments) {
		p.logger.Infof("still processing - ProjectID - %v", j.ProjectID)
		return
	} else if completedStatusCount > len(project.VideoSegments) {
		p.logger.Errorf("unexpected count of video segments. ProjectID - %v :: VideoSegmentCount - %v :: CompletedCount - %v", j.ProjectID, len(project.VideoSegments), completedStatusCount)
		p.jobsStore.Delete(context.TODO(), j.ID)
		return
	}

	p.logger.Infof("video segment processing completed. Will trigger video concat - ProjectID - %v", j.ProjectID)
	vidSegmentList, err := project.GetVideoSegmentList()
	if err != nil {
		p.logger.Errorf("unable to get video segment list from project. will retry. ProjectID - %v :: Err - %v", j.ProjectID, err)
	}
	err = p.videoconcater.Start(context.TODO(), j.ProjectID, j.UserID, vidSegmentList)
	if err != nil {
		p.logger.Errorf("unable to send msg to start video concatenation")
	}
	p.jobsStore.Delete(context.TODO(), j.ID)
}
