package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Runs database migration (if necessary)",
		Long:  `If one utilizes relational databases such as MySQL, Postgresql - that would require usage data schema migration to happen`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("VERSION - PENDING")
		},
	}
)
