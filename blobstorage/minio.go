package blobstorage

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

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
	if b.Client == nil {
		return fmt.Errorf("S3 Client not initialized")
	}
	_, err := b.Client.PutObject(ctx, b.BucketName, fileName, bytes.NewReader(content), -1, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (b Minio) Load(ctx context.Context, fileName string) (content []byte, err error) {
	obj, err := b.Client.GetObject(ctx, b.BucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return []byte{}, err
	}
	rawData, err := ioutil.ReadAll(obj)
	if err != nil {
		return []byte{}, err
	}
	return rawData, nil
}
