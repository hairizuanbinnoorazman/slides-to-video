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
	concatmgrclient "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/mgrclient"
	concatworker "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/videoconcater"
	img2vidconverter "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/image-to-video/image2videoconverter"
	img2vidmgrclient "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/image-to-video/mgrclient"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/pdfsplitter"
	pdfmgrclient "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/pdf-splitter/mgrclient"
	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/slides-to-video-manager/handlers"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/slides-to-video-manager/workers"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/imageimporter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/job"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videogenerator"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
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
				validate.RegisterStructValidation(ConfigStructLevelValidation, config{})
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
				} else if cfg.BlobStorage.Type == localBlobStorage {
					slideToVideoStorage, err = blobstorage.NewLocalStorage(logger, cfg.BlobStorage.Local.BasePath)
					if err != nil {
						logger.Errorf("Unable to create local storage client %v", err)
						os.Exit(1)
					}
				} else if cfg.BlobStorage.Type == s3BlobStorage {
					slideToVideoStorage, err = blobstorage.NewS3Storage(
						logger,
						cfg.BlobStorage.S3.Region,
						cfg.BlobStorage.S3.Bucket,
						cfg.BlobStorage.S3.CredentialMode,
						cfg.BlobStorage.S3.AccessKeyID,
						cfg.BlobStorage.S3.SecretAccessKey,
						cfg.BlobStorage.S3.SharedCredFile,
						cfg.BlobStorage.S3.SharedCredProfile,
					)
					if err != nil {
						logger.Errorf("Unable to create S3 storage client %v", err)
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
				var jobStore job.Store
				if cfg.Datastore.Type == googleDatastore {
					datastoreClient, err := datastore.NewClient(context.Background(), cfg.Datastore.GoogleDatastoreConfig.ProjectID, svcAcctOptions...)
					if err != nil {
						logger.Errorf("Unable to create datastore client. %v", err)
						os.Exit(1)
					}
					projectStore = project.NewGoogleDatastore(logger, datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)
					pdfSlideImagesStore = pdfslideimages.NewGoogleDatastore(logger, datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.PDFSlidesTableName)
					userStore = user.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.UserTableName)
					videoSegmentsStore = videosegment.NewGoogleDatastore(datastoreClient, cfg.Datastore.GoogleDatastoreConfig.ProjectTableName, cfg.Datastore.GoogleDatastoreConfig.VideoSegmentsTableName)
					aclStore, _ = acl.NewGoogleDatastore(logger, datastoreClient, "acl")
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
					jobStore = job.NewMySQL(logger, db)
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
				} else if cfg.Queue.Type == channelsQueue {
					bufferSize := cfg.Queue.Channels.BufferSize
					if bufferSize == 0 {
						bufferSize = 100
					}
					pdfToImageQueue = queue.NewChannels(logger, cfg.Queue.Channels.PDFToImageTopic, bufferSize)
					imageToVideoQueue = queue.NewChannels(logger, cfg.Queue.Channels.ImageToVideoTopic, bufferSize)
					concatQueue = queue.NewChannels(logger, cfg.Queue.Channels.VideoConcatTopic, bufferSize)
				}

				if pdfToImageQueue == nil || imageToVideoQueue == nil || concatQueue == nil {
					logger.Errorf("Some of the queue instatiation is nil")
					os.Exit(1)
				}

				auth := services.Auth{
					Secret:     cfg.Server.AuthSecret,
					Issuer:     cfg.Server.AuthIssuer,
					ExpiryTime: cfg.Server.AuthExpiryTime,
					HashKey:    securecookie.GenerateRandomKey(64),
					BlockKey:   securecookie.GenerateRandomKey(32),
					CookieName: "slidestovideo",
				}

				pdfSlideImporter := imageimporter.NewBasicPDFImporter(pdfToImageQueue)
				videoGenerator := videogenerator.NewBasic(imageToVideoQueue, videoSegmentsStore)
				videoConcater := videoconcater.NewBasic(concatQueue, projectStore, auth)

				jobProcessor, err := job.NewProcessor(logger, jobStore, projectStore, videoConcater)
				if err != nil {
					logger.Errorf("Unable to start job processor. Err - %v", err)
					os.Exit(1)
				}
				go jobProcessor.Start()

				// Initialize and start workers if enabled
				var workerList []workers.Worker
				workersCtx, cancelWorkers := context.WithCancel(context.Background())
				defer cancelWorkers()

				if cfg.Workers.Enabled {
					logger.Info("Workers enabled, initializing worker processors")

					// Manager URL for embedded workers to call back
					scheme := cfg.Server.Scheme
					if scheme == "" {
						scheme = "http"
					}
					var mgrURL string
					if (scheme == "http" && cfg.Server.Port == 80) || (scheme == "https" && cfg.Server.Port == 443) {
						// Omit port for standard ports
						mgrURL = fmt.Sprintf("%s://%v/api/v1", scheme, cfg.Server.Host)
					} else {
						mgrURL = fmt.Sprintf("%s://%v:%v/api/v1", scheme, cfg.Server.Host, cfg.Server.Port)
					}

					// PDF Splitter Worker
					if cfg.Workers.PDFSplitter.Enabled {
						logger.Info("Initializing PDF splitter workers")
						pdfMgrClient := pdfmgrclient.NewBasic(logger, mgrURL, http.DefaultClient)

						var pdfFolder, imagesFolder string
						if cfg.BlobStorage.Type == gcsBlobStorage {
							pdfFolder = cfg.BlobStorage.GCS.PDFFolder
							imagesFolder = cfg.BlobStorage.GCS.ImagesFolder
						} else if cfg.BlobStorage.Type == minioBlobStorage {
							pdfFolder = cfg.BlobStorage.Minio.PDFFolder
							imagesFolder = cfg.BlobStorage.Minio.ImagesFolder
						} else if cfg.BlobStorage.Type == localBlobStorage {
							pdfFolder = cfg.BlobStorage.Local.PDFFolder
							imagesFolder = cfg.BlobStorage.Local.ImagesFolder
						} else if cfg.BlobStorage.Type == s3BlobStorage {
							pdfFolder = cfg.BlobStorage.S3.PDFFolder
							imagesFolder = cfg.BlobStorage.S3.ImagesFolder
						}

						pdfProcessor := pdfsplitter.NewBasic(logger, slideToVideoStorage, pdfMgrClient, pdfFolder, imagesFolder)

						concurrency := cfg.Workers.PDFSplitter.Concurrency
						if concurrency == 0 {
							concurrency = 1
						}
						for i := 0; i < concurrency; i++ {
							worker := workers.NewPDFSplitterWorker(logger, pdfToImageQueue, &pdfProcessor)
							workerList = append(workerList, worker)
						}
						logger.Infof("Created %d PDF splitter worker(s)", concurrency)
					}

					// Image-to-Video Worker
					if cfg.Workers.ImageToVideo.Enabled {
						logger.Info("Initializing image-to-video workers")
						img2vidMgrClient := img2vidmgrclient.NewBasic(logger, mgrURL, http.DefaultClient)

						// Initialize text-to-speech client
						text2speechClient, err := texttospeech.NewClient(context.Background(), svcAcctOptions...)
						if err != nil {
							logger.Errorf("Unable to create text to speech client: %v", err)
							os.Exit(1)
						}
						defer text2speechClient.Close()

						var imagesFolder, videoSnippetsFolder string
						if cfg.BlobStorage.Type == gcsBlobStorage {
							imagesFolder = cfg.BlobStorage.GCS.ImagesFolder
							videoSnippetsFolder = cfg.BlobStorage.GCS.VideoSnippetsFolder
						} else if cfg.BlobStorage.Type == minioBlobStorage {
							imagesFolder = cfg.BlobStorage.Minio.ImagesFolder
							videoSnippetsFolder = cfg.BlobStorage.Minio.VideoSnippetsFolder
						} else if cfg.BlobStorage.Type == localBlobStorage {
							imagesFolder = cfg.BlobStorage.Local.ImagesFolder
							videoSnippetsFolder = cfg.BlobStorage.Local.VideoSnippetsFolder
						} else if cfg.BlobStorage.Type == s3BlobStorage {
							imagesFolder = cfg.BlobStorage.S3.ImagesFolder
							videoSnippetsFolder = cfg.BlobStorage.S3.VideoSnippetsFolder
						}

						textToSpeechEngine := img2vidconverter.NewGoogleTextToSpeech(logger, text2speechClient)
						img2vidProcessor := img2vidconverter.NewBasic(logger, slideToVideoStorage, img2vidMgrClient, imagesFolder, videoSnippetsFolder, &textToSpeechEngine)

						concurrency := cfg.Workers.ImageToVideo.Concurrency
						if concurrency == 0 {
							concurrency = 1
						}
						for i := 0; i < concurrency; i++ {
							worker := workers.NewImage2VideoWorker(logger, imageToVideoQueue, &img2vidProcessor)
							workerList = append(workerList, worker)
						}
						logger.Infof("Created %d image-to-video worker(s)", concurrency)
					}

					// Concatenate Video Worker
					if cfg.Workers.ConcatenateVideo.Enabled {
						logger.Info("Initializing concatenate-video workers")
						concatMgrClient := concatmgrclient.NewBasic(logger, mgrURL, http.DefaultClient)

						var videoSnippetsFolder, videoFolder string
						if cfg.BlobStorage.Type == gcsBlobStorage {
							videoSnippetsFolder = cfg.BlobStorage.GCS.VideoSnippetsFolder
							videoFolder = cfg.BlobStorage.GCS.VideoFolder
						} else if cfg.BlobStorage.Type == minioBlobStorage {
							videoSnippetsFolder = cfg.BlobStorage.Minio.VideoSnippetsFolder
							videoFolder = cfg.BlobStorage.Minio.VideoFolder
						} else if cfg.BlobStorage.Type == localBlobStorage {
							videoSnippetsFolder = cfg.BlobStorage.Local.VideoSnippetsFolder
							videoFolder = cfg.BlobStorage.Local.VideoFolder
						} else if cfg.BlobStorage.Type == s3BlobStorage {
							videoSnippetsFolder = cfg.BlobStorage.S3.VideoSnippetsFolder
							videoFolder = cfg.BlobStorage.S3.VideoFolder
						}

						concatProcessor := concatworker.NewBasic(logger, slideToVideoStorage, concatMgrClient, videoSnippetsFolder, videoFolder)

						concurrency := cfg.Workers.ConcatenateVideo.Concurrency
						if concurrency == 0 {
							concurrency = 1
						}
						for i := 0; i < concurrency; i++ {
							worker := workers.NewVideoConcaterWorker(logger, concatQueue, &concatProcessor)
							workerList = append(workerList, worker)
						}
						logger.Infof("Created %d concatenate-video worker(s)", concurrency)
					}

					// Start all workers in goroutines
					for idx, w := range workerList {
						go func(worker workers.Worker, workerIdx int) {
							logger.Infof("Starting worker #%d", workerIdx)
							if err := worker.Start(workersCtx); err != nil && err != context.Canceled {
								logger.Errorf("Worker #%d stopped with error: %v", workerIdx, err)
							}
						}(w, idx)
					}
					logger.Infof("Started %d total worker(s)", len(workerList))
				}

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
						ACLStore:     aclStore,
					},
				}).Methods("GET")
				s.Handle("/project/{project_id}", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.UpdateProject{
						Logger:       logger,
						ProjectStore: projectStore,
						ACLStore:     aclStore,
					},
				}).Methods("PUT")
				s.Handle("/project/{project_id}:concat", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.StartVideoConcat{
						Logger:        logger,
						ProjectStore:  projectStore,
						VideoConcater: videoConcater,
						ACLStore:      aclStore,
					},
				}).Methods("POST")
				s.Handle("/project/{project_id}:generate-video", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.StartProjectGenerateVideo{
						Logger:             logger,
						JobStore:           jobStore,
						ProjectStore:       projectStore,
						ACLStore:           aclStore,
						VideoSegmentsStore: videoSegmentsStore,
						VideoGenerator:     videoGenerator,
					},
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
				s.Handle("/project/{project_id}/video/{video_id}", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.DownloadVideo{
						Logger:        logger,
						StorageClient: slideToVideoStorage,
					},
				}).Methods("GET")
				s.Handle("/project/{project_id}/image/{image_id}", h.RequireJWTAuth{
					Auth:   auth,
					Logger: logger,
					NextHandler: h.DownloadImage{
						Logger:        logger,
						StorageClient: slideToVideoStorage,
					},
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
				s.Handle("/users/resetpassword", h.ResetPassword{
					Logger:    logger,
					UserStore: userStore,
				}).Methods("POST")
				s.Handle("/login", h.Login{
					Logger:    logger,
					UserStore: userStore,
					Auth:      auth,
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

				// cors := handlers.CORS(
				// 	handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Set-Cookie"}),
				// 	handlers.AllowedOrigins([]string{"http://localhost:8080"}),
				// 	handlers.AllowedMethods([]string{"GET", "POST", "PUT", "OPTIONS"}),
				// 	handlers.AllowCredentials(),
				// )

				srv := http.Server{
					Handler:      r,
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
