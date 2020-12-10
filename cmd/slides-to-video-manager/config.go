package main

import (
	"os"
	"strconv"

	"gopkg.in/go-playground/validator.v9"
)

var mysqlDatastore = "mysql"
var googleDatastore = "google_datastore"
var natsQueue = "nats"
var googlePubsubQueue = "google_pubsub"
var gcsBlobStorage = "gcs"
var minioBlobStorage = "minio"

type datastoreConfig struct {
	Type                  string                 `yaml:"type"`
	GoogleDatastoreConfig *googleDatastoreConfig `yaml:"googleDataStore"`
	MySQLConfig           *mysqlConfig           `yaml:"mysql"`
}

type googleDatastoreConfig struct {
	ProjectID              string `yaml:"projectID"`
	UserTableName          string `yaml:"userTableName"`
	ProjectTableName       string `yaml:"projectTableName"`
	PDFSlidesTableName     string `yaml:"pdfSlidesTableName"`
	VideoSegmentsTableName string `yaml:"videoSegmentsTableName"`
}

type mysqlConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"dbName"`
}

type queueConfig struct {
	Type         string             `yaml:"type"`
	GooglePubsub googlePubsubConfig `yaml:"googlePubsub"`
	NatsConfig   natsConfig         `yaml:"nats"`
}

type googlePubsubConfig struct {
	ProjectID         string `yaml:"projectID"`
	PDFToImageTopic   string `yaml:"pdfToImageTopic"`
	ImageToVideoTopic string `yaml:"imageToVideoTopic"`
	VideoConcatTopic  string `yaml:"videoConcatTopic"`
}

type natsConfig struct {
	Endpoint          string `yaml:"endpoint"`
	PDFToImageTopic   string `yaml:"pdfToImageTopic"`
	ImageToVideoTopic string `yaml:"imageToVideoTopic"`
	VideoConcatTopic  string `yaml:"videoConcatTopic"`
}

type serverConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Trace          bool   `yaml:"trace"`
	SvcAcctFile    string `yaml:"svcAccFile"`
	ClientID       string `yaml:"clientID"`
	ClientSecret   string `yaml:"clientSecret"`
	Scope          string `yaml:"scope"`
	RedirectURI    string `yaml:"redirectURI"`
	AuthSecret     string `yaml:"authSecret"`
	AuthIssuer     string `yaml:"issuer"`
	AuthExpiryTime int    `yaml:"expiryTime"`
}

type blobConfig struct {
	Type  string          `yaml:"type"`
	GCS   gcsConfig       `yaml:"gcs"`
	Minio minioConfig     `yaml:"minio"`
	Local localBlobConfig `yaml:"local"`
}

type gcsConfig struct {
	ProjectID string `yaml:"projectID"`
	Bucket    string `yaml:"bucket"`
	PDFFolder string `yaml:"pdfFolder"`
}

type minioConfig struct {
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	PDFFolder       string `yaml:"pdfFolder"`
}

type localBlobConfig struct {
	Folder    string `yaml:"folder"`
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

func ConfigStructLevelValidation(sl validator.StructLevel) {
	cfg := sl.Current().Interface().(config)

	if cfg.Datastore.Type == mysqlDatastore {
		if cfg.Datastore.GoogleDatastoreConfig != nil {
			sl.ReportError(cfg.Datastore.GoogleDatastoreConfig, "googleDatastore", "GoogleDatastoreConfig", "", "")
		}
		if cfg.Datastore.MySQLConfig.DBName == "" || cfg.Datastore.MySQLConfig.Host == "" || cfg.Datastore.MySQLConfig.Password == "" || cfg.Datastore.MySQLConfig.User == "" || cfg.Datastore.MySQLConfig.Port == 0 {
			sl.ReportError(cfg.Datastore.GoogleDatastoreConfig, "mysql", "MySQLConfig", "", "")
		}
	}
}
