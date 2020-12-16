package image2videoconverter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/image-to-video/mgrclient"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"gopkg.in/go-playground/validator.v9"
)

type basic struct {
	logger              logger.Logger
	blobStorage         blobstorage.BlobStorage
	mgrClient           mgrclient.Client
	imagesFolder        string
	videoSnippetsFolder string
	textToSpeechEngine  TextToSpeechEngine
}

func NewBasic(l logger.Logger, store blobstorage.BlobStorage, mgr mgrclient.Client, imagesFolder, videoSnippetsFolder string, engine TextToSpeechEngine) basic {
	return basic{
		logger:              l,
		blobStorage:         store,
		mgrClient:           mgr,
		imagesFolder:        imagesFolder,
		videoSnippetsFolder: videoSnippetsFolder,
		textToSpeechEngine:  engine,
	}
}

func (h *basic) Process(ctx context.Context, job JobDetails) error {
	v := validator.New()
	err := v.Struct(job)
	if err != nil {
		h.logger.Errorf("Error validating for struct. Err: %v", err)
		return err
	}
	h.mgrClient.UpdateRunning(ctx, job.ProjectID, job.ID, job.RunningIdemKey)

	imageFileName := job.ImageID
	audioFileName := job.ID + ".mp3"
	adjustedAudioFileName := "adjusted_" + job.ID + ".mp3"
	convertedAudioFileName := "converted_" + job.ID + ".m4a"
	silentVideoFileName := "silent_" + job.ID + ".mp4"
	outputVideoFileName := job.ID + ".mp4"
	defer func() {
		// Cleanup
		os.Remove(imageFileName)
		os.Remove(audioFileName)
		os.Remove(silentVideoFileName)
		os.Remove(outputVideoFileName)
		os.Remove(adjustedAudioFileName)
		os.Remove(convertedAudioFileName)
	}()

	rawImage, err := h.blobStorage.Load(ctx, h.imagesFolder+"/"+imageFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to load image from blobstorage. Err: %v", err)
	}
	err = ioutil.WriteFile(imageFileName, rawImage, 777)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to write image to file system for further processing. Err: %v", err)
	}

	audioContent, err := h.textToSpeechEngine.Generate(job.Text)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to retrieve speech content from Google Cloud. Err: %v", err)
	}
	err = ioutil.WriteFile(audioFileName, audioContent, 777)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to write speech to file system for further processing. Err: %v", err)
	}

	err = addSilentAudio(audioFileName, adjustedAudioFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to create silent audio. Err: %v", err)
	}

	err = convertToUseAAC(adjustedAudioFileName, convertedAudioFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to convert audio to be acc format. Err: %v", err)
	}

	audioDuration, err := getAudioDuration(convertedAudioFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to get duration of the audio. Err: %v", err)
	}

	err = generateSilentVideo(imageFileName, audioDuration, silentVideoFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to generate the silent video. Err: %v", err)
	}

	err = muxSilentVideoAndAudio(silentVideoFileName, convertedAudioFileName, outputVideoFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to mux the silent video and audio into a single video. Err: %v", err)
	}

	videoContent, err := ioutil.ReadFile(outputVideoFileName)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to write the video content file to fs. Err: %v", err)
	}

	err = h.blobStorage.Save(ctx, outputVideoFileName, videoContent)
	if err != nil {
		h.mgrClient.FailedTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey)
		return fmt.Errorf("Unable to store file into blob storage. Err: %v", err)
	}

	h.mgrClient.CompleteTask(ctx, job.ProjectID, job.ID, job.CompleteRecIdemKey, outputVideoFileName)
	return nil
}
