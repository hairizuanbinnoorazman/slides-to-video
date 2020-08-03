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
	readCfg config

	// Includes default configuration
	// Initial configuration is set to utilize Google Datastore and Google Pubsub for now
	// Immediately replaces value with environment variables on startup
	// TODO: Utilize Inmemory queue and inmemory datastores in the future
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
		Datastore: datastoreConfig{
			Type: envVarOrDefault("DATASTORE_TYPE", "google_datastore"),
			GoogleDatastoreConfig: googleDatastoreConfig{
				ProjectID:              envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PROJECTID", ""),
				UserTableName:          envVarOrDefault("DATASTORE_GOOGLEDATASTORE_USERTABLENAME", "UserTable"),
				ProjectTableName:       envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PROJECTTABLENAME", "ProjectTable"),
				PDFSlidesTableName:     envVarOrDefault("DATASTORE_GOOGLEDATASTORE_PDFSLIDESTABLENAME", "PDFSlideTable"),
				VideoSegmentsTableName: envVarOrDefault("DATASTORE_GOOGLEDATASTORE_VIDEOSEGMENTSTABLENAME", "VideoSegmentsTable"),
			},
		},
		Queue: queueConfig{
			Type: envVarOrDefault("QUEUE_TYPE", "google_pubsub"),
			GooglePubsub: googlePubsubConfig{
				ProjectID:         envVarOrDefault("QUEUE_GOOGLEPUBSUB_PROJECTID", ""),
				PDFToImageTopic:   envVarOrDefault("QUEUE_GOOGLEPUBSUB_PDFTOIMAGEJOBTOPIC", "pdf-splitter"),
				ImageToVideoTopic: envVarOrDefault("QUEUE_GOOGLEPUBSUB_IMAGETOVIDEOTOPIC", "image-to-video"),
				VideoConcatTopic:  envVarOrDefault("QUEUE_GOOGLEPUBSUB_VIDEOCONCATTOPIC", "video-concat"),
			},
		},
		BlobStorage: blobConfig{
			Type: envVarOrDefault("BLOBSTORAGE_TYPE", "gcs"),
			GCS: gcsConfig{
				ProjectID: envVarOrDefault("BLOBSTORAGE_GCS_PROJECTID", ""),
				Bucket:    envVarOrDefault("BLOBSTORAGE_GCS_BUCKET", ""),
				PDFFolder: envVarOrDefault("BLOBSTORAGE_GCS_PDFFOLDER", "pdf"),
			},
		},
	}
	serviceName = "slides-to-video-manager"
	version     = "v0.1.0"

	rootCmd = &cobra.Command{
		Use:   "slides-to-video-manager",
		Short: "Server side manager component to manage slides to video remote workers",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(migrateCmd)

	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(validateCmd)

	serverCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")

	validateCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")

	migrateCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")
}

func main() {
	rootCmd.Execute()
}

func initConfig() {
	if strings.Contains(cfgFile, ".yml") || strings.Contains(cfgFile, ".yaml") {
		raw, err := ioutil.ReadFile(cfgFile)
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
