package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type RequireLogin struct {
	Scheme      string
	IngressPath string
	Logger      logger.Logger
	NextHandler http.Handler
}

func (h RequireLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start RequireLogin Decorator")
	defer h.Logger.Info("End RequireLogin Decorator")

	sourceURL := r.URL.String()

	token := r.URL.Query().Get("token")

	if token != "" {
		expire := time.Now().Add(20 * time.Minute)
		c := http.Cookie{
			Name:    "STVID",
			Value:   token,
			Path:    "/",
			Expires: expire,
			MaxAge:  86400,
		}
		http.SetCookie(w, &c)
	} else {
		_, err := r.Cookie("STVID")
		if err != nil {
			h.Logger.Info("Cookie unset - redirect to login page")
			http.Redirect(w, r, h.IngressPath+fmt.Sprintf("%v://%v/login?source_url=%v://%v%v", h.Scheme, r.Host, h.Scheme, r.Host, sourceURL), http.StatusTemporaryRedirect)
			return
		}
	}

	h.NextHandler.ServeHTTP(w, r)
	return
}
