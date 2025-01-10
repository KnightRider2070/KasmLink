package cmd

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowssh"
)

// Initialize the "image" command and its subcommands.
func init() {
	// Define the root "image" command.
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage Docker image deployments",
		Long:  `Commands for deploying Docker images to remote nodes.`,
	}

	// Add subcommands.
	imageCmd.AddCommand(createDeployImageCommand())

	// Add "image" to the root command.
	RootCmd.AddCommand(imageCmd)
}

// createDeployImageCommand creates a command for deploying a Docker image.
func createDeployImageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy [imageName] [tarDirectory] [sshHost] [sshPort] [sshUser] [sshPassword]",
		Short: "Deploy a Docker image to a remote node",
		Long: `This command deploys a specified Docker image to a remote node. 
It ensures the image exists locally, exports it to a tarball if necessary, transfers the tarball, 
and loads it on the remote node.`,
		Args: cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			imageName := args[0]
			tarDirectory := args[1]
			sshHost := args[2]
			sshPort := args[3]
			sshUser := args[4]
			sshPassword := args[5]

			log.Info().
				Str("image_name", imageName).
				Str("tar_directory", tarDirectory).
				Str("host", sshHost).
				Msg("Starting Docker image deployment")

			startTime := time.Now()

			// sshPort to int conversion
			sshPortInt, err := strconv.Atoi(sshPort)

			if err != nil {
				log.Error().Err(err).Msg("Failed to convert SSH port to integer")
				return
			}

			// Prepare the SSH configuration.
			sshConfig := &shadowssh.Config{
				Host:     sshHost,
				Port:     sshPortInt,
				Username: sshUser,
				Password: sshPassword,
			}

			// Initialize the Docker client.
			dockerClient := dockercli.NewDockerClient(
				&dockercli.DefaultCommandExecutor{},
				nil, // Assuming no custom file system interface is needed.
			)

			// Deploy the Docker image.
			ctx := context.Background()
			err = internal.DeployImage(ctx, imageName, tarDirectory, dockerClient, sshConfig)
			if err != nil {
				log.Error().Err(err).Msg("Failed to deploy Docker image")
				return
			}

			duration := time.Since(startTime)
			log.Info().
				Dur("duration", duration).
				Msg("Docker image deployment completed successfully")
		},
	}
}
