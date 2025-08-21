package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check server connectivity",
	Long:  "Test the connection to the Data Vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		if service.PingServer(context.Background()) {
			fmt.Println("✓ Server is reachable!")
		} else {
			fmt.Println("✗ Server is not reachable!")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
