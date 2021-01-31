package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Projects struct {
	Logger logger.Logger
}

func (h Projects) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Projects Handler")
	defer h.Logger.Info("End Projects Handler")

	c, err := r.Cookie("STVID")
	if err != nil {
		http.Redirect(w, r, "/healthz", http.StatusTemporaryRedirect)
		return
	}

	h.Logger.Info("Cookie found? :: %v :: %v", c.Name, c.Value)

	type resp struct {
		Status string `json:"status"`
	}
	raw, _ := json.Marshal(resp{Status: "found"})
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
	return
}
