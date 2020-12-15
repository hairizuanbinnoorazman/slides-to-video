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
	serviceName = "concatenate-video"
	version     = "v0.1.0"

	cfg = config{}

	rootCmd = func() *cobra.Command {
		rootCmd := &cobra.Command{
			Use:   "concatenate-video",
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
