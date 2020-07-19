package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/imageimporter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type CreatePDFSlideImages struct {
	Logger              logger.Logger
	PDFSlideImagesStore pdfslideimages.Store
	Blobstorage         blobstorage.BlobStorage
	BucketFolderName    string
	PDFSlideImporter    imageimporter.PDFImporter
}

func (h CreatePDFSlideImages) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start CreatePDFSlideImages API Handler")
	defer h.Logger.Info("End CreatePDFSlideImages API Handler")

	projectID := mux.Vars(r)["project_id"]
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve parse multipart form data. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	file, _, err := r.FormFile("myfile")
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

	slideImages := pdfslideimages.New(projectID)
	err = h.PDFSlideImagesStore.Create(context.Background(), slideImages)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save pdf slide images. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	h.Blobstorage.Save(context.Background(), h.BucketFolderName+"/"+slideImages.PDFFile, b.Bytes())

	err = h.PDFSlideImporter.Start(slideImages)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send job. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusCreated)
	rawItem, _ := json.Marshal(slideImages)
	w.Write(rawItem)
	return
}

type UpdatePDFSlideImages struct {
	Logger              logger.Logger
	PDFSlideImagesStore pdfslideimages.Store
	VideoSegmentStore   videosegment.Store
}

func (h UpdatePDFSlideImages) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start UpdatePDFSlideImages API Handler")
	defer h.Logger.Info("End UpdatePDFSlideImages API Handler")

	projectID := mux.Vars(r)["project_id"]
	pdfSlideImagesID := mux.Vars(r)["pdfslideimages_id"]
	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	type updatePDFSlideImagesReq struct {
		Status                  string                      `json:"status"`
		SlideAssets             []pdfslideimages.SlideAsset `json:"slide_assets"`
		ClearSetRunningIdemKey  string                      `json:"idem_key_set_running"`
		ClearCompleteRecIdemKey string                      `json:"idem_key_complete_rec"`
	}
	req := updatePDFSlideImagesReq{}
	json.Unmarshal(rawReq, &req)

	h.Logger.Info(req)

	var zz []func(*pdfslideimages.PDFSlideImages) error
	zz = append(zz, pdfslideimages.SetStatus(req.Status))
	zz = append(zz, pdfslideimages.SetSlideAssets(req.SlideAssets))
	if req.ClearSetRunningIdemKey != "" {
		zz = append(zz, pdfslideimages.ClearSetRunningIdemKey(req.ClearSetRunningIdemKey))
	}
	if req.ClearCompleteRecIdemKey != "" {
		zz = append(zz, pdfslideimages.ClearCompleteRecIdemKey(req.ClearCompleteRecIdemKey))
	}
	item, err := h.PDFSlideImagesStore.Update(context.Background(), projectID, pdfSlideImagesID, zz...)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to update record. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	if item.IsComplete() {
		for _, s := range item.SlideAssets {
			videoSegment := videosegment.New(projectID, s.ImageID, s.Order)
			err := h.VideoSegmentStore.Create(context.Background(), videoSegment)
			if err != nil {
				errMsg := fmt.Sprintf("Error - unable to update video segment. Error: %v", err)
				h.Logger.Error(errMsg)
				w.WriteHeader(500)
				w.Write([]byte(errMsg))
				return
			}
		}
	}

	rawItem, err := json.Marshal(item)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to render update pdfslides response. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawItem)
}

type GetPDFSlideImages struct {
	Logger              logger.Logger
	PDFSlideImagesStore pdfslideimages.Store
}

func (h GetPDFSlideImages) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start GetPDFSlideImages API Handler")
	defer h.Logger.Info("End GetPDFSlideImages API Handler")

	projectID := mux.Vars(r)["project_id"]
	pdfSlideImagesID := mux.Vars(r)["pdfslideimages_id"]
	pdfslideimages, err := h.PDFSlideImagesStore.Get(context.Background(), projectID, pdfSlideImagesID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve pdfslideimages. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	rawPDFSlideImages, _ := json.Marshal(pdfslideimages)

	w.WriteHeader(http.StatusOK)
	w.Write(rawPDFSlideImages)
	return
}
