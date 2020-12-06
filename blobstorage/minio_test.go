package blobstorage

import (
	"context"
	"testing"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
)

func minioClientHelper(port int) (*minio.Client, error) {
	minioClient, err := minio.New("localhost:9999", &minio.Options{
		Creds: credentials.NewStaticV4("s3_user", "s3_password", ""),
	})
	return minioClient, err
}

func TestMinio_Save(t *testing.T) {
	compose := testcontainers.NewLocalDockerCompose([]string{"docker-compose.yaml"}, "1234")
	execError := compose.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	err := execError.Error
	if err != nil {
		t.Fatalf("Setup failed for minio setup. Err: %v", err)
	}
	time.Sleep(5 * time.Second)
	defer compose.Down()

	mc, err := minioClientHelper(9999)
	if err != nil {
		t.Fatalf("Unable to connect to setup client connection. Err: %v", err)
	}

	type fields struct {
		Logger     logger.Logger
		Client     *minio.Client
		BucketName string
	}
	type args struct {
		ctx      context.Context
		fileName string
		content  []byte
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
				Logger:     logger.LoggerForTests{Tester: t},
				Client:     mc,
				BucketName: "test-bucket",
			},
			args: args{
				ctx:      context.TODO(),
				fileName: "test/test",
				content:  []byte("acjknakcnk"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Minio{
				Logger:     tt.fields.Logger,
				Client:     tt.fields.Client,
				BucketName: tt.fields.BucketName,
			}
			if err := b.Save(tt.args.ctx, tt.args.fileName, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("Minio.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
