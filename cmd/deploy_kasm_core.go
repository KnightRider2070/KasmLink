package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [image-tag] [base-image (optional)] [dockerfile-path] [target-node-address] [target-node-path] [ssh-user] [ssh-password]",
	Short: "Deploy a Docker image",
	Long:  `Build, export, copy, and load a Docker image to the specified target node.`,
	Args:  cobra.RangeArgs(6, 7),
	Run: func(cmd *cobra.Command, args []string) {
		// Assign required arguments
		imageTag := args[0]
		baseImage := ""
		dockerfilePath := ""
		targetNodeAddress := ""
		targetNodePath := ""
		sshUser := ""
		sshPassword := ""

		// If baseImage is provided as the second argument
		if len(args) == 7 {
			baseImage = args[1]
			dockerfilePath = args[2]
			targetNodeAddress = args[3]
			targetNodePath = args[4]
			sshUser = args[5]
			sshPassword = args[6]
		} else { // If baseImage is not provided
			dockerfilePath = args[1]
			targetNodeAddress = args[2]
			targetNodePath = args[3]
			sshUser = args[4]
			sshPassword = args[5]
		}

		// Run deployment with final arguments
		err := procedures.DeployKasmDockerImage(imageTag, baseImage, dockerfilePath, targetNodeAddress, targetNodePath, sshUser, sshPassword)
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
