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
var channelsQueue = "channels"
var gcsBlobStorage = "gcs"
var minioBlobStorage = "minio"
var localBlobStorage = "local"
var s3BlobStorage = "s3"

var googleTTS = "google"
var amazonPollyTTS = "amazon_polly"

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
	Channels     channelsConfig     `yaml:"channels"`
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

type channelsConfig struct {
	BufferSize        int    `yaml:"bufferSize"`
	PDFToImageTopic   string `yaml:"pdfToImageTopic"`
	ImageToVideoTopic string `yaml:"imageToVideoTopic"`
	VideoConcatTopic  string `yaml:"videoConcatTopic"`
}

type serverConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Scheme         string `yaml:"scheme"` // "http" or "https"
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
	S3    s3Config        `yaml:"s3"`
}

type gcsConfig struct {
	ProjectID          string `yaml:"projectID"`
	Bucket             string `yaml:"bucket"`
	PDFFolder          string `yaml:"pdfFolder"`
	ImagesFolder       string `yaml:"imagesFolder"`
	VideoSnippetsFolder string `yaml:"videoSnippetsFolder"`
	VideoFolder        string `yaml:"videoFolder"`
}

type minioConfig struct {
	Bucket              string `yaml:"bucket"`
	Endpoint            string `yaml:"endpoint"`
	AccessKeyID         string `yaml:"accessKeyId"`
	SecretAccessKey     string `yaml:"secretAccessKey"`
	PDFFolder           string `yaml:"pdfFolder"`
	ImagesFolder        string `yaml:"imagesFolder"`
	VideoSnippetsFolder string `yaml:"videoSnippetsFolder"`
	VideoFolder         string `yaml:"videoFolder"`
}

type localBlobConfig struct {
	BasePath            string `yaml:"basePath"`
	PDFFolder           string `yaml:"pdfFolder"`
	ImagesFolder        string `yaml:"imagesFolder"`
	VideoSnippetsFolder string `yaml:"videoSnippetsFolder"`
	VideoFolder         string `yaml:"videoFolder"`
}

type s3Config struct {
	Region              string `yaml:"region"`
	Bucket              string `yaml:"bucket"`
	PDFFolder           string `yaml:"pdfFolder"`
	ImagesFolder        string `yaml:"imagesFolder"`
	VideoSnippetsFolder string `yaml:"videoSnippetsFolder"`
	VideoFolder         string `yaml:"videoFolder"`
}

type ttsConfig struct {
	Type        string            `yaml:"type"`
	AmazonPolly amazonPollyConfig `yaml:"amazonPolly"`
}

type amazonPollyConfig struct {
	Region  string `yaml:"region"`
	VoiceID string `yaml:"voiceId"`
	Engine  string `yaml:"engine"`
}

type workersConfig struct {
	Enabled          bool                 `yaml:"enabled"`
	PDFSplitter      workerInstanceConfig `yaml:"pdfSplitter"`
	ImageToVideo     workerInstanceConfig `yaml:"imageToVideo"`
	ConcatenateVideo workerInstanceConfig `yaml:"concatenateVideo"`
}

type workerInstanceConfig struct {
	Enabled     bool `yaml:"enabled"`
	Concurrency int  `yaml:"concurrency"`
}

