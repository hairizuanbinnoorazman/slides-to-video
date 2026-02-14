package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cfgFile string

	// Includes default configuration
	// Default configuration is set for all-in-one mode with embedded workers
	// Uses MySQL for datastore, Minio for storage, and channels queue for embedded workers
	// Immediately replaces value with environment variables on startup
	cfg = config{
		Server: serverConfig{
			Host:           envVarOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:           envVarOrDefaultInt("SERVER_PORT", 8080),
			Scope:          "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/drive.metadata.readonly",
			SvcAcctFile:    envVarOrDefault("SERVER_SVCACCTFILE", ""),
			ClientID:       envVarOrDefault("SERVER_CLIENTID", ""),
			ClientSecret:   envVarOrDefault("SERVER_CLIENTSECRET", ""),
			RedirectURI:    envVarOrDefault("SERVER_REDIRECTURI", "http://localhost:8000/api/v1/callback"),
			AuthSecret:     envVarOrDefault("SERVER_AUTHSECRET", "secret"),
			AuthIssuer:     envVarOrDefault("SERVER_AUTHISSUER", "issuer"),
			AuthExpiryTime: envVarOrDefaultInt("SERVER_AUTHEXPIRYTIME", 3600),
		},
		Workers: workersConfig{
			Enabled: envVarOrDefaultBool("WORKERS_ENABLED", true),
			PDFSplitter: workerInstanceConfig{
				Enabled:     envVarOrDefaultBool("WORKERS_PDFSPLITTER_ENABLED", true),
				Concurrency: envVarOrDefaultInt("WORKERS_PDFSPLITTER_CONCURRENCY", 1),
			},
			ImageToVideo: workerInstanceConfig{
				Enabled:     envVarOrDefaultBool("WORKERS_IMAGETOVIDEO_ENABLED", true),
				Concurrency: envVarOrDefaultInt("WORKERS_IMAGETOVIDEO_CONCURRENCY", 2),
			},
			ConcatenateVideo: workerInstanceConfig{
				Enabled:     envVarOrDefaultBool("WORKERS_CONCATENATEVIDEO_ENABLED", true),
				Concurrency: envVarOrDefaultInt("WORKERS_CONCATENATEVIDEO_CONCURRENCY", 1),
			},
		},
		Datastore: datastoreConfig{
			Type: envVarOrDefault("DATASTORE_TYPE", "mysql"),
			GoogleDatastoreConfig: &googleDatastoreConfig{
				ProjectID:              envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PROJECTID", ""),
				UserTableName:          envVarOrDefault("DATASTORE_GOOGLEDATASTORE_USERTABLENAME", "UserTable"),
				ProjectTableName:       envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PROJECTTABLENAME", "ProjectTable"),
				PDFSlidesTableName:     envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PDFSLIDESTABLENAME", "PDFSlideTable"),
				VideoSegmentsTableName: envVarOrDefault("DATASTORE_GOOGLEDATASTORE_VIDEOSEGMENTSTABLENAME", "VideoSegmentsTable"),
			},
			MySQLConfig: &mysqlConfig{
				User:     envVarOrDefault("DATASTORE_MYSQL_USER", "user"),
				Password: envVarOrDefault("DATASTORE_MYSQL_PASSWORD", "password"),
				Host:     envVarOrDefault("DATASTORE_MYSQL_HOST", "db"),
				Port:     envVarOrDefaultInt("DATASTORE_MYSQL_PORT", 3306),
				DBName:   envVarOrDefault("DATASTORE_MYSQL_DBNAME", "some-database"),
			},
		},
		Queue: queueConfig{
			Type: envVarOrDefault("QUEUE_TYPE", "channels"),
			GooglePubsub: googlePubsubConfig{
				ProjectID:         envVarOrDefault("QUEUE_GOOGLEPUBSUB_PROJECTID", ""),
				PDFToImageTopic:   envVarOrDefault("QUEUE_GOOGLEPUBSUB_PDFTOIMAGEJOBTOPIC", "pdf-splitter"),
				ImageToVideoTopic: envVarOrDefault("QUEUE_GOOGLEPUBSUB_IMAGETOVIDEOTOPIC", "image-to-video"),
				VideoConcatTopic:  envVarOrDefault("QUEUE_GOOGLEPUBSUB_VIDEOCONCATTOPIC", "concatenate-video"),
			},
			Channels: channelsConfig{
				BufferSize:        envVarOrDefaultInt("QUEUE_CHANNELS_BUFFERSIZE", 100),
				PDFToImageTopic:   envVarOrDefault("QUEUE_CHANNELS_PDFTOIMAGETOPIC", "pdf-splitter"),
				ImageToVideoTopic: envVarOrDefault("QUEUE_CHANNELS_IMAGETOVIDEOTOPIC", "image-to-video"),
				VideoConcatTopic:  envVarOrDefault("QUEUE_CHANNELS_VIDEOCONCATTOPIC", "concatenate-video"),
			},
		},
		BlobStorage: blobConfig{
			Type: envVarOrDefault("BLOBSTORAGE_TYPE", "minio"),
			GCS: gcsConfig{
				ProjectID: envVarOrDefault("BLOBSTORAGE_GCS_PROJECTID", ""),
				Bucket:    envVarOrDefault("BLOBSTORAGE_GCS_BUCKET", ""),
				PDFFolder: envVarOrDefault("BLOBSTORAGE_GCS_PDFFOLDER", "pdf"),
			},
			Minio: minioConfig{
				Bucket:              envVarOrDefault("BLOBSTORAGE_MINIO_BUCKET", "videos"),
				Endpoint:            envVarOrDefault("BLOBSTORAGE_MINIO_ENDPOINT", "s3:9000"),
				AccessKeyID:         envVarOrDefault("BLOBSTORAGE_MINIO_ACCESSKEYID", "s3_user"),
				SecretAccessKey:     envVarOrDefault("BLOBSTORAGE_MINIO_SECRETACCESSKEY", "s3_password"),
				PDFFolder:           envVarOrDefault("BLOBSTORAGE_MINIO_PDFFOLDER", "pdf"),
				ImagesFolder:        envVarOrDefault("BLOBSTORAGE_MINIO_IMAGESFOLDER", "images"),
				VideoSnippetsFolder: envVarOrDefault("BLOBSTORAGE_MINIO_VIDEOSNIPPETSFOLDER", "video-snippets"),
				VideoFolder:         envVarOrDefault("BLOBSTORAGE_MINIO_VIDEOFOLDER", "videos"),
			},
		},
	}
	serviceName = "slides-to-video-manager"
	version     = "v0.1.0"

	rootCmd = func() *cobra.Command {
		rootCmd := &cobra.Command{
			Use:   "slides-to-video-manager",
			Short: "Server side manager component to manage slides to video remote workers",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		rootCmd.AddCommand(versionCmd)
		rootCmd.AddCommand(configCmd())
		rootCmd.AddCommand(serverCmd())
		rootCmd.AddCommand(migrateCmd())
		return rootCmd
	}
)

func init() {
	cobra.OnInitialize(initConfig)
}

func main() {
	rootCmd().Execute()
}

func initConfig() {
	configurationFiles := strings.Split(cfgFile, ",")
	for _, cFile := range configurationFiles {
		var readCfg config
		if strings.Contains(cFile, ".yml") || strings.Contains(cFile, ".yaml") {
			raw, err := ioutil.ReadFile(cFile)
			if err != nil {
				fmt.Println("unable to read config file")
				os.Exit(1)
			}
			err = yaml.Unmarshal(raw, &readCfg)
			if err != nil {
				fmt.Println("unable to process config")
				os.Exit(1)
			}
		}
		mergo.Merge(&cfg, readCfg, mergo.WithOverride)
	}
}
