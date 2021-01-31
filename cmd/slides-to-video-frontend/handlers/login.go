package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Login struct {
	IngressPath string
	Logger      logger.Logger
}

func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Login Handler")
	defer h.Logger.Info("End Login Handler")

	expire := time.Now().Add(20 * time.Minute)
	c := http.Cookie{
		Name:    "STVID",
		Value:   "nonsecureuser",
		Path:    "/",
		Expires: expire,
		MaxAge:  86400,
	}

	http.SetCookie(w, &c)

	destinationPath := r.URL.Query().Get("destination_path")
	if destinationPath == "" {
		destinationPath = strings.ReplaceAll(h.IngressPath+"/", "//", "/")
	}

	http.Redirect(w, r, destinationPath, http.StatusTemporaryRedirect)
}