type config struct {
	Server      serverConfig    `yaml:"server"`
	Workers     workersConfig   `yaml:"workers"`
	Datastore   datastoreConfig `yaml:"datastore"`
	Queue       queueConfig     `yaml:"queue"`
	BlobStorage blobConfig      `yaml:"blobStorage"`
	TTS         ttsConfig       `yaml:"tts"`
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

func envVarOrDefaultBool(envVar string, defaultVal bool) bool {
	overrideVal, exists := os.LookupEnv(envVar)
	if exists {
		b, err := strconv.ParseBool(overrideVal)
		if err != nil {
			return defaultVal
		}
		return b
	}
	return defaultVal
}

func ConfigStructLevelValidation(sl validator.StructLevel) {
	cfg := sl.Current().Interface().(config)

	// Validate datastore configuration
	if cfg.Datastore.Type == mysqlDatastore {
		if cfg.Datastore.MySQLConfig == nil {
			sl.ReportError(cfg.Datastore.MySQLConfig, "mysql", "MySQLConfig", "required", "MySQL configuration is required when using MySQL datastore")
		} else if cfg.Datastore.MySQLConfig.DBName == "" || cfg.Datastore.MySQLConfig.Host == "" || cfg.Datastore.MySQLConfig.Password == "" || cfg.Datastore.MySQLConfig.User == "" || cfg.Datastore.MySQLConfig.Port == 0 {
			sl.ReportError(cfg.Datastore.MySQLConfig, "mysql", "MySQLConfig", "required", "MySQL configuration fields (DBName, Host, Password, User, Port) must all be set")
		}
	}

	// Validate worker concurrency when workers are enabled
	if cfg.Workers.Enabled {
		if cfg.Workers.PDFSplitter.Enabled && cfg.Workers.PDFSplitter.Concurrency <= 0 {
			sl.ReportError(cfg.Workers.PDFSplitter, "pdfSplitter", "Concurrency", "concurrency_positive", "PDF splitter worker is enabled but concurrency is not set or is <= 0")
		}
		if cfg.Workers.ImageToVideo.Enabled && cfg.Workers.ImageToVideo.Concurrency <= 0 {
			sl.ReportError(cfg.Workers.ImageToVideo, "imageToVideo", "Concurrency", "concurrency_positive", "Image-to-video worker is enabled but concurrency is not set or is <= 0")
		}
		if cfg.Workers.ConcatenateVideo.Enabled && cfg.Workers.ConcatenateVideo.Concurrency <= 0 {
			sl.ReportError(cfg.Workers.ConcatenateVideo, "concatenateVideo", "Concurrency", "concurrency_positive", "Concatenate-video worker is enabled but concurrency is not set or is <= 0")
		}
	}

	// Validate blob storage folder fields for the selected backend
	if cfg.BlobStorage.Type == gcsBlobStorage {
		if cfg.BlobStorage.GCS.ProjectID == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.ProjectID", "ProjectID", "required", "GCS ProjectID is required when using GCS blob storage")
		}
		if cfg.BlobStorage.GCS.Bucket == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.Bucket", "Bucket", "required", "GCS Bucket is required when using GCS blob storage")
		}
		if cfg.BlobStorage.GCS.PDFFolder == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.PDFFolder", "PDFFolder", "required", "GCS PDFFolder is required when using GCS blob storage")
		}
		if cfg.BlobStorage.GCS.ImagesFolder == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.ImagesFolder", "ImagesFolder", "required", "GCS ImagesFolder is required when using GCS blob storage")
		}
		if cfg.BlobStorage.GCS.VideoSnippetsFolder == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.VideoSnippetsFolder", "VideoSnippetsFolder", "required", "GCS VideoSnippetsFolder is required when using GCS blob storage")
		}
		if cfg.BlobStorage.GCS.VideoFolder == "" {
			sl.ReportError(cfg.BlobStorage.GCS, "gcs.VideoFolder", "VideoFolder", "required", "GCS VideoFolder is required when using GCS blob storage")
		}
	} else if cfg.BlobStorage.Type == minioBlobStorage {
		if cfg.BlobStorage.Minio.Bucket == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.Bucket", "Bucket", "required", "Minio Bucket is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.Endpoint == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.Endpoint", "Endpoint", "required", "Minio Endpoint is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.AccessKeyID == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.AccessKeyID", "AccessKeyID", "required", "Minio AccessKeyID is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.SecretAccessKey == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.SecretAccessKey", "SecretAccessKey", "required", "Minio SecretAccessKey is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.PDFFolder == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.PDFFolder", "PDFFolder", "required", "Minio PDFFolder is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.ImagesFolder == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.ImagesFolder", "ImagesFolder", "required", "Minio ImagesFolder is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.VideoSnippetsFolder == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.VideoSnippetsFolder", "VideoSnippetsFolder", "required", "Minio VideoSnippetsFolder is required when using Minio blob storage")
		}
		if cfg.BlobStorage.Minio.VideoFolder == "" {
			sl.ReportError(cfg.BlobStorage.Minio, "minio.VideoFolder", "VideoFolder", "required", "Minio VideoFolder is required when using Minio blob storage")
		}
	} else if cfg.BlobStorage.Type == localBlobStorage {
		if cfg.BlobStorage.Local.BasePath == "" {
			sl.ReportError(cfg.BlobStorage.Local, "local.BasePath", "BasePath", "required", "Local BasePath is required when using local blob storage")
		}
		if cfg.BlobStorage.Local.PDFFolder == "" {
			sl.ReportError(cfg.BlobStorage.Local, "local.PDFFolder", "PDFFolder", "required", "Local PDFFolder is required when using local blob storage")
		}
		if cfg.BlobStorage.Local.ImagesFolder == "" {
			sl.ReportError(cfg.BlobStorage.Local, "local.ImagesFolder", "ImagesFolder", "required", "Local ImagesFolder is required when using local blob storage")
		}
		if cfg.BlobStorage.Local.VideoSnippetsFolder == "" {
			sl.ReportError(cfg.BlobStorage.Local, "local.VideoSnippetsFolder", "VideoSnippetsFolder", "required", "Local VideoSnippetsFolder is required when using local blob storage")
		}
		if cfg.BlobStorage.Local.VideoFolder == "" {
			sl.ReportError(cfg.BlobStorage.Local, "local.VideoFolder", "VideoFolder", "required", "Local VideoFolder is required when using local blob storage")
		}
	} else if cfg.BlobStorage.Type == s3BlobStorage {
		if cfg.BlobStorage.S3.Region == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.Region", "Region", "required", "S3 Region is required when using S3 blob storage")
		}
		if cfg.BlobStorage.S3.Bucket == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.Bucket", "Bucket", "required", "S3 Bucket is required when using S3 blob storage")
		}
		if cfg.BlobStorage.S3.PDFFolder == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.PDFFolder", "PDFFolder", "required", "S3 PDFFolder is required when using S3 blob storage")
		}
		if cfg.BlobStorage.S3.ImagesFolder == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.ImagesFolder", "ImagesFolder", "required", "S3 ImagesFolder is required when using S3 blob storage")
		}
		if cfg.BlobStorage.S3.VideoSnippetsFolder == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.VideoSnippetsFolder", "VideoSnippetsFolder", "required", "S3 VideoSnippetsFolder is required when using S3 blob storage")
		}
		if cfg.BlobStorage.S3.VideoFolder == "" {
			sl.ReportError(cfg.BlobStorage.S3, "s3.VideoFolder", "VideoFolder", "required", "S3 VideoFolder is required when using S3 blob storage")
		}
	}
}
