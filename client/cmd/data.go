package main

import (
	"context"
	"fmt"
	"os"

	"data-vault/client/internal/auth"

	"github.com/spf13/cobra"
)

var (
	jwtToken string // JWT token for authentication
	dataText string // Text data to be stored
	dataID   string // Data ID for operations
)

// dataCmd represents the data command group
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Data operations (post, get, delete)",
	Long:  "Perform data operations like storing, retrieving, and deleting data from the vault.",
}

// postCmd represents the data post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Store data in the vault",
	Long:  "Store encrypted data in the Data Vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		// Auto-load JWT if not provided
		if jwtToken == "" {
			savedJWT, err := auth.LoadJWT()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading saved credentials: %v\n", err)
				os.Exit(1)
			}
			jwtToken = savedJWT
		}

		if jwtToken == "" {
			fmt.Fprintf(os.Stderr, "Error: Not authenticated. Please login first with 'data-vault-client login'\n")
			os.Exit(1)
		}

		if dataText == "" {
			fmt.Print("Enter data to store: ")
			fmt.Scanln(&dataText)
		}

		if dataText == "" {
			fmt.Fprintf(os.Stderr, "Error: data is required\n")
			os.Exit(1)
		}

		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		err = service.PostData(context.Background(), jwtToken, dataText)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to post data: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Data posted successfully!")
	},
}

// getCmd represents the data get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve data from the vault",
	Long:  "Retrieve all your stored data from the Data Vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		// Auto-load JWT if not provided
		if jwtToken == "" {
			savedJWT, err := auth.LoadJWT()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading saved credentials: %v\n", err)
				os.Exit(1)
			}
			jwtToken = savedJWT
		}

		if jwtToken == "" {
			fmt.Fprintf(os.Stderr, "Error: Not authenticated. Please login first with 'data-vault-client login'\n")
			os.Exit(1)
		}

		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		data, err := service.GetData(context.Background(), jwtToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get data: %v\n", err)
			os.Exit(1)
		}

		if len(data) == 0 {
			fmt.Println("No data found.")
			return
		}

		fmt.Println("Your stored data:")
		for i, item := range data {
			fmt.Printf("%d. ID: %s\n   Data: %s\n   Uploaded: %s\n\n",
				i+1, item.ID, item.Data, item.UploadedAt)
		}
	},
}

// deleteCmd represents the data delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete data from the vault",
	Long:  "Delete a specific data entry from the Data Vault server by ID.",
	Run: func(cmd *cobra.Command, args []string) {
		// Auto-load JWT if not provided
		if jwtToken == "" {
			savedJWT, err := auth.LoadJWT()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading saved credentials: %v\n", err)
				os.Exit(1)
			}
			jwtToken = savedJWT
		}

		if jwtToken == "" {
			fmt.Fprintf(os.Stderr, "Error: Not authenticated. Please login first with 'data-vault-client login'\n")
			os.Exit(1)
		}

		if dataID == "" {
			fmt.Print("Enter data ID to delete: ")
			fmt.Scanln(&dataID)
		}

		if dataID == "" {
			fmt.Fprintf(os.Stderr, "Error: data ID is required\n")
			os.Exit(1)
		}

		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		err = service.DeleteData(context.Background(), jwtToken, dataID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete data: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Data with ID %s deleted successfully!\n", dataID)
	},
}

// init registers data commands and sets up their flags
func init() {
	rootCmd.AddCommand(dataCmd)
	dataCmd.AddCommand(postCmd)
	dataCmd.AddCommand(getCmd)
	dataCmd.AddCommand(deleteCmd)

	dataCmd.PersistentFlags().StringVar(&jwtToken, "jwt", "", "JWT token for authentication")

	postCmd.Flags().StringVarP(&dataText, "data", "d", "", "Data to store")
	deleteCmd.Flags().StringVar(&dataID, "id", "", "ID of data to delete")
}
