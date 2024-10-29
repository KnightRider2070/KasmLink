package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
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

		err := procedures.BuildContainer(buildContextDir, dockerfilePath, imageTag)
		if err != nil {
			log.Fatal().Err(err).Msg("Build process failed")
		} else {
			log.Info().Msg("Build process completed successfully")
		}
	},
}

func init() {
	// Adding build command to root
	RootCmd.AddCommand(buildCmd)
}
