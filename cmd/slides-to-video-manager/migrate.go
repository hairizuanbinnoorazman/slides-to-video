package main

import (
	"fmt"
	"os"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/pdfslideimages"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	migrateCmd = func() *cobra.Command {
		migrateCmd := &cobra.Command{
			Use:   "migrate",
			Short: "Runs database migration (if necessary)",
			Long:  `If one utilizes relational databases such as MySQL, Postgresql - that would require usage data schema migration to happen`,
			Run: func(cmd *cobra.Command, args []string) {
				logger := logrus.New()
				logger.Formatter = stackdriver.NewFormatter(
					stackdriver.WithService(serviceName),
					stackdriver.WithVersion(version),
				)
				logger.Level = logrus.InfoLevel
				logger.Info("Run migration")
				defer logger.Info("Migration completed")

				switch cfg.Datastore.Type {
				case "mysql":
					logger.Info("Run mysql migration")
					defer logger.Info("mysql migration complete")
					connectionString := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True",
						cfg.Datastore.MySQLConfig.User,
						cfg.Datastore.MySQLConfig.Password,
						cfg.Datastore.MySQLConfig.Host,
						cfg.Datastore.MySQLConfig.Port,
						cfg.Datastore.MySQLConfig.DBName,
					)
					db, err := gorm.Open("mysql", connectionString)
					if err != nil {
						logger.Errorf("Unable to connect to database. %v", err)
						os.Exit(1)
					}
					defer db.Close()
					db.AutoMigrate(&project.Project{})
					db.AutoMigrate(&videosegment.VideoSegment{})
					db.AutoMigrate(&pdfslideimages.PDFSlideImages{})
					db.AutoMigrate(&pdfslideimages.SlideAsset{})
					db.Model(&pdfslideimages.PDFSlideImages{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "RESTRICT")
					db.Model(&videosegment.VideoSegment{}).AddForeignKey("project_id", "projects(id)", "CASCADE", "RESTRICT")
					db.Model(&pdfslideimages.SlideAsset{}).AddForeignKey("pdf_slide_image_id", "pdf_slide_images(id)", "CASCADE", "RESTRICT")
					if db.Error != nil {
						logger.Errorf("unable to migrate project table. %v", db.Error)
					}
				default:
					logger.Errorf("Database defined is not meant to run migration")
					os.Exit(1)
				}

			},
		}
		migrateCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")
		return migrateCmd
	}
)
