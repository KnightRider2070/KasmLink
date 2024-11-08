package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"os"
)

// Command to build the core image for Kasm.
var buildCoreImageCmd = &cobra.Command{
	Use:   "build-core-image [imageTag] [baseImage]",
	Short: "Build the core Docker image for Kasm",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		baseImage := args[1]

		err := procedures.BuildCoreImageKasm(imageTag, baseImage)
		if err != nil {
			fmt.Printf("Error building Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image built successfully")
	},
}

// Command to deploy a Docker image on a remote node.
var deployImageCmd = &cobra.Command{
	Use:   "deploy-image [imageTag] [baseImage] [dockerfilePath] [targetNodePath]",
	Short: "Deploy the Docker image on a remote node",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		baseImage := args[1]
		dockerfilePath := args[2]
		targetNodePath := args[3]

		err := procedures.DeployKasmDockerImage(imageTag, baseImage, dockerfilePath, targetNodePath)
		if err != nil {
			fmt.Printf("Error deploying Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image deployed successfully on remote node")
	},
}

// Command to deploy a Docker Compose file to a remote node.
var deployComposeCmd = &cobra.Command{
	Use:   "deploy-compose [composeFilePath] [targetNodePath]",
	Short: "Deploy Docker Compose services on a remote node",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		composeFilePath := args[0]
		targetNodePath := args[1]

		err := procedures.DeployComposeFile(composeFilePath, targetNodePath)
		if err != nil {
			fmt.Printf("Error deploying Docker Compose file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker Compose file deployed successfully on remote node")
	},
}

// Initialize and add all commands to root.
func init() {
	RootCmd.AddCommand(buildCoreImageCmd)
	RootCmd.AddCommand(deployImageCmd)
	RootCmd.AddCommand(deployComposeCmd)
}
