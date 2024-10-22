package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercli"
)

var imageName string

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage Docker images (pull, push, remove, etc.)",
}

var pullImageCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull a Docker image from a registry",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.PullImage(imageName)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var removeImageCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Docker image",
	Run: func(cmd *cobra.Command, args []string) {
		err := dockercli.RemoveImage(imageName)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	dockerCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(pullImageCmd)
	imageCmd.AddCommand(removeImageCmd)

	// Flags for image commands
	pullImageCmd.Flags().StringVarP(&imageName, "image", "i", "", "Image name")
	removeImageCmd.Flags().StringVarP(&imageName, "image", "i", "", "Image name")

	// Make image name required
	pullImageCmd.MarkFlagRequired("image")
	removeImageCmd.MarkFlagRequired("image")
}
