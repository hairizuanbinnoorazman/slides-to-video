package jobs_test

import (
	"context"
	"io/ioutil"
	"reflect"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"
	"google.golang.org/api/option"
)

func datastoreClientHelper() *datastore.Client {
	credJSON, _ := ioutil.ReadFile("../slides-to-video-manager.json")
	xClient, _ := datastore.NewClient(context.Background(), "expanded-league-162223", option.WithCredentialsJSON(credJSON))
	return xClient
}

func TestGoogleDatastore_StoreParentJob(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx context.Context
		e   jobs.ParentJob
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
				EntityName: "testing-case",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				e: jobs.ParentJob{
					ID:               "test3",
					OriginalFilename: "test3",
					Filename:         "test3",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &jobs.GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			if err := g.StoreParentJob(tt.args.ctx, tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.StoreParentJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
		want    jobs.ParentJob
		wantErr bool
	}{
		{
			name: "Successful get",
			fields: fields{
				EntityName: "testing-case",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				ID:  "test",
			},
			want: jobs.ParentJob{
				ID:               "test",
				OriginalFilename: "test",
				Filename:         "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &jobs.GoogleDatastore{
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

func TestGoogleDatastore_GetAllParentJobs(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []jobs.ParentJob
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				EntityName: "testing-case",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
			},
			want: []jobs.ParentJob{
				jobs.ParentJob{
					ID:               "test",
					OriginalFilename: "test",
					Filename:         "test",
				},
				jobs.ParentJob{
					ID:               "test2",
					OriginalFilename: "test",
					Filename:         "test",
				},
				jobs.ParentJob{
					ID:               "test3",
					OriginalFilename: "test3",
					Filename:         "test3",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &jobs.GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			got, err := g.GetAllParentJobs(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.GetAllParentJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleDatastore.GetAllParentJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}
