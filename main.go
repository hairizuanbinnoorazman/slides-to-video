package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/handlers"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/jobs"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/handlers"
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
var ProjectTableName = "test-Project"
var JobTableName = "test-Job"
var PDFToImageJobTableName = "test-PDFToImageJob"
var ImageToVideoJobTableName = "test-ImageToVideoJob"
var VideoConcatJobTableName = "test-VideoConcatJob"
var UserTableName = "test-User"

// Topics
var PDFToImageJobTopic = "pdf-splitter"
var ImageToVideoJobTopic = "image-to-video"
var VideoConcatJobTopic = "concatenate-video"

// Config is a reflection of the configuration of the values that needs to be set of the application
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
	RedirectURI  string `json:"redirect_uri"`
	Auth         h.Auth `json:"auth"`
}

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

	rawWebCredJSON, err := ioutil.ReadFile("config.json")
	if err != nil {
		logger.Error("Unable to load web application config")
	}
	var webCredJSON Config
	json.Unmarshal(rawWebCredJSON, &webCredJSON)

	projectStore := project.NewGoogleDatastore(datastoreClient, ProjectTableName)
	jobStore := jobs.NewGoogleDatastore(datastoreClient, JobTableName)
	userStore := user.NewGoogleDatastore(datastoreClient, UserTableName)
	slideToVideoStorage := blobstorage.NewGCSStorage(logger, xClient, BucketName)
	parentStore := jobs.NewGoogleDatastore(datastoreClient, ParentJobTableName)
	pdfToImageStore := jobs.NewGoogleDatastore(datastoreClient, PDFToImageJobTableName)
	pdfToImageQueue := queue.NewGooglePubsub(logger, pubsubClient, PDFToImageJobTopic)
	imageToVideoStore := jobs.NewGoogleDatastore(datastoreClient, ImageToVideoJobTableName)
	imageToVideoQueue := queue.NewGooglePubsub(logger, pubsubClient, ImageToVideoJobTopic)
	videoConcatStore := jobs.NewGoogleDatastore(datastoreClient, VideoConcatJobTableName)
	videoConcatQueue := queue.NewGooglePubsub(logger, pubsubClient, VideoConcatJobTopic)

	r := mux.NewRouter()
	r.Handle("/report_pdf_split", reportPDFSplit{
		Logger:            logger,
		ParentStore:       parentStore,
		PdfToImageStore:   pdfToImageStore,
		ImageToVideoStore: imageToVideoStore,
	})
	r.Handle("/report_image_to_video", reportImageToVideo{
		Logger:            logger,
		ImageToVideoStore: imageToVideoStore,
		VideoConcatStore:  videoConcatStore,
		VideoConcatQueue:  videoConcatQueue,
	})
	r.Handle("/report_video_concat", reportVideoConcat{
		Logger:           logger,
		ParentStore:      parentStore,
		VideoConcatStore: videoConcatStore,
	})

	s := r.PathPrefix("/api/v1").Subrouter()
	// Project based routes
	s.Handle("/project", h.CreateProject{
		Logger:           logger,
		Blobstorage:      slideToVideoStorage,
		PDFToImageQueue:  pdfToImageQueue,
		BucketFolderName: BucketFolder,
		ProjectStore:     projectStore,
		JobStore:         jobStore,
	}).Methods("POST")
	s.Handle("/projects", h.GetAllProjects{
		Logger:       logger,
		ProjectStore: projectStore,
	}).Methods("GET")
	s.Handle("/project/{project_id}", h.GetProject{
		Logger:       logger,
		ProjectStore: projectStore,
	}).Methods("GET")
	// Job based routes

	s.Handle("/job/{parent_job_id}:generate", startVideoGeneration{
		Logger:            logger,
		ParentStore:       parentStore,
		PdfToImageStore:   pdfToImageStore,
		ImageToVideoStore: imageToVideoStore,
		ImageToVideoQueue: imageToVideoQueue,
	}).Methods("POST")
	s.Handle("/video/{video_id}", downloadVideo{
		logger:        logger,
		storageClient: xClient,
		bucketName:    BucketName,
	}).Methods("GET")
	s.Handle("/image/{image_id}", downloadImage{
		logger:        logger,
		storageClient: xClient,
		bucketName:    BucketName,
	}).Methods("GET")
	s.Handle("/login", h.Login{
		Logger:      logger,
		ClientID:    webCredJSON.ClientID,
		RedirectURI: webCredJSON.RedirectURI,
		Scope:       webCredJSON.Scope,
	})
	s.Handle("/callback", h.Authenticate{
		Logger:       logger,
		TableName:    UserTableName,
		ClientID:     webCredJSON.ClientID,
		ClientSecret: webCredJSON.ClientSecret,
		RedirectURI:  webCredJSON.RedirectURI,
		Auth:         webCredJSON.Auth,
		UserStore:    userStore,
	})

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST"}),
	)

	srv := http.Server{
		Handler:      cors(r),
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(srv.ListenAndServe())
}
