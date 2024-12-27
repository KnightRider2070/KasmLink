package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowscp"
	"kasmlink/pkg/shadowssh"
	"kasmlink/pkg/userParser"
)

// CreateTestEnvironment creates a test environment based on the deployment configuration file.
func CreateTestEnvironment(ctx context.Context, deploymentConfigFilePath, dockerfilePath, buildContextDir string, sshConfig *shadowssh.Config) error {
	// Validate required file paths
	if deploymentConfigFilePath == "" || dockerfilePath == "" || buildContextDir == "" {
		return fmt.Errorf("deployment configuration file, Dockerfile path, and build context directory must be specified")
	}

	// Initialize UserParser
	userParserInstance := userParser.NewUserParser()

	// Establish SSH connection
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

	// Step 1: Load deployment configuration from YAML file
	log.Info().Str("config_file", deploymentConfigFilePath).Msg("Loading deployment configuration from YAML file")
	deploymentConfig, err := userParserInstance.LoadDeploymentConfig(deploymentConfigFilePath)
	if err != nil {
		log.Error().Err(err).Str("config_file", deploymentConfigFilePath).Msg("Failed to load deployment configuration")
		return fmt.Errorf("failed to load deployment configuration: %w", err)
	}
	log.Info().Int("workspace_count", len(deploymentConfig.Workspaces)).
		Int("user_count", len(deploymentConfig.Users)).Msg("Successfully loaded deployment configuration")

	// Create a map of workspaces for quick lookup
	workspaceMap := make(map[string]userParser.WorkspaceConfig)
	for _, workspace := range deploymentConfig.Workspaces {
		workspaceMap[workspace.WorkspaceID] = workspace
	}

	// Step 2: Process each user in the configuration
	for _, user := range deploymentConfig.Users {
		log.Info().
			Str("username", user.TargetUser.Username).
			Str("workspace_id", user.WorkspaceID).
			Msg("Processing user")

		// Lookup the workspace configuration
		workspaceConfig, ok := workspaceMap[user.WorkspaceID]
		if !ok {
			log.Error().Str("workspace_id", user.WorkspaceID).Msg("Workspace configuration not found")
			return fmt.Errorf("workspace configuration not found for workspace ID: %s", user.WorkspaceID)
		}

		// Step 2.1: Check if the Docker image exists on the remote node
		if err := ensureDockerImage(ctx, dockerClient, sshClient, workspaceConfig.ImageConfig.Name, dockerfilePath, buildContextDir, sshConfig); err != nil {
			return fmt.Errorf("failed to ensure Docker image for user %s: %w", user.TargetUser.Username, err)
		}

		// Step 2.2: Assign resources and update configuration
		if err := assignResourcesAndUpdateConfig(userParserInstance, deploymentConfigFilePath, user, workspaceConfig); err != nil {
			return fmt.Errorf("failed to assign resources or update configuration for user %s: %w", user.TargetUser.Username, err)
		}
	}

	log.Info().Msg("Test environment creation completed successfully")
	return nil
}

// ensureDockerImage checks if a Docker image exists and deploys it if missing, falling back to local build and transfer.
func ensureDockerImage(ctx context.Context, dockerClient *dockercli.DockerClient, sshClient *shadowssh.Client, imageTag, dockerfilePath, buildContextDir string, sshConfig *shadowssh.Config) error {
	log.Info().Str("image_tag", imageTag).Msg("Retrieving Docker images")
	images, err := dockerClient.ListImages(ctx, dockercli.ListImagesOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve Docker images")
		return fmt.Errorf("failed to retrieve Docker images: %w", err)
	}

	// Check if the image exists
	for _, image := range images {
		if image.Repository == imageTag {
			log.Info().Str("image_tag", imageTag).Msg("Docker image already exists on remote node. Skipping deployment.")
			return nil
		}
	}

	log.Info().Str("image_tag", imageTag).Msg("Docker image not found. Attempting remote build.")

	// Attempt to build the image remotely
	options := dockercli.BuildImageOptions{
		ContextDir:     buildContextDir,
		DockerfilePath: dockerfilePath,
		ImageTag:       imageTag,
		SSH:            sshConfig,
	}
	if err := dockercli.BuildImage(ctx, dockerClient, options); err == nil {
		log.Info().Str("image_tag", imageTag).Msg("Docker image built successfully on remote node.")
		return nil
	}

	log.Warn().Str("image_tag", imageTag).Msg("Remote build failed. Falling back to local build and transfer.")

	// Build locally and transfer the image
	tempTarPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.tar", imageTag))
	if err := dockerClient.SaveImage(ctx, imageTag, tempTarPath); err != nil {
		log.Error().Err(err).Msg("Failed to save image locally.")
		return fmt.Errorf("failed to save image locally: %w", err)
	}
	defer os.Remove(tempTarPath)

	// Transfer the tarball to the remote node
	remoteTarPath := fmt.Sprintf("/tmp/%s.tar", imageTag)
	if err := shadowscp.CopyFileToRemote(ctx, tempTarPath, remoteTarPath, sshConfig); err != nil {
		log.Error().Err(err).Msg("Failed to transfer tarball to remote node.")
		return fmt.Errorf("failed to transfer tarball to remote node: %w", err)
	}

	// Load the image on the remote node
	loadCommand := fmt.Sprintf("docker load -i %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, loadCommand); err != nil {
		log.Error().Err(err).Msg("Failed to load image on remote node.")
		return fmt.Errorf("failed to load image on remote node: %w", err)
	}
	log.Info().Str("image_tag", imageTag).Msg("Image successfully loaded on remote node.")

	// Clean up remote tarball
	removeCommand := fmt.Sprintf("rm -f %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, removeCommand); err != nil {
		log.Warn().Err(err).Str("tar_path", remoteTarPath).Msg("Failed to remove tarball from remote node.")
	}

	return nil
}

// assignResourcesAndUpdateConfig assigns resources to a user and updates the configuration.
func assignResourcesAndUpdateConfig(parser *userParser.UserParser, configFilePath string, user userParser.UserDetails, workspaceConfig userParser.WorkspaceConfig) error {
	log.Info().Str("username", user.TargetUser.Username).Msg("Creating user and assigning resources")

	// Generate resource identifiers
	userID := fmt.Sprintf("generated_user_%s", user.TargetUser.Username)
	kasmSessionID := fmt.Sprintf("session_%s", workspaceConfig.WorkspaceID)

	// Update user configuration with session details
	log.Info().Str("username", user.TargetUser.Username).Msg("Updating user configuration with session details")
	if err := parser.UpdateUserDetails(configFilePath, user.TargetUser.Username, userID, kasmSessionID); err != nil {
		log.Error().Err(err).Str("username", user.TargetUser.Username).Msg("Failed to update user configuration")
		return fmt.Errorf("failed to update user configuration for user %s: %w", user.TargetUser.Username, err)
	}
	return nil
}
