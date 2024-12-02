package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	shadowssh "kasmlink/pkg/sshmanager"
	"os"
	"path/filepath"
	"strings"
)

// findDockerfileForService searches for a Dockerfile in the ./dockerfiles/ directory that contains the serviceName.
func findDockerfileForService(serviceName string) (string, error) {
	log.Debug().
		Str("service_name", serviceName).
		Msg("Searching for Dockerfile matching service name")

	dockerfilesDir := "./dockerfiles"
	pattern := fmt.Sprintf("*%s*", serviceName)

	matchedFiles, err := filepath.Glob(filepath.Join(dockerfilesDir, pattern))
	if err != nil {
		log.Error().
			Err(err).
			Str("pattern", pattern).
			Msg("Failed to glob Dockerfiles")
		return "", fmt.Errorf("failed to glob dockerfiles: %w", err)
	}

	if len(matchedFiles) == 0 {
		log.Error().
			Str("service_name", serviceName).
			Str("directory", dockerfilesDir).
			Msg("No Dockerfile found containing service name")
		return "", fmt.Errorf("no Dockerfile found containing service name '%s' in %s", serviceName, dockerfilesDir)
	}

	if len(matchedFiles) > 1 {
		log.Error().
			Str("service_name", serviceName).
			Strs("matched_files", matchedFiles).
			Msg("Multiple Dockerfiles found for service name")
		return "", fmt.Errorf("multiple Dockerfiles found for service '%s' in %s: %v", serviceName, dockerfilesDir, matchedFiles)
	}

	log.Debug().
		Str("dockerfile", matchedFiles[0]).
		Msg("Found matching Dockerfile")

	return matchedFiles[0], nil
}

// sanitizeImageName sanitizes the image name to create valid filenames.
func sanitizeImageName(imageName string) string {
	// Replace '/' and ':' with underscores to prevent directory traversal or invalid filenames.
	return strings.ReplaceAll(strings.ReplaceAll(imageName, "/", "_"), ":", "_")
}

// checkLocalImageTarExists checks if the image tar file exists locally.
func checkLocalImageTarExists(imageName string) (bool, string) {
	sanitizedImageName := sanitizeImageName(imageName)
	localTarPath := filepath.Join("./tarfiles", fmt.Sprintf("%s.tar", sanitizedImageName))

	_, err := os.Stat(localTarPath)
	if err == nil {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar exists locally")
		return true, localTarPath
	}
	if os.IsNotExist(err) {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar does not exist locally")
		return false, localTarPath
	}
	// For other errors, log and treat as non-existent
	log.Error().
		Err(err).
		Str("local_tar_path", localTarPath).
		Msg("Error checking local tar file existence")
	return false, localTarPath
}

// CreateTestEnvironment creates a test environment based on the user configuration file.
func CreateTestEnvironment(userConfigurationFilePath string, sshConfig shadowssh.SSHConfig) error {
	// Step 1: Load user configuration from YAML file
	log.Info().
		Str("config_file", userConfigurationFilePath).
		Msg("Loading user configuration from YAML file")

	usersConfig, err := userParser.LoadConfig(userConfigurationFilePath)
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
		Str("host", sshConfig.NodeAddress).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
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

	// Iterate over each user in the configuration
	for _, user := range usersConfig.UserDetails {
		log.Info().
			Str("username", user.Username).
			Str("docker_image_tag", user.AssignedContainerTag).
			Msg("Processing user")

		// Step 3: Ensure that DockerImageTag exists on the remote node
		imageExists, err := checkRemoteImages(client, user.AssignedContainerTag)
		if err != nil {
			log.Error().
				Err(err).
				Str("image_tag", user.AssignedContainerTag).
				Msg("Error checking Docker image on remote node")
			return fmt.Errorf("error checking Docker image %s on remote node: %w", user.AssignedContainerTag, err)
		}

		if !imageExists {
			log.Error().
				Str("image_tag", user.AssignedContainerTag).
				Msg("Required Docker image tag does not exist on remote node")
			return fmt.Errorf("Docker image tag %s does not exist on remote node", user.AssignedContainerTag)
		}

		// Step 4: Create the user via KASM API
		log.Info().
			Str("username", user.Username).
			Msg("Creating user via KASM API")

		// Check if user already exists
		existingUser, err := api.GetUser("", user.Username)
		if err != nil {
			// If the error indicates that the user does not exist, proceed to create
			// Else, return the error
			if !strings.Contains(err.Error(), "not found") {
				log.Error().
					Err(err).
					Str("username", user.Username).
					Msg("Failed to retrieve user from KASM API")
				return fmt.Errorf("failed to retrieve user %s: %w", user.Username, err)
			}
			existingUser = nil
		}

		if existingUser != nil {
			log.Info().
				Str("username", user.Username).
				Str("user_id", existingUser.ID).
				Msg("User already exists in KASM")
			user.UserID = existingUser.ID
		} else {
			// Create user
			createdUser, err := api.CreateUser(api.TargetUser{
				Username: user.Username,
				ImageTag: user.AssignedContainerTag,
				// Add other necessary fields if required
			})
			if err != nil {
				log.Error().
					Err(err).
					Str("username", user.Username).
					Msg("Failed to create user via KASM API")
				return fmt.Errorf("failed to create user %s: %w", user.Username, err)
			}
			log.Info().
				Str("username", user.Username).
				Str("user_id", createdUser.ID).
				Msg("Successfully created user via KASM API")
			user.UserID = createdUser.ID
		}

		// Step 5: Add the user to the specified group
		log.Info().
			Str("username", user.Username).
			Str("role", user.Role).
			Msg("Adding user to the specified group via KASM API")

		groupID, err := api.GetGroupIDByName(user.Role)
		if err != nil {
			log.Error().
				Err(err).
				Str("role", user.Role).
				Msg("Failed to retrieve group ID from KASM API")
			return fmt.Errorf("failed to retrieve group ID for role %s: %w", user.Role, err)
		}

		err = api.AddUserToGroup(user.UserID, groupID)
		if err != nil {
			log.Error().
				Err(err).
				Str("user_id", user.UserID).
				Str("group_id", groupID).
				Msg("Failed to add user to group via KASM API")
			return fmt.Errorf("failed to add user %s to group %s: %w", user.Username, groupID, err)
		}
		log.Info().
			Str("user_id", user.UserID).
			Str("group_id", groupID).
			Msg("Successfully added user to group via KASM API")

		// Step 6: Update the YAML file with userId and KasmSessionOfContainer
		// Assume that KasmSessionOfContainer is retrieved or generated somehow
		// Replace the following line with actual logic to obtain the session ID
		kasmSessionOfContainer := "session-id-placeholder" // Replace with actual logic

		log.Info().
			Str("username", user.Username).
			Str("user_id", user.UserID).
			Str("kasm_session_of_container", kasmSessionOfContainer).
			Msg("Updating user configuration in YAML file")

		err = userParser.UpdateUserConfig(userConfigurationFilePath, user.Username, user.UserID, kasmSessionOfContainer)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", user.Username).
				Msg("Failed to update user configuration in YAML file")
			return fmt.Errorf("failed to update user %s configuration: %w", user.Username, err)
		}
		log.Info().
			Str("username", user.Username).
			Msg("Successfully updated user configuration in YAML file")
	}

	log.Info().
		Msg("Test environment creation completed successfully")

	return nil
}
