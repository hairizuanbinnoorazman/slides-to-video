package project_test

import (
	"context"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"google.golang.org/api/option"
)

func datastoreClientHelper() *datastore.Client {
	credJSON, _ := ioutil.ReadFile("../slides-to-video-manager.json")
	xClient, _ := datastore.NewClient(context.Background(), "expanded-league-162223", option.WithCredentialsJSON(credJSON))
	return xClient
}

func TestGoogleDatastore_StoreProject(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx context.Context
		e   project.Project
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful test",
			fields: fields{
				EntityName: "test-test",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				e: project.Project{
					ID:      "Initial verison",
					PDFFile: "testtest",
					SlideAssets: []project.SlideAsset{
						project.SlideAsset{
							ImageID: "name",
							VideoID: "namename",
							SlideNo: 0,
						},
						project.SlideAsset{
							ImageID: "name1",
							VideoID: "namename1",
							SlideNo: 1,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &project.GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			if err := g.CreateProject(tt.args.ctx, tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.CreateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoogleDatastore_GetProject(t *testing.T) {
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
		want    project.Project
		wantErr bool
	}{
		{
			name: "Succesful case",
			fields: fields{
				EntityName: "test-test",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				ID:  "Initial verison",
			},
			want: project.Project{
				ID:      "Initial verison",
				PDFFile: "testtest",
				SlideAssets: []project.SlideAsset{
					project.SlideAsset{
						ImageID: "name",
						VideoID: "namename",
						SlideNo: 0,
					},
					project.SlideAsset{
						ImageID: "name1",
						VideoID: "namename1",
						SlideNo: 1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &project.GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			got, err := g.GetProject(tt.args.ctx, tt.args.ID)
			location, _ := time.LoadLocation("UTC")
			got.DateModified = got.DateModified.In(location)
			got.DateCreated = got.DateCreated.In(location)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.GetProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GoogleDatastore.GetProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleDatastore_UpdateProject(t *testing.T) {
	type fields struct {
		EntityName string
		Client     *datastore.Client
	}
	type args struct {
		ctx     context.Context
		ID      string
		setters []func(*project.Project)
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
				EntityName: "test-test",
				Client:     datastoreClientHelper(),
			},
			args: args{
				ctx: context.Background(),
				ID:  "Initial verison",
				setters: []func(*project.Project){
					project.SetEmptySlideAsset(),
					project.SetImage("first", 0),
					project.SetImage("second", 1),
					project.SetSlideText("first", "this is a test"),
					project.SetPDFFile("this is a test.pdf"),
					project.SetVideoOutputID("this is a test video"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &project.GoogleDatastore{
				EntityName: tt.fields.EntityName,
				Client:     tt.fields.Client,
			}
			if err := g.UpdateProject(tt.args.ctx, tt.args.ID, tt.args.setters...); (err != nil) != tt.wantErr {
				t.Errorf("GoogleDatastore.UpdateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoogleDatastore_SortProject(t *testing.T) {
	tests := []struct {
		name     string
		unsorted []project.SlideAsset
		expected []project.SlideAsset
	}{
		{
			name: "Successful case",
			unsorted: []project.SlideAsset{
				project.SlideAsset{
					ImageID: "first",
					SlideNo: 1,
				},
				project.SlideAsset{
					ImageID: "zeroth",
					SlideNo: 0,
				},
			},
			expected: []project.SlideAsset{
				project.SlideAsset{
					ImageID: "zeroth",
					SlideNo: 0,
				},
				project.SlideAsset{
					ImageID: "first",
					SlideNo: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(project.BySlideNo(tt.unsorted))
			reflect.DeepEqual(tt.unsorted, tt.expected)
		})
	}
}
