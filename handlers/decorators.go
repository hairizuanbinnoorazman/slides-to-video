package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
)

type RequireJWTAuth struct {
	Auth        services.Auth
	Logger      logger.Logger
	NextHandler http.Handler
}

var (
	userIDKey = "userID"
)

func (a RequireJWTAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("RequireJWTAuth Exists Check")

	ctx := r.Context()
	copiedReq := r.Clone(ctx)
	s := securecookie.New(a.Auth.HashKey, a.Auth.BlockKey)
	value := make(map[string]string)
	var userID string
	var err error

	type failedResp struct {
		Msg string `json:"msg"`
	}
	rawErrMsg, _ := json.Marshal(failedResp{Msg: "Invalid authorization token"})

	cookie, cookieErr := r.Cookie(a.Auth.CookieName)
	if cookieErr == nil {
		err := s.Decode(a.Auth.CookieName, cookie.Value, &value)
		if err != nil || value["user_id"] == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(rawErrMsg)
			return
		}
		userID = value["user_id"]
	} else {
		a.Logger.Error(cookieErr)
	}

	if userID == "" {
		rawAuthorizationToken := copiedReq.Header.Get("Authorization")
		userID, err = services.ExtractToken(rawAuthorizationToken, a.Auth.Secret)
		if err != nil || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(rawErrMsg)
			a.Logger.Info("Cookie is empty but header is also empty")
			return
		}
	}

	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(rawErrMsg)
		a.Logger.Info("UserID is empty after checking cookie and header")
		return
	}

	ctx = context.WithValue(ctx, userIDKey, userID)

	a.NextHandler.ServeHTTP(w, r.WithContext(ctx))
}
