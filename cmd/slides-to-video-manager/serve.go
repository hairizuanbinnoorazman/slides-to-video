package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
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
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"

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
	serverCmd = func() *cobra.Command {
		serverCmd := &cobra.Command{
			Use:   "server",
			Short: "Run the API server of the slides to video manager tool",
			Long: `Runs the API server of the slides to video manager tool
	This tool forms the centerpiece of the whole integration.`,
			Run: func(cmd *cobra.Command, args []string) {
				logger := logrus.New()
				logger.Formatter = stackdriver.NewFormatter(
					stackdriver.WithService(serviceName),
					stackdriver.WithVersion(version),
				)
				logger.Level = logrus.InfoLevel
				logger.Info("Application Start Up")
				defer logger.Info("Application Ended")

				validate := validator.New()
				err := validate.Struct(cfg)
				if err != nil {
					logger.Errorf("Error with loading configuration. %v", err)
					os.Exit(1)
				}

				var svcAcctOptions []option.ClientOption
				if cfg.Server.SvcAcctFile != "" {
					credJSON, err := ioutil.ReadFile(cfg.Server.SvcAcctFile)
					if err != nil {
						logger.Errorf("Unable to load slides-to-video-manager cred file. err: %v", err)
					}
					svcAcctOptions = append(svcAcctOptions, option.WithCredentialsJSON(credJSON))
				}

				var slideToVideoStorage blobstorage.BlobStorage
				if cfg.BlobStorage.Type == gcsBlobStorage {
					var xClient *storage.Client
					xClient, err = storage.NewClient(context.Background(), svcAcctOptions...)
					if err != nil {
						logger.Errorf("Unable to create storage client %v", err)
						os.Exit(1)
					}
					slideToVideoStorage = blobstorage.NewGCSStorage(logger, xClient, cfg.BlobStorage.GCS.Bucket)
				} else if cfg.BlobStorage.Type == minioBlobStorage {
					slideToVideoStorage, err = blobstorage.NewMinio(logger, cfg.BlobStorage.Minio.Endpoint, cfg.BlobStorage.Minio.AccessKeyID, cfg.BlobStorage.Minio.SecretAccessKey, cfg.BlobStorage.Minio.Bucket)
					if err != nil {
						logger.Errorf("Unable to create storage client %v", err)
						os.Exit(1)
					}
				}

				if slideToVideoStorage == nil {
					logger.Errorf("Some of the storage instantiation is nil")
					os.Exit(1)
				}

				var projectStore project.Store
				var pdfSlideImagesStore pdfslideimages.Store
				var userStore user.Store
				var videoSegmentsStore videosegment.Store
				var aclStore acl.Store
				if cfg.Datastore.Type == googleDatastore {
					datastoreClient, err := datastore.NewClient(context.Background(), cfg.Datastore.GoogleDatastoreConfig.ProjectID, svcAcctOptions...)
					if err != nil {
						logger.Errorf("Unable to create datastore client. %v", err)
						os.Exit(1)
					}
					projectStore = project.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)
					pdfSlideImagesStore = pdfslideimages.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName)
					userStore = user.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.UserTableName)
					videoSegmentsStore = videosegment.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)
				} else if cfg.Datastore.Type == mysqlDatastore {
					connectionString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True", cfg.Datastore.MySQLConfig.User, cfg.Datastore.MySQLConfig.Password, cfg.Datastore.MySQLConfig.Host, cfg.Datastore.MySQLConfig.Port, cfg.Datastore.MySQLConfig.DBName)
					db, err := gorm.Open("mysql", connectionString)
					if err != nil {
						logger.Errorf("Unable to create mysql client. %v", err)
						os.Exit(1)
					}
					projectStore = project.NewMySQL(logger, db)
					pdfSlideImagesStore = pdfslideimages.NewMySQL(logger, db)
					userStore = user.NewMySQL(logger, db)
					videoSegmentsStore = videosegment.NewMySQL(logger, db)
					aclStore = acl.NewMySQL(logger, db)
				}

				if projectStore == nil || pdfSlideImagesStore == nil || userStore == nil || videoSegmentsStore == nil {
					logger.Errorf("Some of the database instantiation is nil")
					os.Exit(1)
				}

				var pdfToImageQueue queue.Queue
				var imageToVideoQueue queue.Queue
				var concatQueue queue.Queue
				if cfg.Queue.Type == googlePubsubQueue {
					pubsubClient, err := pubsub.NewClient(context.Background(), cfg.Queue.GooglePubsub.ProjectID, svcAcctOptions...)
					if err != nil {
						logger.Errorf("Unable to create pubsub client. %v", err)
						os.Exit(1)
					}

					pdfToImageQueue = queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.PDFToImageTopic)
					imageToVideoQueue = queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.ImageToVideoTopic)
					concatQueue = queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.GooglePubsub.VideoConcatTopic)
				} else if cfg.Queue.Type == natsQueue {
					pdfToImageQueue, err = queue.NewNats(logger, cfg.Queue.NatsConfig.Endpoint, cfg.Queue.NatsConfig.PDFToImageTopic)
					if err != nil {
						logger.Errorf("Unable to create Nats client. %v", err)
					}
					imageToVideoQueue, err = queue.NewNats(logger, cfg.Queue.NatsConfig.Endpoint, cfg.Queue.NatsConfig.ImageToVideoTopic)
					if err != nil {
						logger.Errorf("Unable to create Nats client. %v", err)
					}
					concatQueue, err = queue.NewNats(logger, cfg.Queue.NatsConfig.Endpoint, cfg.Queue.NatsConfig.VideoConcatTopic)
					if err != nil {
						logger.Errorf("Unable to create Nats client. %v", err)
					}
				}

				if pdfToImageQueue == nil || imageToVideoQueue == nil || concatQueue == nil {
					logger.Errorf("Some of the queue instatiation is nil")
					os.Exit(1)
				}

				pdfSlideImporter := imageimporter.NewBasicPDFImporter(pdfToImageQueue)
				videoGenerator := videogenerator.NewBasic(imageToVideoQueue, videoSegmentsStore)
				videoConcater := videoconcater.NewBasic(concatQueue, projectStore)

				r := mux.NewRouter()
				r.Handle("/status", h.Status{
					Logger: logger,
				})
				r.Handle("/healthz", h.Status{
					Logger: logger,
				})
				r.Handle("/readyz", h.Status{
					Logger: logger,
				})

				auth := h.Auth{
					Secret:     cfg.Server.AuthSecret,
					Issuer:     cfg.Server.AuthIssuer,
					ExpiryTime: cfg.Server.AuthExpiryTime,
				}
				s := r.PathPrefix("/api/v1").Subrouter()
				// Project based routes
				s.Handle("/project", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.CreateProject{
						Logger:       logger,
						ProjectStore: projectStore,
						ACLStore:     aclStore,
					},
				}).Methods("POST")
				s.Handle("/projects", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.GetAllProjects{
						Logger:       logger,
						ProjectStore: projectStore,
					},
				}).Methods("GET")
				s.Handle("/project/{project_id}", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.GetProject{
						Logger:       logger,
						ProjectStore: projectStore,
					},
				}).Methods("GET")
				s.Handle("/project/{project_id}", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.UpdateProject{
						Logger:       logger,
						ProjectStore: projectStore,
					},
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

				// User based endpoints
				s.Handle("/user/{user_id}", h.GetUser{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("GET")
				s.Handle("/users/register", h.CreateUser{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("POST")
				s.Handle("/users/activate", h.ActivateUser{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("GET")
				s.Handle("/users/forgetpassword", h.ForgetPassword{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("POST")
				s.Handle("/users/resetpassword", h.ForgetPassword{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("POST")
				s.Handle("/login", h.Login{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("POST")
				s.Handle("/connect/google", h.GoogleLogin{
					Logger:      logger,
					ClientID:    cfg.Server.ClientID,
					RedirectURI: cfg.Server.RedirectURI,
					Scope:       cfg.Server.Scope,
				})
				s.Handle("/callback/google", h.Authenticate{
					Logger:       logger,
					TableName:    cfg.Datastore.GoogleDatastoreConfig.UserTableName,
					ClientID:     cfg.Server.ClientID,
					ClientSecret: cfg.Server.ClientSecret,
					RedirectURI:  cfg.Server.RedirectURI,
					Auth:         auth,
					UserStore:    userStore,
				})

				cors := handlers.CORS(
					handlers.AllowedHeaders([]string{"content-type", "Authorization"}),
					handlers.AllowedOrigins([]string{"*"}),
					handlers.AllowedMethods([]string{"GET", "POST", "PUT"}),
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
		serverCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")
		return serverCmd
	}
)
