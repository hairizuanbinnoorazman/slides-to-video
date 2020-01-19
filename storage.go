package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type BlobStorage struct {
	logger     logger.Logger
	client     *storage.Client
	bucketName string
}

func (b BlobStorage) Save(ctx context.Context, fileName string, content []byte) error {
	writer := b.client.Bucket(b.bucketName).Object(fileName).NewWriter(ctx)
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

func (b BlobStorage) Load(ctx context.Context, fileName string) (content []byte, err error) {
	reader, err := b.client.Bucket(b.bucketName).Object(fileName).NewReader(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to retrieve file. Bucket Name: %v, File Name: %v, Error: %v", b.bucketName, fileName, err)
	}
	defer reader.Close()

	content, err = ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, fmt.Errorf("Unable to write content out to writer %v", err)
	}

	return content, nil
}
