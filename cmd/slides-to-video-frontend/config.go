package main

import (
	"os"
	"strconv"
)

type config struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Trace          bool   `yaml:"trace"`
	IngressPath    string `yaml:"ingressPath"`
	ServerEndpoint string `yaml:"serverEndpoint"`
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
