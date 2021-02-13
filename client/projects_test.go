package client

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
)

func TestClient_GetProjects(t *testing.T) {
	type fields struct {
		mgrURL     string
		httpClient *http.Client
		logger     logger.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    []project.Project
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				mgrURL:     "http://localhost:8880",
				httpClient: http.DefaultClient,
				logger:     logger.LoggerForTests{Tester: t},
			},
			want:    []project.Project{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				mgrURL:     tt.fields.mgrURL,
				httpClient: tt.fields.httpClient,
				logger:     tt.fields.logger,
			}
			got, err := c.GetProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetProjects() = %v, want %v", got, tt.want)
			}
		})
	}
}
