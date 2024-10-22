package cmd

import (
	"fmt"
	"kasmlink/pkg/dockercli"

	"github.com/spf13/cobra"
)

var composeFilePath string

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Manage Docker Compose (up, down, etc.)",
}

var composeUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring up Docker Compose services",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.ComposeUp(composeFilePath)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	dockerCmd.AddCommand(composeCmd)
	composeCmd.AddCommand(composeUpCmd)

	// Flags for compose commands
	composeUpCmd.Flags().StringVarP(&composeFilePath, "file", "f", "docker-compose.yml", "Path to docker-compose file")
}
