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
	cfgFile     string
	readCfg     config
	serviceName = "pdf-splitter"
	version     = "v0.1.0"

	cfg = config{
		Server: serverConfig{
			Host:         envVarOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:         envVarOrDefaultInt("SERVER_PORT", 8081),
			Mode:         envVarOrDefault("SERVER_MODE", "http"),
			ProcessRoute: envVarOrDefault("SERVER_PROCESSROUTE", "/"),
			ManagerHost:  envVarOrDefault("SERVER_MANAGERHOST", "localhost"),
			ManagerPort:  envVarOrDefaultInt("SERVER_MANAGERPORT", 8080),
		},
		BlobStorage: blobConfig{
			Type:         envVarOrDefault("BLOBSTORAGE_TYPE", "minio"),
			PDFFolder:    envVarOrDefault("BLOBSTORAGE_PDFFOLDER", "pdf"),
			ImagesFolder: envVarOrDefault("BLOBSTORAGE_IMAGESFOLDER", "images"),
			GCS: gcsConfig{
				ProjectID: envVarOrDefault("BLOBSTORAGE_GCS_PROJECTID", ""),
				Bucket:    envVarOrDefault("BLOBSTORAGE_GCS_BUCKET", ""),
			},
			Minio: minioConfig{
				Bucket:          envVarOrDefault("BLOBSTORAGE_MINIO_BUCKET", "videos"),
				Endpoint:        envVarOrDefault("BLOBSTORAGE_MINIO_ENDPOINT", "locahost:9000"),
				AccessKeyID:     envVarOrDefault("BLOBSTORAGE_MINIO_ACCESSKEY", "s3_user"),
				SecretAccessKey: envVarOrDefault("BLOBSTORAGE_MINIO_SECRETKEY", "s3_password"),
			},
		},
	}

	rootCmd = func() *cobra.Command {
		rootCmd := &cobra.Command{
			Use:   "pdf-splitter",
			Short: "Component of slides to video to split pdf to multiple images",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		rootCmd.AddCommand(versionCmd)
		rootCmd.AddCommand(configCmd())
		rootCmd.AddCommand(serverCmd())
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
