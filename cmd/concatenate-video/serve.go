package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/handlers"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/mgrclient"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/queuehandler"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/videoconcater"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/queue"
	"google.golang.org/api/option"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/validator.v9"
)

var (
	serverCmd = func() *cobra.Command {
		serverCmd := &cobra.Command{
			Use:   "server",
			Short: "Run the image to video component of the slides to video manager tool",
			Long: `Runs the image to video component of the slides to video manager tool
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

				var credJSON []byte
				var svcAcctOptions []option.ClientOption
				if cfg.Server.SvcAcctFile != "" {
					credJSON, err = ioutil.ReadFile(cfg.Server.SvcAcctFile)
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

				mgrURL := fmt.Sprintf("http://%v:%v/api/v1", cfg.Server.ManagerHost, cfg.Server.ManagerPort)
				if cfg.Server.ManagerPort == 443 {
					mgrURL = fmt.Sprintf("https://%v/api/v1", cfg.Server.ManagerHost)
				}

				mgrClient := mgrclient.NewBasic(logger, mgrURL, http.DefaultClient)
				videoConcater := videoconcater.NewBasic(logger, slideToVideoStorage, mgrClient, cfg.BlobStorage.VideoSnippetsFolder, cfg.BlobStorage.VideoFolder)

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

				if cfg.Server.Mode == "http" {
					r.Handle(cfg.Server.ProcessRoute, h.ProcessHandler{
						Logger:        logger,
						VideoConcater: &videoConcater,
					})
				}

				if cfg.Server.Mode == "queue" {
					var imageToVideoQueue queue.Queue
					if cfg.Queue.Type == googlePubsubQueue {
						pubsubClient, err := pubsub.NewClient(context.Background(), cfg.Queue.GooglePubsub.ProjectID, svcAcctOptions...)
						if err != nil {
							logger.Errorf("Unable to create pubsub client. %v", err)
							os.Exit(1)
						}

						imageToVideoQueue = queue.NewGooglePubsub(logger, pubsubClient, cfg.Queue.ConcatenateVideoTopic)
					} else if cfg.Queue.Type == natsQueue {
						imageToVideoQueue, err = queue.NewNats(logger, cfg.Queue.NatsConfig.Endpoint, cfg.Queue.ConcatenateVideoTopic)
						if err != nil {
							logger.Errorf("Unable to create Nats client. %v", err)
						}
					}

					if imageToVideoQueue == nil {
						logger.Errorf("Some of the queue instantiation is nil")
						os.Exit(1)
					}

					queueHandler := queuehandler.NewBasic(logger, imageToVideoQueue, &videoConcater)
					go queueHandler.HandleMessages()
				}

				srv := http.Server{
					Handler: r,
					Addr:    fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
				}

				logger.Fatal(srv.ListenAndServe())
			},
		}
		serverCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")
		return serverCmd
	}
)
