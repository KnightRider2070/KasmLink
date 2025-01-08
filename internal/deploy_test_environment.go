package internal

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/userService"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowssh"
	"kasmlink/pkg/userParser"
)

// CreateTestEnvironment creates a test environment based on the deployment configuration file.
func CreateTestEnvironment(ctx context.Context, deploymentConfigFilePath, dockerfilePath, buildContextDir string, sshConfig *shadowssh.Config, handler http.RequestHandler) error {
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

	// Initialize Docker Client remote
	executor := dockercli.NewSSHCommandExecutor(sshConfig)
	fs := dockercli.NewRemoteFileSystem(sshClient)
	dockerClientRemote := dockercli.NewDockerClient(executor, fs)

	// Initialize Docker Client local
	executorLocal := dockercli.NewDefaultCommandExecutor()
	fsLocal := dockercli.NewLocalFileSystem()
	dockerClientLocal := dockercli.NewDockerClient(executorLocal, fsLocal)

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

	// Step 2: Process each userService in the configuration
	for _, user := range deploymentConfig.Users {
		log.Info().
			Str("username", user.TargetUser.Username).
			Str("workspace_id", user.WorkspaceID).
			Msg("Processing userService")

		// Lookup the workspace configuration
		workspaceConfig, ok := workspaceMap[user.WorkspaceID]
		if !ok {
			log.Error().Str("workspace_id", user.WorkspaceID).Msg("Workspace configuration not found")
			return fmt.Errorf("workspace configuration not found for workspace ID: %s", user.WorkspaceID)
		}

		// Step 2.1: Check if the Docker image exists and deploy if missing
		if err := ensureDockerImage(ctx, dockerClientLocal, dockerClientRemote, workspaceConfig.ImageConfig.DockerImageName, dockerfilePath, buildContextDir, sshConfig); err != nil {
			return fmt.Errorf("failed to ensure Docker image for userService %s: %w", user.TargetUser.Username, err)
		}

		// Step 2.2: Create or get the userService via the API
		service := userService.NewUserService(handler)

		userID, err := CreateOrGetUser(ctx, service, user)

		if err != nil {
			return fmt.Errorf("failed to create or get userService %s: %w", user.TargetUser.Username, err)
		}

		log.Info().Str("user_id", userID).Msg("User created or retrieved successfully")

		user.TargetUser.UserID = userID

		//TODO: Step 2.3: Assign the users a workspace or start it and dont allow them to access other workspaces

		// Step 2.4: Update deployment configuration with new user details
		if err := userParserInstance.UpdateUserDetails(deploymentConfigFilePath, user.TargetUser.Username, user.TargetUser.UserID, user.KasmSessionID); err != nil {
			log.Error().Err(err).Str("username", user.TargetUser.Username).Msg("Failed to update userService configuration")
			return fmt.Errorf("failed to update userService configuration for userService %s: %w", user.TargetUser.Username, err)
		}
		return nil
	}

	log.Info().Msg("Test environment creation completed successfully")
	return nil
}

// ensureDockerImage checks if a Docker image exists and deploys it if missing.
func ensureDockerImage(ctx context.Context, dockerClientLocal, dockerClientRemote *dockercli.DockerClient, imageTag, dockerfilePath, buildContextDir string, sshConfig *shadowssh.Config) error {
	log.Info().Str("image_tag", imageTag).Msg("Retrieving Docker images on remote node")
	images, err := dockerClientRemote.ListImages(ctx, dockercli.ListImagesOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve Docker images")
		return fmt.Errorf("failed to retrieve Docker images: %w", err)
	}

	// Check if the image exists
	for _, image := range images {
		if image.Tag == imageTag {
			log.Info().Str("image_tag", imageTag).Msg("Docker image already exists on remote node. Skipping deployment.")
			return nil
		}
	}

	log.Info().Str("image_tag", imageTag).Msg("Docker image not found. Attempting remote pull.")

	// Attempt to pull the image remotely
	if err := dockerClientRemote.PullImage(ctx, imageTag); err == nil {
		log.Info().Str("image_tag", imageTag).Msg("Docker image pulled successfully.")
		return nil
	}

	log.Warn().Str("image_tag", imageTag).Msg("Remote pull failed. Falling back to local pull.")

	// Try to pull the image locally
	if err := dockerClientLocal.PullImage(ctx, imageTag); err == nil {
		// Export the image to a tarball and transfer it to the remote node
		if err := dockerClientLocal.TransferImage(ctx, imageTag, sshConfig); err != nil {
			log.Error().Err(err).Msg("Failed to transfer Docker image")
			return fmt.Errorf("failed to transfer Docker image: %w", err)
		}
		log.Info().Str("image_tag", imageTag).Msg("Docker image transferred successfully.")
		return nil
	}

	log.Info().Str("image_tag", imageTag).Msg("Docker image failed to pull locally. Attempting to build remote")

	// Attempt to build the image remotely
	options := dockercli.BuildImageOptions{
		ContextDir:     buildContextDir,
		DockerfilePath: dockerfilePath,
		ImageTag:       imageTag,
		SSH:            sshConfig,
	}
	if err := dockercli.BuildImage(ctx, dockerClientRemote, options); err == nil {
		log.Info().Str("image_tag", imageTag).Msg("Docker image built successfully on remote node.")
		return nil
	}

	log.Warn().Str("image_tag", imageTag).Msg("Remote build failed. Falling back to local build and transfer.")

	if err := dockerClientLocal.TransferImage(ctx, imageTag, sshConfig); err != nil {
		log.Error().Err(err).Msg("Failed to transfer Docker image")
		return fmt.Errorf("failed to transfer Docker image: %w", err)
	}

	return nil
}
