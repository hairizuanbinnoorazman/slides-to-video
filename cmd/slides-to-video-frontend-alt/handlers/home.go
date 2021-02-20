package handlers

import (
	"html/template"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Home struct {
	Logger logger.Logger
}

func (h Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Home Handler")
	defer h.Logger.Info("End Home Handler")

	tmpl, err := template.ParseFiles("./templates/dashboard.html", "./templates/head.html", "./templates/header.html")
	if err != nil {
		h.Logger.Info("Error: %v", err)
		w.Write([]byte("Failed to render login page"))
		return
	}
	err = tmpl.Execute(w, nil)

	w.WriteHeader(http.StatusOK)
	return
}
