package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Login struct {
	IngressPath string
	Logger      logger.Logger
	MgrEndpoint string
}

func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Login Handler")
	defer h.Logger.Info("End Login Handler")

	tmpl, err := template.ParseFiles("./templates/login.html", "./templates/head.html", "./templates/header.html")
	if err != nil {
		h.Logger.Info("Error: %v", err)
		w.Write([]byte("Failed to render login page"))
		return
	}

	type LoginPage struct {
		MgrLoginURL string
	}

	sourceURL := r.URL.Query().Get("source_url")

	mgrURL := h.MgrEndpoint + fmt.Sprintf("/api/v1/login?source_url=%v", sourceURL)

	err = tmpl.Execute(w, LoginPage{MgrLoginURL: mgrURL})
	if err != nil {
		h.Logger.Info("Error: %v", err)
		w.Write([]byte("Failed to render login page"))
		return
	}
}
