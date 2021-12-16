package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type DownloadVideo struct {
	Logger        logger.Logger
	StorageClient blobstorage.BlobStorage
}

func (h DownloadVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Download Handler")
	defer h.Logger.Info("End Download Handler")

	filename := mux.Vars(r)["video_id"]
	if filename == "" {
		errMsg := "Missing video id field"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	content, err := h.StorageClient.Load(context.Background(), filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to download file from blob storage. Err: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	w.WriteHeader(200)
	w.Write(content)
}
