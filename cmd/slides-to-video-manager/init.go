package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Subcommand to handle config admin functionality of this tool",
		Long:  `Provides capabilities such as initializing an initial configuration as well as parsing`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the configuration for the tool",
		Long: `There are various fields to be filled up in order to run the configuration.
One can try to initialize the configuration in order to quickly get started with it`,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := yaml.Marshal(cfg)
			fmt.Println(string(raw))
		},
	}

	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  `Check the configuration to make sure that the configuration`,
		Run: func(cmd *cobra.Command, args []string) {
			raw, _ := yaml.Marshal(cfg)
			fmt.Println(string(raw))
		},
	}
)
