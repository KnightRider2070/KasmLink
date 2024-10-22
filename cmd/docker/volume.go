package cmd

import (
	"fmt"
	"kasmlink/pkg/dockercli"

	"github.com/spf13/cobra"
)

var volumeName string

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage Docker volumes (create, inspect, remove)",
}

var createVolumeCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Docker volume",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.CreateVolume(volumeName)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	dockerCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(createVolumeCmd)

	// Flags for volume commands
	createVolumeCmd.Flags().StringVarP(&volumeName, "name", "n", "", "Volume name")

	// Make volume name required
	createVolumeCmd.MarkFlagRequired("name")
}
