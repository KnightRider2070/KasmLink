package internal

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowssh"
	"kasmlink/pkg/userParser"
	"path/filepath"
)

// CreateTestEnvironment creates a test environment based on the user configuration file.
func CreateTestEnvironment(ctx context.Context, userConfigurationFilePath string, sshConfig *shadowssh.Config) error {
	// Initialize UserParser
	userParserInstance := userParser.NewUserParser()

	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Str("host", sshConfig.Host).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	// Initialize Docker Client
	executor := dockercli.NewSSHCommandExecutor(sshConfig)
	fs := dockercli.NewRemoteFileSystem(sshClient)
	dockerClient := dockercli.NewDockerClient(executor, fs)

	// Step 1: Load user configuration from YAML file
	log.Info().Str("config_file", userConfigurationFilePath).Msg("Loading user configuration from YAML file")

	usersConfig, err := userParserInstance.LoadConfig(userConfigurationFilePath)
	if err != nil {
		log.Error().Err(err).Str("config_file", userConfigurationFilePath).Msg("Failed to load user configuration")
		return fmt.Errorf("failed to load user configuration: %w", err)
	}

	log.Info().Int("user_count", len(usersConfig.UserDetails)).Msg("Successfully loaded user configuration")

	// Step 2: Establish SSH connection with remote node
	log.Info().Str("host", sshConfig.Host).Str("user", sshConfig.Username).Msg("Establishing SSH connection to remote node")

	sshClient, err = shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Str("host", sshConfig.Host).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	// Step 3: Iterate over each user in the configuration
	for _, user := range usersConfig.UserDetails {
		log.Info().Str("username", user.TargetUser.Username).Str("docker_image_tag", user.AssignedContainerTag).Msg("Processing user")

		// Step 3.1: Check if the Docker image exists on the remote node
		checkImageCmd := fmt.Sprintf("docker images -q %s", user.AssignedContainerTag)
		output, err := sshClient.ExecuteCommand(ctx, checkImageCmd)
		if err != nil || output == "" {
			log.Info().Str("image_tag", user.AssignedContainerTag).Msg("Docker image not found on remote node. Deploying image.")

			// Step 3.2: Deploy the missing Docker image
			dockerfilePath := filepath.Join("path", "to", "Dockerfile")
			if err := DeployImage(ctx, dockerfilePath, user.AssignedContainerTag, dockerClient, sshConfig); err != nil {
				log.Error().Err(err).Str("image_tag", user.AssignedContainerTag).Msg("Failed to deploy Docker image")
				return fmt.Errorf("failed to deploy Docker image %s: %w", user.AssignedContainerTag, err)
			}
		} else {
			log.Info().Str("image_tag", user.AssignedContainerTag).Msg("Docker image already exists on remote node. Skipping deployment.")
		}

		// Step 3.3: Create user logic (standalone approach)
		log.Info().Str("username", user.TargetUser.Username).Msg("Creating user and assigning resources")
		// Assume standalone logic to create user and assign session ID
		userID := fmt.Sprintf("generated_user_%s", user.TargetUser.Username)
		kasmSessionID := fmt.Sprintf("session_%s", user.AssignedContainerTag)
		containerID := fmt.Sprintf("container_%s", user.AssignedContainerTag)

		// Step 3.4: Update the YAML configuration with session details
		log.Info().Str("username", user.TargetUser.Username).Msg("Updating user configuration with session details")
		if err := userParserInstance.UpdateUserConfig(userConfigurationFilePath, user.TargetUser.Username, userID, kasmSessionID, containerID); err != nil {
			log.Error().Err(err).Str("username", user.TargetUser.Username).Msg("Failed to update user configuration")
			return fmt.Errorf("failed to update user %s configuration: %w", user.TargetUser.Username, err)
		}
	}

	log.Info().Msg("Test environment creation completed successfully")
	return nil
}
