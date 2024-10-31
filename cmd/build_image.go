package cmd

import (
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockerutils"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build [build-context-dir] [dockerfile-path] [image-tag]",
	Short: "Build a Docker image",
	Long:  `Build a Docker image from a specified build context directory and Dockerfile.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		buildContextDir := args[0]
		dockerfilePath := args[1]
		imageTag := args[2]

		// Create a Docker client
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal().Err(err).Msg("Could not create Docker client")
		}

		// Create tar archive from the build context directory
		buildContextTar, err := dockerutils.CreateTarWithContext(buildContextDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create build context tar")
		}

		// Call BuildDockerImage with the tar build context
		err = dockerutils.BuildDockerImage(cli, imageTag, dockerfilePath, buildContextTar, nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Build process failed")
		} else {
			log.Info().Msg("Build process completed successfully")
		}
	},
}

func init() {
	// Adding build command to the root command
	RootCmd.AddCommand(buildCmd)
}
