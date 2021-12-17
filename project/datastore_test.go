package project_test

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/testcontainers/testcontainers-go"
)

func Test_datastore_ops(t *testing.T) {
	// Following command is similar to this docker command:
	// docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=test-database -e MYSQL_USER=user -e MYSQL_PASSWORD=password -d -p 3306:3306 mysql:5.7
	req, _ := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "google/cloud-sdk:latest",
			Name:         "some-datastore",
			Cmd:          []string{"gcloud", "beta", "emulators", "datastore", "start", "--host-port", "0.0.0.0:8432", "--project", "test-datastore"},
			ExposedPorts: []string{"8432/tcp"},
		},
		Started: true,
	})
	defer req.Terminate(context.TODO())
	port, err := req.MappedPort(context.TODO(), "8432")

	os.Setenv("DATASTORE_DATASET", "test-datastore")
	os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:"+port.Port())
	os.Setenv("DATASTORE_EMULATOR_HOST_PATH", "localhost:"+port.Port()+"/datastore")
	os.Setenv("DATASTORE_HOST", "http://localhost:"+port.Port())
	os.Setenv("DATASTORE_PROJECT_ID", "test-datastore")

	time.Sleep(10 * time.Second)

	xClient, err := datastore.NewClient(context.Background(), "test-datastore")
	if err != nil {
		t.Fatalf("Unable to connect to datastore. Err :: %v", err)
	}
	projectStore := project.NewGoogleDatastore(xClient, "project", "pdfslideimages", "videosegments")

	p := project.Project{
		ID:           "1234",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
	p2 := project.Project{
		ID:           "1235",
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}

	err = projectStore.Create(context.TODO(), p)
	if err != nil {
		t.Fatalf("Unable to create project record in datastore. Record :: %+v, Err :: %v", p, err)
	}
	err = projectStore.Create(context.TODO(), p2)
	if err != nil {
		t.Fatalf("Unable to create project record in datastore. Record :: %+v, Err :: %v", p2, err)
	}

}

// func TestGoogleDatastore_StoreProject(t *testing.T) {
// 	type fields struct {
// 		EntityName string
// 		Client     *datastore.Client
// 	}
// 	type args struct {
// 		ctx context.Context
// 		e   project.Project
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "Successful test",
// 			fields: fields{
// 				EntityName: "test-test",
// 				Client:     datastoreClientHelper(),
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				e: project.Project{
// 					ID: "Initial verison",
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			g := project.NewGoogleDatastore(tt.fields.Client, tt.fields.EntityName)
// 			if err := g.Create(tt.args.ctx, tt.args.e); (err != nil) != tt.wantErr {
// 				t.Errorf("GoogleDatastore.CreateProject() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestGoogleDatastore_GetProject(t *testing.T) {
// 	type fields struct {
// 		EntityName string
// 		Client     *datastore.Client
// 	}
// 	type args struct {
// 		ctx context.Context
// 		ID  string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    project.Project
// 		wantErr bool
// 	}{
// 		{
// 			name: "Succesful case",
// 			fields: fields{
// 				EntityName: "test-test",
// 				Client:     datastoreClientHelper(),
// 			},
// 			args: args{
// 				ctx: context.Background(),
// 				ID:  "Initial verison",
// 			},
// 			want: project.Project{
// 				ID: "Initial verison",
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			g := project.NewGoogleDatastore(tt.fields.Client, tt.fields.EntityName)
// 			got, err := g.Get(tt.args.ctx, tt.args.ID, "")
// 			location, _ := time.LoadLocation("UTC")
// 			got.DateModified = got.DateModified.In(location)
// 			got.DateCreated = got.DateCreated.In(location)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GoogleDatastore.GetProject() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GoogleDatastore.GetProject() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
