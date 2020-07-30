package main

import (
	"os"
	"strconv"
)

type datastoreConfig struct {
	Type                  string                `yaml:"type"`
	GoogleDatastoreConfig googleDatastoreConfig `yaml:"googleDataStore"`
}

type googleDatastoreConfig struct {
	ProjectID              string `yaml:"projectID"`
	UserTableName          string `yaml:"userTableName"`
	ProjectTableName       string `yaml:"projectTableName"`
	PDFSlidesTableName     string `yaml:"pdfSlidesTableName"`
	VideoSegmentsTableName string `yaml:"videoSegmentsTableName"`
}

type queueConfig struct {
	Type         string             `yaml:"type"`
	GooglePubsub googlePubsubConfig `yaml:"googlePubsub"`
}

type googlePubsubConfig struct {
	ProjectID         string `yaml:"projectID"`
	PDFToImageTopic   string `yaml:"pdfToImageTopic"`
	ImageToVideoTopic string `yaml:"imageToVideoTopic"`
	VideoConcatTopic  string `yaml:"videoConcatTopic"`
}

type serverConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Trace          bool   `yaml:"trace"`
	ClientID       string `yaml:"clientID"`
	ClientSecret   string `yaml:"clientSecret"`
	Scope          string `yaml:"scope"`
	RedirectURI    string `yaml:"redirectURI"`
	AuthSecret     string `yaml:"authSecret"`
	AuthIssuer     string `yaml:"issuer"`
	AuthExpiryTime int    `yaml:"expiryTime"`
}

type blobConfig struct {
	Type string    `yaml:"type"`
	GCS  gcsConfig `yaml:"gcs"`
}

type gcsConfig struct {
	ProjectID string `yaml:"projectID"`
	Bucket    string `yaml:"bucket"`
	PDFFolder string `yaml:"pdfFolder"`
}

type config struct {
	Server      serverConfig    `yaml:"server"`
	Datastore   datastoreConfig `yaml:"datastore"`
	Queue       queueConfig     `yaml:"queue"`
	BlobStorage blobConfig      `yaml:"blobStorage"`
}

func envVarOrDefault(envVar, defaultVal string) string {
	overrideVal, exists := os.LookupEnv(envVar)
	if exists {
		return overrideVal
	}
	return defaultVal
}

func envVarOrDefaultInt(envVar string, defaultVal int) int {
	overrideVal, exists := os.LookupEnv(envVar)
	if exists {
		num, err := strconv.Atoi(overrideVal)
		if err != nil {
			return defaultVal
		}
		return num
	}
	return defaultVal
}
