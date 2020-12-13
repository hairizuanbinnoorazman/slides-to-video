package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/pdfsplitter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type ProcessHandler struct {
	Logger      logger.Logger
	PDFSplitter pdfsplitter.PDFSplitter
}

func (h ProcessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Process Handler")
	defer h.Logger.Info("End Process Handler")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.Logger.Errorf("Error in reading body of request")
		w.Write([]byte("Error"))
		w.WriteHeader(200)
		return
	}

	message := PubsubMsg{}
	json.Unmarshal(data, &message)
	h.Logger.Infof("%+v", message)
	decodedMsg, _ := base64.StdEncoding.DecodeString(message.Message.Data)
	h.Logger.Infof("Decoded message %+v", string(decodedMsg))

	job := pdfsplitter.PdfSplitJob{}
	err = json.Unmarshal(decodedMsg, &job)
	if err != nil {
		w.WriteHeader(200)
		w.Write([]byte("Error"))
		h.Logger.Errorf("%+v", err)
		return
	}

	h.Logger.Infof("%+v", job)

	err = h.PDFSplitter.Process(job)
	if err != nil {
		w.WriteHeader(200)
		w.Write([]byte("Error"))
		h.Logger.Errorf("%+v", err)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success"))
	return
}

type PubsubMsg struct {
	Message struct {
		Data        string `json:"data"`
		MessageID   string `json:"messageId"`
		PublishTime string `json:"publishTime"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}
