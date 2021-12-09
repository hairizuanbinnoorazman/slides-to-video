package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
)

type RequireJWTAuth struct {
	Auth        Auth
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

	type failedResp struct {
		Msg string `json:"msg"`
	}

	rawErrMsg, _ := json.Marshal(failedResp{Msg: "Invalid JWT Token Provided"})

	rawAuthorizationToken := copiedReq.Header.Get("Authorization")
	userID, err := services.ExtractToken(rawAuthorizationToken, a.Auth.Secret)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(rawErrMsg)
		return
	}

	ctx = context.WithValue(ctx, userIDKey, userID)

	a.NextHandler.ServeHTTP(w, r.WithContext(ctx))
}
