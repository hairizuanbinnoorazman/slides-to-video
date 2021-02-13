// Package client wraps up golang functionality for dealing and contacting the manager component
package client

import (
	"net/http"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type Client struct {
	mgrURL     string
	httpClient *http.Client
	logger     logger.Logger
}

func NewClient(mgrURL string, httpClient *http.Client, logger logger.Logger) *Client {
	return &Client{
		mgrURL:     mgrURL,
		httpClient: httpClient,
		logger:     logger,
	}
}
