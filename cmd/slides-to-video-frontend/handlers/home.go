package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Home struct {
	Logger logger.Logger
}

func (h Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Home Handler")
	defer h.Logger.Info("End Home Handler")

	type resp struct {
		Status string `json:"status"`
	}
	raw, _ := json.Marshal(resp{Status: "ok"})
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
	return
}
