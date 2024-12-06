package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	shadowssh "kasmlink/pkg/sshmanager"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
	"path/filepath"
)

// CreateTestEnvironment creates a test environment based on the user configuration file.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - userConfigurationFilePath: Path to the user configuration YAML file.
// - sshConfig: SSH configuration for connecting to the remote node.
// Returns:
// - An error if any step in the environment creation process fails.
func CreateTestEnvironment(ctx context.Context, userConfigurationFilePath string, sshConfig *shadowssh.SSHConfig, kasmApi *webApi.KasmAPI) error {
	// Initialize UserParser
	userParserInstance := userParser.NewUserParser()

	// Step 1: Load user configuration from YAML file
	log.Info().
		Str("config_file", userConfigurationFilePath).
		Msg("Loading user configuration from YAML file")

	usersConfig, err := userParserInstance.LoadConfig(userConfigurationFilePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("config_file", userConfigurationFilePath).
			Msg("Failed to load user configuration")
		return fmt.Errorf("failed to load user configuration: %w", err)
	}

	log.Info().
		Int("user_count", len(usersConfig.UserDetails)).
		Msg("Successfully loaded user configuration")

	// Step 2: Establish SSH connection with remote node using sshConfig
	log.Info().
		Str("host", sshConfig.Host).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(ctx, sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.Host).
			Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().
				Err(cerr).
				Msg("Failed to close SSH connection gracefully")
		} else {
			log.Debug().
				Msg("SSH connection closed")
		}
	}()

	// Step 3: Iterate over each user in the configuration
	for _, user := range usersConfig.UserDetails {
		log.Info().
			Str("username", user.TargetUser.Username).
			Str("docker_image_tag", user.AssignedContainerTag).
			Msg("Processing user")

		// Step 3.1: Ensure that DockerImageTag exists on the remote node
		missingImages, err := checkRemoteImages(ctx, client, []string{user.AssignedContainerTag})
		if err != nil {
			log.Error().
				Err(err).
				Str("image_tag", user.AssignedContainerTag).
				Msg("Error checking Docker image on remote node")
			return fmt.Errorf("error checking Docker image %s on remote node: %w", user.AssignedContainerTag, err)
		}

		if len(missingImages) > 0 {
			log.Info().
				Str("image_tag", user.AssignedContainerTag).
				Msg("Required Docker image tag does not exist on remote node. Deploying image.")

			// Step 3.2: Deploy the missing Docker image
			// Assume DockerfilePath is known or derived based on image tag
			//TODO: Implement this function as needed
			//dockerfilePath := getDockerfilePath(user.AssignedContainerTag)
			dockerfilePath := filepath.Join("path", "to", "Dockerfile")

			if err := DeployImages(ctx, dockerfilePath, user.AssignedContainerTag, sshConfig); err != nil {
				log.Error().
					Err(err).
					Str("image_tag", user.AssignedContainerTag).
					Msg("Failed to deploy Docker image to remote node")
				return fmt.Errorf("failed to deploy Docker image %s: %w", user.AssignedContainerTag, err)
			}
		} else {
			log.Info().
				Str("image_tag", user.AssignedContainerTag).
				Msg("Docker image tag already exists on remote node. Skipping deployment.")
		}

		// Step 3.3: Create or retrieve the user via KASM API
		log.Info().
			Str("username", user.TargetUser.Username).
			Msg("Creating or retrieving user via KASM API")

		userID, err := createOrGetUser(ctx, kasmApi, user)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", user.TargetUser.Username).
				Msg("Failed to create or retrieve user via KASM API")
			return fmt.Errorf("failed to create or retrieve user %s: %w", user.TargetUser.Username, err)
		}
		user.TargetUser.UserID = userID

		/*	// Step 3.4: Add the user to the specified group via KASM API
			log.Info().
				Str("username", user.TargetUser.Username).
				Str("role", user.Role).
				Msg("Adding user to the specified group via KASM API")

			groupID, err := getGroupIDByName(ctx, kasmApi, user.Role)
			if err != nil {
				log.Error().
					Err(err).
					Str("role", user.Role).
					Msg("Failed to retrieve group ID from KASM API")
				return fmt.Errorf("failed to retrieve group ID for role %s: %w", user.Role, err)
			}

			if err := addUserToGroup(ctx, user.TargetUser.UserID, groupID); err != nil {
				log.Error().
					Err(err).
					Str("user_id", user.TargetUser.UserID).
					Str("group_id", groupID).
					Msg("Failed to add user to group via KASM API")
				return fmt.Errorf("failed to add user %s to group %s: %w", user.TargetUser.Username, groupID, err)
			}

			log.Info().
				Str("user_id", user.TargetUser.UserID).
				Str("group_id", groupID).
				Msg("Successfully added user to group via KASM API")
		*/

		// Step 3.5: Update the YAML file with UserID and KasmSessionOfContainer
		// TODO: Implement logic to obtain the actual KasmSessionOfContainer
		iamgeID, _ := getImageIDbyTag(ctx, kasmApi, user.AssignedContainerTag)
		kasmRequestResponse, err := kasmApi.RequestKasmSession(ctx, user.TargetUser.UserID, iamgeID, user.EnvironmentArgs)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", user.TargetUser.Username).
				Msg("Failed to generate KasmSessionOfContainer")
			return fmt.Errorf("failed to generate KasmSessionOfContainer for user %s: %w", user.TargetUser.Username, err)
		}

		log.Info().
			Str("username", user.TargetUser.Username).
			Str("user_id", user.TargetUser.UserID).
			Str("kasm_session_of_container", kasmRequestResponse.KasmID).
			Str("url: ", kasmRequestResponse.KasmURL).
			Msg("Updating user configuration in YAML file")

		if err := userParserInstance.UpdateUserConfig(userConfigurationFilePath, user.TargetUser.Username, user.TargetUser.UserID, kasmRequestResponse.KasmID); err != nil {
			log.Error().
				Err(err).
				Str("username", user.TargetUser.Username).
				Msg("Failed to update user configuration in YAML file")
			return fmt.Errorf("failed to update user %s configuration: %w", user.TargetUser.Username, err)
		}

		log.Info().
			Str("username", user.TargetUser.Username).
			Msg("Successfully updated user configuration in YAML file")
	}

	log.Info().
		Msg("Test environment creation completed successfully")

	return nil
}
