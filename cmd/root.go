package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Version of the CLI tool
const Version = "1.0.0"

// RootCmd is the base command for the Kasm CLI tool.
var RootCmd = &cobra.Command{
	Use:   "kasmlink",
	Short: "Kasm Link CLI",
	Long:  `Kasm Link CLI - A command line tool to manage Kasm resources and Docker components.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Kasm Link CLI. Use 'kasmlink --help' to see available commands.")
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
	// Persistent flag for setting verbosity
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Version flag to print the version
	RootCmd.PersistentFlags().Bool("version", false, "Display the version of Kasm Link CLI")

	// Hook to handle version flag
	RootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("Kasm Link CLI Version: %s\n", Version)
			os.Exit(0)
		}
	}
}
