package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/handlers"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/imageimporter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videogenerator"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Print the version number of slides-to-video-manager",
		Long:  `Print the version number of slides-to-video-manager`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logrus.New()
			logger.Formatter = stackdriver.NewFormatter(
				stackdriver.WithService(serviceName),
				stackdriver.WithVersion(version),
			)
			logger.Level = logrus.InfoLevel
			logger.Info("Application Start Up")
			defer logger.Info("Application Ended")

			credJSON, err := ioutil.ReadFile("slides-to-video-manager.json")
			if err != nil {
				logger.Error("Unable to load slides-to-video-manager cred file")
			}
			xClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(credJSON))
			if err != nil {
				logger.Error("Unable to create storage client")
			}
			datastoreClient, err := datastore.NewClient(context.Background(), cfg.Datastore.GoogleDatastoreConfig.ProjectID, option.WithCredentialsJSON(credJSON))
			if err != nil {
				logger.Error("Unable to create pubsub client")
			}
			pubsubClient, err := pubsub.NewClient(context.Background(), cfg.Datastore.GoogleDatastoreConfig.ProjectID, option.WithCredentialsJSON(credJSON))
			if err != nil {
				logger.Error("Unable to create pubsub client")
			}

			projectStore := project.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)
			pdfSlideImagesStore := pdfslideimages.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName)
			userStore := user.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.UserTableName)
			videoSegmentsStore := videosegment.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)

			slideToVideoStorage := blobstorage.NewGCSStorage(logger, xClient, cfg.BlobStorage.GCS.Bucket)
			pdfToImageQueue := queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.PDFToImageJobTopic)
			imageToVideoQueue := queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.ImageToVideoJobTopic)
			concatQueue := queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.VideoConcatJobTopic)

			pdfSlideImporter := imageimporter.NewBasicPDFImporter(pdfToImageQueue)
			videoGenerator := videogenerator.NewBasic(imageToVideoQueue, videoSegmentsStore)
			videoConcater := videoconcater.NewBasic(concatQueue, projectStore)

			r := mux.NewRouter()

			s := r.PathPrefix("/api/v1").Subrouter()
			// Project based routes
			s.Handle("/project", h.CreateProject{
				Logger:       logger,
				ProjectStore: projectStore,
			}).Methods("POST")
			s.Handle("/projects", h.GetAllProjects{
				Logger:       logger,
				ProjectStore: projectStore,
			}).Methods("GET")
			s.Handle("/project/{project_id}", h.GetProject{
				Logger:       logger,
				ProjectStore: projectStore,
			}).Methods("GET")
			s.Handle("/project/{project_id}", h.UpdateProject{
				Logger:       logger,
				ProjectStore: projectStore,
			}).Methods("PUT")
			s.Handle("/project/{project_id}:concat", h.StartVideoConcat{
				Logger:        logger,
				ProjectStore:  projectStore,
				VideoConcater: videoConcater,
			}).Methods("POST")
			s.Handle("/project/{project_id}/pdfslideimages", h.CreatePDFSlideImages{
				Logger:              logger,
				PDFSlideImagesStore: pdfSlideImagesStore,
				Blobstorage:         slideToVideoStorage,
				BucketFolderName:    cfg.BlobStorage.GCS.PDFFolder,
				PDFSlideImporter:    pdfSlideImporter,
			}).Methods("POST")
			s.Handle("/project/{project_id}/pdfslideimages/{pdfslideimages_id}", h.UpdatePDFSlideImages{
				Logger:              logger,
				PDFSlideImagesStore: pdfSlideImagesStore,
				VideoSegmentStore:   videoSegmentsStore,
			}).Methods("PUT")
			s.Handle("/project/{project_id}/pdfslideimages/{pdfslideimages_id}", h.GetPDFSlideImages{
				Logger:              logger,
				PDFSlideImagesStore: pdfSlideImagesStore,
			}).Methods("GET")
			s.Handle("/project/{project_id}/videosegment", h.CreateVideoSegment{
				Logger:            logger,
				VideoSegmentStore: videoSegmentsStore,
			}).Methods("POST")
			s.Handle("/project/{project_id}/videosegment/{videosegment_id}", h.UpdateVideoSegment{
				Logger:            logger,
				VideoSegmentStore: videoSegmentsStore,
			}).Methods("PUT")
			s.Handle("/project/{project_id}/videosegment/{videosegment_id}", h.GetVideoSegment{
				Logger:            logger,
				VideoSegmentStore: videoSegmentsStore,
			}).Methods("GET")
			s.Handle("/project/{project_id}/videosegment/{videosegment_id}:generate", h.StartVideoSegmentGeneration{
				Logger:            logger,
				VideoSegmentStore: videoSegmentsStore,
				VideoGenerator:    videoGenerator,
			}).Methods("POST")
			// Asset retriver routes
			s.Handle("/project/{project_id}/video/{video_id}", h.DownloadVideo{
				Logger:        logger,
				StorageClient: slideToVideoStorage,
			}).Methods("GET")
			s.Handle("/project/{project_id}/image/{image_id}", h.DownloadImage{
				Logger:        logger,
				StorageClient: slideToVideoStorage,
			}).Methods("GET")
			s.Handle("/login", h.Login{
				Logger:      logger,
				ClientID:    cfg.Server.ClientID,
				RedirectURI: cfg.Server.RedirectURI,
				Scope:       cfg.Server.Scope,
			})
			auth := h.Auth{
				Secret:     cfg.Server.AuthSecret,
				Issuer:     cfg.Server.AuthIssuer,
				ExpiryTime: cfg.Server.AuthExpiryTime,
			}
			s.Handle("/callback", h.Authenticate{
				Logger:       logger,
				TableName:    cfg.Datastore.GoogleDatastoreConfig.UserTableName,
				ClientID:     cfg.Server.ClientID,
				ClientSecret: cfg.Server.ClientSecret,
				RedirectURI:  cfg.Server.RedirectURI,
				Auth:         auth,
				UserStore:    userStore,
			})

			cors := handlers.CORS(
				handlers.AllowedHeaders([]string{"content-type"}),
				handlers.AllowedOrigins([]string{"*"}),
				handlers.AllowedMethods([]string{"GET", "POST"}),
			)

			srv := http.Server{
				Handler:      cors(r),
				Addr:         fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}

			logger.Fatal(srv.ListenAndServe())
		},
	}
)
