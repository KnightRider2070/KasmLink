package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercli"
)

var containerID string

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Manage Docker containers (start, stop, remove, etc.)",
}

var startContainerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Docker container by ID or name",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.StartContainer(containerID)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var stopContainerCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Docker container by ID or name",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.StopContainer(containerID)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	dockerCmd.AddCommand(containerCmd)
	containerCmd.AddCommand(startContainerCmd)
	containerCmd.AddCommand(stopContainerCmd)

	// Flags for container commands
	startContainerCmd.Flags().StringVarP(&containerID, "id", "i", "", "Container ID or name")
	stopContainerCmd.Flags().StringVarP(&containerID, "id", "i", "", "Container ID or name")

	// Make container ID a required flag
	startContainerCmd.MarkFlagRequired("id")
	stopContainerCmd.MarkFlagRequired("id")
}
