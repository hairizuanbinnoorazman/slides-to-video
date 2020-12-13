package main

import (
	"os"
	"strconv"
)

var natsQueue = "nats"
var googlePubsubQueue = "google_pubsub"
var gcsBlobStorage = "gcs"
var minioBlobStorage = "minio"

type config struct {
	Server      serverConfig `yaml:"server"`
	Queue       queueConfig  `yaml:"queue"`
	BlobStorage blobConfig   `yaml:"blobStorage"`
}

type serverConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Trace        bool   `yaml:"trace"`
	SvcAcctFile  string `yaml:"svcAccFile"`
	Mode         string `yaml:"mode"`         // Accepts either http or queue - defaults to http
	ProcessRoute string `yaml:"processRoute"` // Only needed when in http mode
	ManagerHost  string `yaml:"managerHost"`
	ManagerPort  int    `yaml:"managerPort"`
}

type blobConfig struct {
	Type         string      `yaml:"type"`
	GCS          gcsConfig   `yaml:"gcs"`
	Minio        minioConfig `yaml:"minio"`
	PDFFolder    string      `yaml:"pdfFolder"`
	ImagesFolder string      `yaml:"imagesFolder"`
}

type gcsConfig struct {
	ProjectID string `yaml:"projectID"`
	Bucket    string `yaml:"bucket"`
}

type minioConfig struct {
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

type queueConfig struct {
	Type            string             `yaml:"type"`
	GooglePubsub    googlePubsubConfig `yaml:"googlePubsub"`
	NatsConfig      natsConfig         `yaml:"nats"`
	PDFToImageTopic string             `yaml:"pdfToImageTopic"`
}

type googlePubsubConfig struct {
	ProjectID string `yaml:"projectID"`
}

type natsConfig struct {
	Endpoint string `yaml:"endpoint"`
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
