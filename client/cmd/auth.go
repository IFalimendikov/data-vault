package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"data-vault/client/internal/auth"
	"data-vault/client/internal/models"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Authentication command variables
var (
	username string
	password string
)

// registerCmd handles user registration
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  "Register a new user account with the Data Vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
			password = string(passwordBytes)
			fmt.Println()
		}

		if username == "" || password == "" {
			fmt.Fprintf(os.Stderr, "Error: username and password are required\n")
			os.Exit(1)
		}

		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		user := models.User{Login: username, Password: password}
		jwt, err := service.Register(context.Background(), user)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Registration failed: %v\n", err)
			os.Exit(1)
		}

		if err := auth.SaveJWT(jwt, username); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save JWT token: %v\n", err)
		}

		fmt.Printf("Registration successful!\n")
		fmt.Printf("JWT Token: %s\n", jwt)
		fmt.Println("Token saved for future operations.")
	},
}

// loginCmd handles user authentication
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with existing credentials",
	Long:  "Authenticate with the Data Vault server using existing credentials.",
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if password == "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
			password = string(passwordBytes)
			fmt.Println()
		}

		if username == "" || password == "" {
			fmt.Fprintf(os.Stderr, "Error: username and password are required\n")
			os.Exit(1)
		}

		service, err := initService()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
			os.Exit(1)
		}

		user := models.User{Login: username, Password: password}
		jwt, err := service.Login(context.Background(), user)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
			os.Exit(1)
		}

		if err := auth.SaveJWT(jwt, username); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save JWT token: %v\n", err)
		}

		fmt.Printf("Login successful!\n")
		fmt.Printf("JWT Token: %s\n", jwt)
		fmt.Println("Token saved for future operations.")
	},
}

// logoutCmd handles user logout and credential clearing
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear saved credentials",
	Long:  "Remove saved JWT token and logout from the Data Vault client.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.ClearJWT(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing credentials: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Logged out successfully. Credentials cleared.")
	},
}

// init registers authentication commands and their flags
func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)

	registerCmd.Flags().StringVarP(&username, "username", "u", "", "Username for registration")
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "Password for registration")

	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for login")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password for login")
}
