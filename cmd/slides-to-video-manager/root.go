package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cfgFile     string
	cfg         config
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
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)

	configCmd.AddCommand(initCmd)

	serveCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Configuration File")
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
		err = yaml.Unmarshal(raw, &cfg)
		if err != nil {
			fmt.Println("unable to process config")
			os.Exit(1)
		}
	}
}
