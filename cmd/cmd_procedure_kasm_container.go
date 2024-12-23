package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/internal"
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

		err := internal.BuildCoreImageKasm(imageTag, baseImage)
		if err != nil {
			fmt.Printf("Error building Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image built successfully")
	},
}

// Command to deploy a Docker image on a remote node.
var deployImageCmd = &cobra.Command{
	Use:   "deploy-image [imageTag] [baseImage] [targetNodePath]",
	Short: "Deploy the Docker image on a remote node",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		baseImage := args[1]
		targetNodePath := args[2]

		// Get the local tar file path flag
		localTarFilePath, err := cmd.Flags().GetString("local-tar-file")
		if err != nil {
			fmt.Printf("Error reading local-tar-file flag: %v\n", err)
			os.Exit(1)
		}

		// Call the deploy function with the optional localTarFilePath
		err = internal.DeployKasmDockerImage(imageTag, baseImage, targetNodePath, localTarFilePath)
		if err != nil {
			fmt.Printf("Error deploying Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image deployed successfully on remote node")
	},
}

func init() {
	// Register the local-tar-file flag for optional local file path
	deployImageCmd.Flags().String("local-tar-file", "", "Optional path to a local tar file to use instead of building a new image")
}

// Command to deploy a Docker Compose file to a remote node.
var deployComposeCmd = &cobra.Command{
	Use:   "deploy-compose [composeFilePath] [targetNodePath]",
	Short: "Deploy Docker Compose services on a remote node",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		composeFilePath := args[0]
		targetNodePath := args[1]

		err := internal.DeployComposeFile(composeFilePath, targetNodePath)
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
