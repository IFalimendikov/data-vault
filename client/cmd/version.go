package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display the version, build date, and commit information for the Data Vault client.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Data Vault Client\n")
		fmt.Printf("Version: %s\n", buildVersion)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("Build Commit: %s\n", buildCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
