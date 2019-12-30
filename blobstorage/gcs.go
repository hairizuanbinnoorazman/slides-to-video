package blobstorage

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type GCSStorage struct {
	Logger     logger.Logger
	Client     *storage.Client
	BucketName string
}

func NewGCSStorage(logger logger.Logger, client *storage.Client, bucketName string) GCSStorage {
	return GCSStorage{
		Logger:     logger,
		Client:     client,
		BucketName: bucketName,
	}
}

func (b GCSStorage) Save(ctx context.Context, fileName string, content []byte) error {
	writer := b.Client.Bucket(b.BucketName).Object(fileName).NewWriter(ctx)
	defer writer.Close()

	// Convert to bytes
	bufWriter := bufio.NewWriter(writer)
	_, err := bufWriter.Write(content)
	if err != nil {
		return fmt.Errorf("Unable to write content out to writer %v", err)
	}

	err = bufWriter.Flush()
	if err != nil {
		return fmt.Errorf("Unable to flush content out to GCS. %v", err)
	}
	return nil
}

func (b GCSStorage) Load(ctx context.Context, fileName string) (content []byte, err error) {
	reader, err := b.Client.Bucket(b.BucketName).Object(fileName).NewReader(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to retrieve file. Bucket Name: %v, File Name: %v, Error: %v", b.BucketName, fileName, err)
	}
	defer reader.Close()

	content, err = ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to write content out to writer %v", err)
	}

	return content, nil
}
