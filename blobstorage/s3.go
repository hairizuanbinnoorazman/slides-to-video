package blobstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type S3Storage struct {
	Logger     logger.Logger
	Client     *s3.Client
	BucketName string
}

// NewS3Storage creates a new S3 storage client using IAM role credentials.
// This implementation relies on the AWS SDK's default credential chain, which will
// automatically use IAM roles from EC2 instance metadata, ECS task roles, or EKS pod roles.
// No static credentials are supported to follow AWS security best practices.
func NewS3Storage(logger logger.Logger, region, bucket string) (S3Storage, error) {
	if region == "" {
		return S3Storage{}, fmt.Errorf("AWS region is required")
	}
	if bucket == "" {
		return S3Storage{}, fmt.Errorf("AWS S3 bucket name is required")
	}

	// Load AWS config using default credential chain (IAM roles)
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return S3Storage{}, fmt.Errorf("unable to load AWS config with IAM role: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	logger.Infof("S3 storage initialized successfully. Region: %s, Bucket: %s (using IAM role credentials)", region, bucket)

	return S3Storage{
		Logger:     logger,
		Client:     client,
		BucketName: bucket,
	}, nil
}

func (s S3Storage) Save(ctx context.Context, fileName string, content []byte) error {
	if s.Client == nil {
		return fmt.Errorf("S3 client not initialized")
	}

	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(content),
	})
	if err != nil {
		return fmt.Errorf("unable to save file to S3. Bucket: %s, Key: %s, Error: %v", s.BucketName, fileName, err)
	}

	s.Logger.Infof("Successfully saved file to S3. Bucket: %s, Key: %s", s.BucketName, fileName)
	return nil
}

func (s S3Storage) Load(ctx context.Context, fileName string) (content []byte, err error) {
	if s.Client == nil {
		return []byte{}, fmt.Errorf("S3 client not initialized")
	}

	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return []byte{}, fmt.Errorf("unable to retrieve file from S3. Bucket: %s, Key: %s, Error: %v", s.BucketName, fileName, err)
	}
	defer result.Body.Close()

	content, err = io.ReadAll(result.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to read file content from S3. Bucket: %s, Key: %s, Error: %v", s.BucketName, fileName, err)
	}

	s.Logger.Infof("Successfully loaded file from S3. Bucket: %s, Key: %s", s.BucketName, fileName)
	return content, nil
}
