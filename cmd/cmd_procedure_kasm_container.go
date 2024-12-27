package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowscp"
	"kasmlink/pkg/shadowssh"
)

// Global Docker Client
var dockerClient = dockercli.NewDockerClient(
	&dockercli.DefaultCommandExecutor{},
	&dockercli.LocalFileSystem{},
)

// SSH configuration flags
var sshHost, sshUser, sshPassword string
var sshPort int

// Init SSH flags for commands requiring SSH
func initSSHFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&sshHost, "ssh-host", "", "Remote host for SSH connection")
	cmd.Flags().StringVar(&sshUser, "ssh-user", "", "Username for SSH connection")
	cmd.Flags().StringVar(&sshPassword, "ssh-password", "", "Password for SSH connection")
	cmd.Flags().IntVar(&sshPort, "ssh-port", 22, "Port for SSH connection")
}

// Build the core Docker image for Kasm.
var buildCoreImageCmd = &cobra.Command{
	Use:   "build-core-image [imageTag] [baseImage]",
	Short: "Build the core Docker image for Kasm",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		baseImage := args[1]

		// Build options
		options := dockercli.BuildImageOptions{
			ContextDir:     "./path/to/build/context", // Update with the correct path
			DockerfilePath: "./path/to/Dockerfile",    // Update with the correct Dockerfile path
			ImageTag:       imageTag,
			BuildArgs: map[string]string{
				"BASE_IMAGE": baseImage,
			},
		}

		// Build the image
		ctx := context.Background()
		if err := dockercli.BuildImage(ctx, dockerClient, options); err != nil {
			log.Error().Err(err).Msg("Failed to build Docker image")
			fmt.Printf("Error building Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image built successfully")
	},
}

// Deploy a Docker image on a remote node.
var deployImageCmd = &cobra.Command{
	Use:   "deploy-image [imageTag]",
	Short: "Deploy the Docker image on a remote node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]

		// Parse SSH configuration from flags
		sshConfig := shadowssh.Config{
			Host:     sshHost,
			Username: sshUser,
			Password: sshPassword,
			Port:     sshPort,
		}

		// Optional tar file for deployment
		localTarFilePath, err := cmd.Flags().GetString("local-tar-file")
		if err != nil {
			log.Error().Err(err).Msg("Failed to read local-tar-file flag")
			fmt.Printf("Error reading local-tar-file flag: %v\n", err)
			os.Exit(1)
		}

		// Deploy the image
		ctx := context.Background()
		if localTarFilePath != "" {
			err = dockerClient.TransferImage(ctx, imageTag, &sshConfig)
		} else {
			options := dockercli.BuildImageOptions{
				ContextDir:     "./path/to/build/context", // Update with the correct path
				DockerfilePath: "./path/to/Dockerfile",    // Update with the correct Dockerfile path
				ImageTag:       imageTag,
				SSH:            &sshConfig,
			}
			err = dockercli.BuildImage(ctx, dockerClient, options)
		}

		if err != nil {
			log.Error().Err(err).Msg("Failed to deploy Docker image")
			fmt.Printf("Error deploying Docker image: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker image deployed successfully on remote node")
	},
}

// Deploy a Docker Compose file to a remote node.
var deployComposeCmd = &cobra.Command{
	Use:   "deploy-compose [composeFilePath]",
	Short: "Deploy Docker Compose services on a remote node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		composeFilePath := args[0]

		// Parse SSH configuration from flags
		sshConfig := shadowssh.Config{
			Host:     sshHost,
			Username: sshUser,
			Password: sshPassword,
			Port:     sshPort,
		}

		// Deploy the Docker Compose file
		ctx := context.Background()
		if err := shadowscp.CopyFileToRemote(ctx, composeFilePath, "/dockercompose", &sshConfig); err != nil {
			log.Error().Err(err).Msg("Failed to deploy Docker Compose file")
			fmt.Printf("Error deploying Docker Compose file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Docker Compose file deployed successfully on remote node")
	},
}

func init() {
	// Initialize SSH flags for commands requiring SSH
	initSSHFlags(deployImageCmd)
	initSSHFlags(deployComposeCmd)

	// Register additional flags for deployImageCmd
	deployImageCmd.Flags().String("local-tar-file", "", "Optional path to a local tar file to use instead of building a new image")

	// Add commands to the root command
	RootCmd.AddCommand(buildCoreImageCmd)
	RootCmd.AddCommand(deployImageCmd)
	RootCmd.AddCommand(deployComposeCmd)
}
