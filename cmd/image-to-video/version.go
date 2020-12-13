package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of image-to-video",
		Long:  `Print the version number of image-to-video`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%v\n", version)
		},
	}
)
