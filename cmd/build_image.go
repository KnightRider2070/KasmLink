package cmd

import (
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"log"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Docker image",
	Long:  `Build a Docker image from a specified build context directory and Dockerfile.`,
	Run: func(cmd *cobra.Command, args []string) {
		buildContextDir, _ := cmd.Flags().GetString("build-context-dir")
		dockerfilePath, _ := cmd.Flags().GetString("dockerfile-path")
		imageTag, _ := cmd.Flags().GetString("image-tag")

		err := procedures.BuildContainer(buildContextDir, dockerfilePath, imageTag)
		if err != nil {
			log.Fatalf("Build process failed: %v", err)
		} else {
			log.Println("Build process completed successfully")
		}
	},
}

func init() {

	// Adding build command to root
	RootCmd.AddCommand(buildCmd)

	// Adding flags to build command
	buildCmd.Flags().String("build-context-dir", "./", "Path to the build context directory")
	buildCmd.Flags().String("dockerfile-path", "Dockerfile", "Path to Dockerfile")
	buildCmd.Flags().String("image-tag", "my-built-image:latest", "Docker image tag")
}
