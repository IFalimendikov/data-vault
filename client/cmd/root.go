package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the CLI application
var rootCmd = &cobra.Command{
	Use:   "data-vault-client",
	Short: "A secure CLI client for Data Vault server",
	Long: `Data Vault Client is a command-line application for securely storing,
retrieving, and managing encrypted data via gRPC connection to Data Vault server.

The client provides user authentication, data operations, and server connectivity
across Windows, Linux, and macOS platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		tuiCmd.Run(cmd, args)
	},
}

// Execute runs the root command and handles errors
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// init configures the root command flags
func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
}
