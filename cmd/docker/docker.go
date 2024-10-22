package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "dockercli",
	Short: "A unified CLI for managing Docker containers, images, networks, volumes, and Docker Compose.",
	Long: `dockercli is a comprehensive tool to help you manage Docker containers, images,
networks, volumes, and Docker Compose with a single command line interface.`,
}

// Execute runs the root command and subcommands
func Execute() {
	if err := dockerCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
