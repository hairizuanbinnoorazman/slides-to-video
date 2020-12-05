package project_test

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

func datastoreClientHelper() *datastore.Client {
	credJSON, _ := ioutil.ReadFile("../slides-to-video-manager.json")
	xClient, _ := datastore.NewClient(context.Background(), "XXXX", option.WithCredentialsJSON(credJSON))
	return xClient
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
