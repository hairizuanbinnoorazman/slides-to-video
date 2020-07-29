package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Print the version number of slides-to-video-manager",
		Long:  `Print the version number of slides-to-video-manager`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("VERSION - PENDING")
		},
	}
)
