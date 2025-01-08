package cmd

import (
	"context"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/server"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowssh"
)

// Initialize the "backend" command and its subcommands.
func init() {
	// Define the root "backend" command.
	backendCmd := &cobra.Command{
		Use:   "backend",
		Short: "Manage backend services deployment",
		Long:  `Commands for managing the deployment of backend services, including validation, transfer, and remote execution.`,
	}

	// Add subcommands.
	backendCmd.AddCommand(createDeployBackendCommand())

	// Add "backend" to the root command.
	RootCmd.AddCommand(backendCmd)
}

// createDeployBackendCommand creates a command for deploying backend services.
func createDeployBackendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy [backendComposePath] [sshHost] [sshPort] [sshUser] [sshPassword]",
		Short: "Deploy backend services using Docker Compose on a remote server",
		Long: `This command validates a Docker Compose file, checks for missing images, 
transfers them to the remote server, and runs 'docker compose up' to deploy services.`,
		Args: cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			backendComposePath := args[0]
			sshHost := args[1]
			sshPort := args[2]
			sshUser := args[3]
			sshPassword := args[4]

			log.Info().
				Str("compose_path", backendComposePath).
				Str("host", sshHost).
				Msg("Starting backend services deployment")

			// sshPort to int conversion
			sshPortInt, err := strconv.Atoi(sshPort)

			if err != nil {
				log.Error().Err(err).Msg("Failed to convert SSH port to integer")
				return
			}

			startTime := time.Now()

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

			// Deploy backend services.
			ctx := context.Background()
			err = internal.DeployBackendServices(ctx, backendComposePath, sshConfig, dockerClient)
			if err != nil {
				log.Error().Err(err).Msg("Failed to deploy backend services")
				return
			}

			// Set server settings for workspace deployment
			serverService := server.NewServerSettingsService(handler)
			serverService.UpdateAddWorkspaceToAllGroupsVar(false)

			duration := time.Since(startTime)
			log.Info().
				Dur("duration", duration).
				Msg("Backend deployment completed successfully")
		},
	}
}
