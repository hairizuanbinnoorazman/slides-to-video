package main

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

func datastoreClientHelper() *datastore.Client {
	credJSON, _ := ioutil.ReadFile("./slides-to-video-manager.json")
	xClient, _ := datastore.NewClient(context.Background(), "expanded-league-162223", option.WithCredentialsJSON(credJSON))
	return xClient
}

func TestGoogleDatastore_GetParentJob(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx context.Context
		ID  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ParentJob
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				EntityName: "SlidesTest",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				ID:  "1234",
			},
			want: ParentJob{
				ID:               "1234",
				OriginalFilename: "aaa",
				Filename:         "aaa",
				Script:           "{\"accacaca\": 12}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			got, err := g.GetParentJob(tt.args.ctx, tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.GetParentJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleDatastore.GetParentJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleDatastore_StoreImageToVideoJob(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx context.Context
		e   ImageToVideoJob
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				EntityName: "test-ImageToVideoJob",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				e: ImageToVideoJob{
					ID:          "aaaa",
					ParentJobID: "aaab",
					ImageID:     "aaaa",
					Text:        "nnajknc",
					Status:      "created",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			if err := g.StoreImageToVideoJob(tt.args.ctx, tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.StoreImageToVideoJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoogleDatastore_GetAllImageToVideoJobs(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx              context.Context
		filterByParentID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []ImageToVideoJob
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				EntityName: "test-ImageToVideoJob",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx:              context.Background(),
				filterByParentID: "aaaa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			got, err := g.GetAllImageToVideoJobs(tt.args.ctx, tt.args.filterByParentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.GetAllImageToVideoJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleDatastore.GetAllImageToVideoJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}
