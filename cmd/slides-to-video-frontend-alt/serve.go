package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	h "github.com/hairizuanbinnoorazman/slides-to-video-manager/cmd/slides-to-video-frontend-alt/handlers"
	"gopkg.in/go-playground/validator.v9"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

				cors := handlers.CORS(
					handlers.AllowedHeaders([]string{"content-type"}),
					handlers.AllowedOrigins([]string{"*"}),
					handlers.AllowedMethods([]string{"GET", "POST"}),
				)

				frontendScheme := "http"
				if cfg.Secure {
					frontendScheme = "https"
				}

				s := r.PathPrefix(cfg.IngressPath).Subrouter()
				s.Handle("/", h.RequireLogin{
					Scheme:      frontendScheme,
					IngressPath: cfg.IngressPath,
					Logger:      logger,
					NextHandler: h.Home{
						Logger: logger,
					},
				})
				s.Handle("/login", h.Login{
					IngressPath: cfg.IngressPath,
					Logger:      logger,
					MgrEndpoint: cfg.ServerEndpoint,
				})
				s.Handle("/projects", h.RequireLogin{
					Scheme:      frontendScheme,
					IngressPath: cfg.IngressPath,
					Logger:      logger,
					NextHandler: h.Projects{
						Logger: logger,
					},
				})

				srv := http.Server{
					Handler:      cors(r),
					Addr:         fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
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
