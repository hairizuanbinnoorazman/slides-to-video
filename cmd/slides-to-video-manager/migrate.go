package main

import (
	"fmt"
	"os"

	stackdriver "github.com/TV4/logrus-stackdriver-formatter"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	migrateCmd = &cobra.Command{
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
				connectionString := fmt.Sprintf("%v:%v@%v:%v/%v?charset=utf8mb4&parseTime=True",
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
				if db.Error != nil {
					logger.Errorf("unable to migrate project table. %v", db.Error)
				}
			default:
				logger.Errorf("Database defined is not meant to run migration")
				os.Exit(1)
			}

		},
	}
)
