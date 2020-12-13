package main

type config struct {
	Server serverConfig `yaml:"server"`
}

type serverConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Trace       bool   `yaml:"trace"`
	SvcAcctFile string `yaml:"svcAccFile"`
}
