package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
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

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
}
