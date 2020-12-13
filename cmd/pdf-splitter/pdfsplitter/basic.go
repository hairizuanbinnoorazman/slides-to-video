package pdfsplitter

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/mgrclient"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type basic struct {
	Logger               logger.Logger
	SlidesToVideoStorage blobstorage.BlobStorage
	MgrClient            mgrclient.Client
	PDFFolder            string
	ImagesFolder         string
}

func NewBasic(logger logger.Logger, storage blobstorage.BlobStorage, mgrclient mgrclient.Client, pdfFolder, imageFolder string) basic {
	return basic{
		Logger:               logger,
		SlidesToVideoStorage: storage,
		MgrClient:            mgrclient,
		PDFFolder:            pdfFolder,
		ImagesFolder:         imageFolder,
	}
}

func (h *basic) Process(job PdfSplitJob) error {
	h.MgrClient.UpdateRunning(context.Background(), job.ProjectID, job.ID, job.IdemKeySetRunning)

	if job.Validate() != nil {
		h.MgrClient.FailedTask(context.Background(), job.ProjectID, job.ID, job.IdemKeyCompleteRec)
		return fmt.Errorf("%+v", job.Validate())
	}

	content, err := h.SlidesToVideoStorage.Load(context.Background(), h.PDFFolder+"/"+job.PdfFileName)
	if err != nil {
		return fmt.Errorf("Error occured while loading file: %v, %v", h.PDFFolder+"/"+job.PdfFileName, err)
	}

	ioutil.WriteFile(job.PdfFileName, content, os.FileMode(0777))

	cmd := exec.Command("convert", "-density", "150", job.ID+".pdf", "-quality", "90", job.ID+".png")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error occured while splitting files %v. Stdout: %v, Stderr: %v", err, out.String(), stderr.String())
	}

	fileInfo, err := ioutil.ReadDir(".")
	if err != nil {
		h.MgrClient.FailedTask(context.Background(), job.ProjectID, job.ID, job.IdemKeyCompleteRec)
		return fmt.Errorf("Error occured while getting file info %v", err)
	}

	var fileList []string
	for _, info := range fileInfo {
		isOutput := strings.Contains(info.Name(), job.ID) && strings.Contains(info.Name(), "png")
		if isOutput {
			fileList = append(fileList, info.Name())
		}
	}

	for _, file := range fileList {
		content, _ := ioutil.ReadFile(file)
		err = h.SlidesToVideoStorage.Save(context.Background(), h.ImagesFolder+"/"+file, content)
		if err != nil {
			h.MgrClient.FailedTask(context.Background(), job.ProjectID, job.ID, job.IdemKeyCompleteRec)
			return fmt.Errorf("Error occured while saving %v", err)
		}
	}

	// Reporting to manager
	var slideDetails []mgrclient.SlideAsset
	for _, f := range fileList {
		splitFileName := strings.Split(f, "-")
		slideNoAndFileFormat := strings.Split(splitFileName[len(splitFileName)-1], ".")
		h.Logger.Warningf("Split File Name: %+v", splitFileName)
		h.Logger.Warningf("Slide No and File Format: %+v", slideNoAndFileFormat)
		s := mgrclient.SlideAsset{
			ImageID: f,
			Order: func() int {
				if len(slideNoAndFileFormat) == 0 {
					h.Logger.Errorf("Unable to retrieve the value of slide no. %v %v", slideNoAndFileFormat)
				}
				num, err := strconv.Atoi(slideNoAndFileFormat[0])
				if err != nil {
					h.Logger.Errorf("Unable to convert the number to required form. %v %v", slideNoAndFileFormat[0], err)
				}
				return num
			}(),
		}
		slideDetails = append(slideDetails, s)
	}

	fileList = append(fileList, job.PdfFileName)

	// Cleanup of files
	for _, info := range fileList {
		err = os.Remove(info)
		if err != nil {
			return fmt.Errorf("Unable to remove file %v. Err: %v", info, err)
		}
	}

	err = h.MgrClient.CompleteTask(context.Background(), job.ProjectID, job.ID, job.IdemKeyCompleteRec, slideDetails)
	if err != nil {
		h.MgrClient.FailedTask(context.Background(), job.ProjectID, job.ID, job.IdemKeyCompleteRec)
		return fmt.Errorf("Error occured while saving %v", err)
	}
	return nil
}
