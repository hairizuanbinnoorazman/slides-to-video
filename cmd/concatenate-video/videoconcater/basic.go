package videoconcater

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/mgrclient"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"gopkg.in/go-playground/validator.v9"
)

type Basic struct {
	logger            logger.Logger
	blobStorage       blobstorage.BlobStorage
	mgrClient         mgrclient.Client
	inputVideoFolder  string
	outputVideoFolder string
}

func NewBasic(l logger.Logger, store blobstorage.BlobStorage, cl mgrclient.Client, inputFolder, outputFolder string) Basic {
	return Basic{
		logger:            l,
		blobStorage:       store,
		mgrClient:         cl,
		inputVideoFolder:  inputFolder,
		outputVideoFolder: outputFolder,
	}
}

func (h *Basic) Process(ctx context.Context, job JobDetails) error {
	v := validator.New()
	err := v.Struct(job)
	if err != nil {
		h.logger.Errorf("Error validating for struct. Err: %v", err)
		return err
	}
	h.mgrClient.UpdateRunning(ctx, job.AuthToken, job.ID, job.RunningIdemKey)

	videosToBeCombined := ""
	for _, videoID := range job.VideoIDs {
		videoContent, err := h.blobStorage.Load(context.TODO(), videoID)
		if err != nil {
			h.mgrClient.FailedTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey)
			return fmt.Errorf("Error while to download video. Error: %v. VideoID: %v", err, videoID)
		}
		defer os.Remove(videoID)

		err = ioutil.WriteFile(videoID, videoContent, 777)
		if err != nil {
			h.mgrClient.FailedTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey)
			return fmt.Errorf("Error while writing video content for specific video snippet file. Error: %v. VideoID: %v", err, videoID)
		}
		videosToBeCombined = videosToBeCombined + fmt.Sprintf("file %s\n", videoID)
	}

	combinedVideoListFileName := fmt.Sprintf("combined_%s.txt", job.ID)
	combinedVideoFileName := job.ID + ".mp4"
	h.logger.Infof("Videos to be combined: %v", videosToBeCombined)
	err = ioutil.WriteFile(combinedVideoListFileName, []byte(videosToBeCombined), 777)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Error while combining videos. Error: %v", err)
	}
	defer os.Remove(combinedVideoListFileName)

	err = combineVideo(combinedVideoListFileName, combinedVideoFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Error while combining videos. Error: %v", err)
	}
	defer os.Remove(combinedVideoFileName)

	videoContent, err := ioutil.ReadFile(combinedVideoFileName)
	err = h.blobStorage.Save(context.TODO(), combinedVideoFileName, videoContent)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Error while combining videos. Error: %v", err)
	}

	h.mgrClient.CompleteTask(ctx, job.AuthToken, job.ID, job.CompleteRecIdemKey, combinedVideoFileName)

	return nil
}
