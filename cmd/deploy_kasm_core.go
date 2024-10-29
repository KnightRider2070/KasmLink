package cmd

import (
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"log"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Docker image",
	Long:  `Build, export, copy, and load a Docker image to the specified target node.`,
	Run: func(cmd *cobra.Command, args []string) {
		imageTag, _ := cmd.Flags().GetString("image-tag")
		dockerfilePath, _ := cmd.Flags().GetString("dockerfile-path")
		targetNodeAddress, _ := cmd.Flags().GetString("target-node-address")
		targetNodePath, _ := cmd.Flags().GetString("target-node-path")
		sshUser, _ := cmd.Flags().GetString("ssh-user")
		sshPassword, _ := cmd.Flags().GetString("ssh-password")

		err := procedures.DeployDockerImage(imageTag, dockerfilePath, targetNodeAddress, targetNodePath, sshUser, sshPassword)
		if err != nil {
			log.Fatalf("Deployment process failed: %v", err)
		} else {
			log.Println("Deployment process completed successfully")
		}
	},
}

func init() {
	// Adding deploy command to root
	RootCmd.AddCommand(deployCmd)

	// Adding flags to deploy command
	deployCmd.Flags().String("image-tag", "my-core-image:latest", "Docker image tag")
	deployCmd.Flags().String("dockerfile-path", "workspace-core-image/dockerfile-kasm-core-suse", "Path to Dockerfile")
	deployCmd.Flags().String("target-node-address", "192.168.0.10:22", "Target node address with port")
	deployCmd.Flags().String("target-node-path", "/tmp/core-image.tar", "Path on target node where tar file will be copied")
	deployCmd.Flags().String("ssh-user", "username", "SSH username for target node")
	deployCmd.Flags().String("ssh-password", "password", "SSH password for target node")
}
