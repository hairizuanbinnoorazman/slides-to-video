package blobstorage

import (
	"context"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	Logger     logger.Logger
	Client     *minio.Client
	BucketName string
}

func NewMinio(logger logger.Logger, endpoint, accessKeyID, secretAccessKey, bucketName string) (Minio, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return Minio{}, err
	}
	return Minio{
		Logger:     logger,
		Client:     minioClient,
		BucketName: bucketName,
	}, nil
}

func (b Minio) Save(ctx context.Context, fileName string, content []byte) error {
	return nil
}

func (b Minio) Load(ctx context.Context, fileName string) (content []byte, err error) {
	return []byte{}, nil
}
