package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Version of the CLI tool
const Version = "1.0.0"

// Config holds the flag values
var Config struct {
	BaseURL      string
	ApiSecret    string
	ApiSecretKey string
	SkipTLS      bool
	Verbose      bool
}

// RootCmd is the base command for the Kasm CLI tool.
var RootCmd = &cobra.Command{
	Use:   "kasmlink",
	Short: "Kasm Link CLI",
	Long:  `Kasm Link CLI - A command line tool to manage Kasm resources and Docker components.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required flags
		if Config.BaseURL == "" || Config.ApiSecret == "" || Config.ApiSecretKey == "" {
			fmt.Println("Error: Missing required flags")
			cmd.Usage()
			os.Exit(1)
		}

		fmt.Println("Welcome to Kasm Link CLI. Use 'kasmlink --help' to see available commands.")
		fmt.Printf("Base URL: %s\n", Config.BaseURL)
		fmt.Printf("API Secret: %s\n", Config.ApiSecret)
		fmt.Printf("API Secret Key: %s\n", Config.ApiSecretKey)
		fmt.Printf("Skip TLS: %t\n", Config.SkipTLS)
	},
	Version: Version, // Adding version information to the root command
}

// Execute runs the RootCmd and handles any top-level errors.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func init() {
	// Define persistent flags for API configuration
	RootCmd.PersistentFlags().StringVar(&Config.BaseURL, "base-url", "", "The base URL for the API (required)")
	RootCmd.PersistentFlags().StringVar(&Config.ApiSecret, "api-secret", "", "The API secret for authentication (required)")
	RootCmd.PersistentFlags().StringVar(&Config.ApiSecretKey, "api-secret-key", "", "The API secret key for authentication (required)")
	RootCmd.PersistentFlags().BoolVar(&Config.SkipTLS, "skip-tls", false, "Skip TLS verification (true/false)")

	// Define a verbosity flag
	RootCmd.PersistentFlags().BoolVarP(&Config.Verbose, "verbose", "v", false, "Enable verbose output")

	// Hook to handle version flag
	RootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("Kasm Link CLI Version: %s\n", Version)
			os.Exit(0)
		}

		// Additional pre-run checks
		if Config.BaseURL == "" || Config.ApiSecret == "" || Config.ApiSecretKey == "" {
			fmt.Println("Error: Missing required flags --base-url, --api-secret, or --api-secret-key")
			cmd.Usage()
			os.Exit(1)
		}
	}
}
