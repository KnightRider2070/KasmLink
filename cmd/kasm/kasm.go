package cmd

import (
	"github.com/spf13/cobra"
)

var kasmCmd = &cobra.Command{
	Use:   "kasm",
	Short: "Manage Kasm users, images, sessions, and utilities",
	Long:  `The kasm command allows you to manage Kasm users, images, sessions, and perform various utility functions through the API.`,
}

func init() {
	rootCmd.AddCommand(kasmCmd)
}
