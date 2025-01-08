package cmd

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal"
	"kasmlink/pkg/shadowssh"
)

// Initialize the "environment" command and its subcommands.
func init() {
	// Define the root "environment" command.
	environmentCmd := &cobra.Command{
		Use:   "environment",
		Short: "Manage test environment creation",
		Long:  `Commands for creating and managing test environments based on deployment configurations.`,
	}

	// Add subcommands.
	environmentCmd.AddCommand(createCreateTestEnvironmentCommand())

	// Add "environment" to the root command.
	RootCmd.AddCommand(environmentCmd)
}

// createCreateTestEnvironmentCommand creates a command for creating a test environment.
func createCreateTestEnvironmentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [deploymentConfigFilePath] [dockerfilePath] [buildContextDir] [sshHost] [sshPort] [sshUser] [sshPassword]",
		Short: "Create a test environment based on a deployment configuration",
		Long: `This command sets up a test environment using the provided deployment configuration file. 
It builds or deploys Docker images, assigns resources, and updates configurations as specified.`,
		Args: cobra.ExactArgs(7),
		Run: func(cmd *cobra.Command, args []string) {
			deploymentConfigFilePath := args[0]
			dockerfilePath := args[1]
			buildContextDir := args[2]
			sshHost := args[3]
			sshPort := args[4]
			sshUser := args[5]
			sshPassword := args[6]

			log.Info().
				Str("config_file", deploymentConfigFilePath).
				Str("dockerfile_path", dockerfilePath).
				Str("context_dir", buildContextDir).
				Str("host", sshHost).
				Msg("Starting test environment creation")

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

			// Execute the CreateTestEnvironment function.
			ctx := context.Background()
			err = internal.CreateTestEnvironment(ctx, deploymentConfigFilePath, dockerfilePath, buildContextDir, sshConfig)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create test environment")
				return
			}

			duration := time.Since(startTime)
			log.Info().
				Dur("duration", duration).
				Msg("Test environment creation completed successfully")
		},
	}
}
