package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd is the base command for the Kasm CLI tool.
var rootCmd = &cobra.Command{
	Use:   "kasmlink",
	Short: "Kasm Link CLI",
	Long:  `T.B.D.A`,
	Run: func(cmd *cobra.Command, args []string) {
		// This will be executed if no subcommands are provided
		fmt.Println("Welcome to the Kasm Link. Use 'kasmlink --help' to see available commands.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("apikey", "k", "", "API key for authentication")
	rootCmd.PersistentFlags().StringP("secret", "s", "", "API secret for authentication")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
