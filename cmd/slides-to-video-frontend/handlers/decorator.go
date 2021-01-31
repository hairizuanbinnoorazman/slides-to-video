package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type RequireLogin struct {
	IngressPath string
	Logger      logger.Logger
	NextHandler http.Handler
}

func (h RequireLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start RequireLogin Decorator")
	defer h.Logger.Info("End RequireLogin Decorator")

	sourcePath := r.URL.Path

	_, err := r.Cookie("STVID")
	if err != nil {
		h.Logger.Info("Cookie unset - redirect to login page")
		http.Redirect(w, r, strings.ReplaceAll(h.IngressPath+fmt.Sprintf("/login?destination_path=%v", sourcePath), "//", "/"), http.StatusTemporaryRedirect)
		return
	}

	h.NextHandler.ServeHTTP(w, r)
	return
}
