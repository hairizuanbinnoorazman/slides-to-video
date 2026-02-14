package blobstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
)

type S3Storage struct {
	Logger     logger.Logger
	Client     *s3.Client
	BucketName string
}

// NewS3Storage creates a new S3 storage client with the specified credentials mode.
// Supported credential modes:
// - "iam" (default): Use IAM role attached to instance
// - "static": Use provided accessKeyID and secretAccessKey
// - "shared": Use shared credentials file (~/.aws/credentials) with profile
// - "env": Use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables
func NewS3Storage(logger logger.Logger, region, bucket, credentialMode, accessKeyID, secretAccessKey, sharedCredFile, sharedCredProfile string) (S3Storage, error) {
	if region == "" {
		return S3Storage{}, fmt.Errorf("AWS region is required")
	}
	if bucket == "" {
		return S3Storage{}, fmt.Errorf("AWS S3 bucket name is required")
	}

	// Default to IAM role if no credential mode specified
	if credentialMode == "" {
		credentialMode = "iam"
	}

	var cfg aws.Config
	var err error

	switch credentialMode {
	case "iam":
		// Use default AWS config (IAM role from instance)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
		if err != nil {
			return S3Storage{}, fmt.Errorf("unable to load AWS config with IAM role: %v", err)
		}

	case "static":
		if accessKeyID == "" || secretAccessKey == "" {
			return S3Storage{}, fmt.Errorf("accessKeyID and secretAccessKey are required for static credential mode")
		}
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			)),
		)
		if err != nil {
			return S3Storage{}, fmt.Errorf("unable to load AWS config with static credentials: %v", err)
		}

	case "shared":
		// Use shared credentials file with optional profile
		configOpts := []func(*config.LoadOptions) error{
			config.WithRegion(region),
		}
		if sharedCredFile != "" {
			configOpts = append(configOpts, config.WithSharedCredentialsFiles([]string{sharedCredFile}))
		}
		if sharedCredProfile != "" {
			configOpts = append(configOpts, config.WithSharedConfigProfile(sharedCredProfile))
		}
		cfg, err = config.LoadDefaultConfig(context.Background(), configOpts...)
		if err != nil {
			return S3Storage{}, fmt.Errorf("unable to load AWS config with shared credentials: %v", err)
		}

	case "env":
		// Use environment variables (AWS SDK will automatically pick up AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
		if err != nil {
			return S3Storage{}, fmt.Errorf("unable to load AWS config from environment variables: %v", err)
		}

	default:
		return S3Storage{}, fmt.Errorf("unsupported credential mode: %s (supported: iam, static, shared, env)", credentialMode)
	}

	client := s3.NewFromConfig(cfg)

	logger.Infof("S3 storage initialized successfully. Region: %s, Bucket: %s, Credential Mode: %s", region, bucket, credentialMode)

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
