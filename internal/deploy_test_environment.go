package internal

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/userService"
	"kasmlink/pkg/api/workspace"
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

	// Ensure Docker images exist and deploy them if missing
	for _, workspace := range deploymentConfig.Workspaces {
		imageTag := workspace.ImageConfig.DockerImageName

		if err := ensureDockerImage(ctx, dockerClientLocal, dockerClientRemote, imageTag, dockerfilePath, buildContextDir, sshConfig); err != nil {
			return fmt.Errorf("failed to ensure Docker image %s: %w", imageTag, err)
		}
	}

	workspaceService := workspace.NewWorkspaceService(handler)

	// Step 1.1: Create workspaces
	for _, workspace := range deploymentConfig.Workspaces {
		log.Info().Str("workspaceName", workspace.ImageConfig.FriendlyName).Msg("Creating workspace")

		targetImageResponse, err := workspaceService.CreateWorkspace(workspace.ImageConfig)

		if err != nil {
			return fmt.Errorf("failed to create workspace %s: %w", workspace.ImageConfig.FriendlyName, err)
		}

		log.Info().Str("workspace_id", targetImageResponse.ImageID).Msg("Workspace created successfully")

		// Update deployment configuration with new workspace details
		workspace.ImageConfig = *targetImageResponse
	}

	// Create groups and add workspaces to them as needed
	for _, group := range deploymentConfig.Groups {
		log.Info().Str("groupName", group.Name).Msg("Creating group")

		userService := userService.NewUserService(handler)

		var groupToCreate = models.Group{
			Name:        group.Name,
			Priority:    group.Priority,
			Description: group.Description,
		}

		groupResponse, err := userService.CreateGroup(groupToCreate)

		if err != nil {
			return fmt.Errorf("failed to create group %s: %w", group.Name, err)
		}

		group := groupResponse.Groups[0]

		log.Info().Str("group_id", group.GroupID).Msg("Group created successfully")

		// Add workspaces to the group
		for _, workspaceName := range group.WorkspaceNames {
			// Lookup the workspaceID in the workspaceMap
			workspaceConfig := workspaceMap[workspaceName]

			workspaceID := workspaceConfig.WorkspaceID
			// Add the workspace to the group
			err := userService.AddImageToGroup(group.GroupID, workspaceID)
			if err != nil {
				return fmt.Errorf("failed to add workspace %s to group %s: %w", workspaceName, group.Name, err)
			}
			log.Info().Str("workspace_id", workspaceID).Str("group_id", group.GroupID).Msg("Workspace added to group successfully")
		}
	}

	// Create users and add them to groups
	for _, user := range deploymentConfig.Users {

		log.Info().Str("username", user.TargetUser.Username).Msg("Creating user")

		userService := userService.NewUserService(handler)

		userResponse, err := userService.CreateUser(user.TargetUser)

		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.TargetUser.Username, err)
		}

		user.TargetUser.UserID = userResponse.UserID

		log.Info().Str("user_id", user.TargetUser.UserID).Msg("User created successfully")

		// Get the group ID from the group name
		var groupID string
		for _, group := range deploymentConfig.Groups {
			if group.Name == user.GroupName {
				groupID = group.GroupID
				break
			}
		}

		// Add user to group
		err = userService.AddUserToGroup(user.TargetUser.UserID, groupID)

		if err != nil {
			return fmt.Errorf("failed to add user %s to group %s: %w", user.TargetUser.Username, user.GroupName, err)
		}

		log.Info().Str("group_name", user.GroupName).Msg("User added to group successfully")

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
