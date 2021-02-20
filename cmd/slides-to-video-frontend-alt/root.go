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
		Secure:         envVarOrDefaultBool("SECURE", false),
		Host:           envVarOrDefault("HOST", "0.0.0.0"),
		Port:           envVarOrDefaultInt("PORT", 8080),
		Trace:          false,
		IngressPath:    envVarOrDefault("INGRESS_PATH", ""),
		ServerEndpoint: envVarOrDefault("SERVER_ENDPOINT", "http://localhost:8080"),
	}
	serviceName = "slides-to-video-frontend"
	version     = "v0.1.0"

	rootCmd = func() *cobra.Command {
		rootCmd := &cobra.Command{
			Use:   "slides-to-video-frontend",
			Short: "Slides to video frontend",
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
