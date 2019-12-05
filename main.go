package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// ServiceName denotes name of service. In order to reduce confusion, try to make it similar to name on Cloud Run UI
var ServiceName = "slides-to-video-manager"

// Version denotes version no of service. Change it as necessary
var Version = "v0.1.0"

// ProjectID denotes Google Project where this is used on
var ProjectID = "expanded-league-162223"

var BucketName = "zontext-pdf-2-videos"
var BucketFolder = "pdf"

// ParentJob denotes which table would be used to save details of the job on top level
var ParentJobTableName = "test-ParentJob"
var PDFToImageJobTableName = "test-PDFToImageJob"
var ImageToVideoJobTableName = "test-ImageToVideoJob"
var VideoConcatJobTableName = "test-VideoConcatJob"

// Topics
var PDFToImageJobTopic = "pdf-splitter"
var ImageToVideoJobTopic = "image-to-video"
var VideoConcatJobTopic = "concatenate-video"

func main() {
	logger := logrus.New()
	logger.Formatter = stackdriver.NewFormatter(
		stackdriver.WithService(ServiceName),
		stackdriver.WithVersion(Version),
	)
	logger.Level = logrus.InfoLevel
	logger.Info("Application Start Up")
	defer logger.Info("Application Ended")

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "LOCAL"
	}
	logger.Infof("Application Mode: %v", mode)

	credJSON, err := ioutil.ReadFile("slides-to-video-manager.json")
	if err != nil {
		logger.Error("Unable to load slides-to-video-manager cred file")
	}
	xClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(credJSON))
	if err != nil {
		logger.Error("Unable to create storage client")
	}
	datastoreClient, err := datastore.NewClient(context.Background(), ProjectID, option.WithCredentialsJSON(credJSON))
	if err != nil {
		logger.Error("Unable to create pubsub client")
	}
	pubsubClient, err := pubsub.NewClient(context.Background(), ProjectID, option.WithCredentialsJSON(credJSON))
	if err != nil {
		logger.Error("Unable to create pubsub client")
	}

	r := mux.NewRouter()
	r.Handle("/upload", mainPage{logger: logger})
	r.Handle("/upload_complete", exampleHandler{
		logger:           logger,
		client:           xClient,
		datastoreClient:  datastoreClient,
		pubsubClient:     pubsubClient,
		bucketName:       BucketName,
		bucketFolderName: BucketFolder,
		parentTableName:  ParentJobTableName,
		tableName:        PDFToImageJobTableName,
		topicName:        PDFToImageJobTopic,
	})
	r.Handle("/report_pdf_split", reportPDFSplit{
		logger:          logger,
		datastoreClient: datastoreClient,
		pubsubClient:    pubsubClient,
		parentTableName: ParentJobTableName,
		tableName:       PDFToImageJobTableName,
		nextTableName:   ImageToVideoJobTableName,
		nextTopicName:   ImageToVideoJobTopic,
	})
	r.Handle("/report_image_to_video", reportImageToVideo{
		logger:          logger,
		datastoreClient: datastoreClient,
		pubsubClient:    pubsubClient,
		tableName:       ImageToVideoJobTableName,
		nextTableName:   VideoConcatJobTableName,
		nextTopicName:   VideoConcatJobTopic,
	})
	r.Handle("/report_video_concat", reportVideoConcat{
		logger:          logger,
		datastoreClient: datastoreClient,
		tableName:       VideoConcatJobTableName,
		parentTableName: ParentJobTableName,
	})
	r.Handle("/jobs", viewAllParentJobs{
		logger:          logger,
		datastoreClient: datastoreClient,
		tableName:       ParentJobTableName,
	})
	r.Handle("/download", downloadJob{
		logger:        logger,
		storageClient: xClient,
		bucketName:    BucketName,
	})

	s := r.PathPrefix("/api/v1").Subrouter()
	s.Handle("/jobs", viewAllParentJobsAPI{
		logger:          logger,
		datastoreClient: datastoreClient,
		tableName:       ParentJobTableName,
	})

	srv := http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(srv.ListenAndServe())
}
