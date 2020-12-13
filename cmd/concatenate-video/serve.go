package main

import (
	"fmt"
	"net/http"
	"os"

	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/concatenate-video/handlers"

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

				r := mux.NewRouter()
				r.Handle("/status", h.Status{
					Logger: logger,
				})

				srv := http.Server{
					Addr: fmt.Sprintf("%v:%v", cfg.Server.Host, cfg.Server.Port),
				}

				logger.Fatal(srv.ListenAndServe())
			},
		}
		return serverCmd
	}
)
