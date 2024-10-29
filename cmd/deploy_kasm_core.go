package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [image-tag] [dockerfile-path] [target-node-address] [target-node-path] [ssh-user] [ssh-password]",
	Short: "Deploy a Docker image",
	Long:  `Build, export, copy, and load a Docker image to the specified target node.`,
	Args:  cobra.ExactArgs(6),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		dockerfilePath := args[1]
		targetNodeAddress := args[2]
		targetNodePath := args[3]
		sshUser := args[4]
		sshPassword := args[5]

		err := procedures.DeployDockerImage(imageTag, dockerfilePath, targetNodeAddress, targetNodePath, sshUser, sshPassword)
		if err != nil {
			log.Fatal().Err(err).Msg("Deployment process failed")
		} else {
			log.Info().Msg("Deployment process completed successfully")
		}
	},
}

func init() {
	// Adding deploy command to root
	RootCmd.AddCommand(deployCmd)
}
