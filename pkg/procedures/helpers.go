package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	shadowssh "kasmlink/pkg/sshmanager"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
	"strings"
	"time"
)

// checkRemoteImages checks which Docker images are missing on the remote node.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - client: SSHClient for executing commands on the remote node.
// - images: List of Docker image names to check.
// Returns:
// - List of missing Docker image names.
// - An error if the check fails.
func checkRemoteImages(ctx context.Context, client *shadowssh.SSHClient, images []string) ([]string, error) {
	log.Debug().
		Msg("Executing remote Docker images command to list available images")

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	// Execute the command with a timeout for logging
	output, err := client.ExecuteCommandWithOutput(ctx, cmd, 30*time.Second)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", cmd).
			Msg("Failed to execute remote Docker images command")
		return nil, fmt.Errorf("failed to execute remote Docker images command: %w", err)
	}

	remoteImages := strings.Split(output, "\n")
	missing := []string{}

	// Create a set of remote images for efficient lookup
	imageSet := make(map[string]struct{})
	for _, img := range remoteImages {
		trimmedImg := strings.TrimSpace(img)
		if trimmedImg != "" {
			imageSet[trimmedImg] = struct{}{}
		}
	}

	// Identify missing images
	for _, img := range images {
		if _, exists := imageSet[img]; !exists {
			missing = append(missing, img)
			log.Debug().
				Str("image", img).
				Msg("Image is missing on remote node")
		}
	}

	return missing, nil
}

import (
"context"
"fmt"

"kasmlink/pkg/webApi"

"github.com/rs/zerolog/log"
)

// createOrGetUser creates a new user via KASM API or retrieves the existing user's ID.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - api: Pointer to KasmAPI instance for API interactions.
// - username: Username to create or retrieve.
// - imageTag: Docker image tag assigned to the user.
// Returns:
// - userID: The ID of the created or existing user.
// - An error if the operation fails.
func createOrGetUser(ctx context.Context, api *webApi.KasmAPI, user userParser.UserDetails) (string, error) {
	log.Info().
		Str("username", user.TargetUser.Username).
		Msg("Attempting to retrieve or create user via KASM API")

	// Step 1: Try to retrieve the user by username
	userExisting, err := api.GetUser(ctx, "", user.TargetUser.Username)
	if err != nil {
		// Assuming that an error containing "not found" indicates the user does not exist
		if userExisting != nil {
			log.Info().
				Str("username", user.TargetUser.Username).
				Msg("User already exists in KASM API")
			return userExisting.UserID, nil
		}

		// User does not exist; proceed to create
		log.Info().
			Str("username", user.TargetUser.Username).
			Msg("User not found. Proceeding to create a new user.")

		// Define the target user details
		targetUser := webApi.TargetUser{
			Username: user.TargetUser.Username,
			imageTag: imageTag,
			// Populate other necessary fields as required by your API
		}

		// Step 2: Create the user via the API
		createdUser, err := api.CreateUser(ctx, targetUser)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", user.TargetUser.Username).
				Msg("Failed to create user via KASM API")
			return "", fmt.Errorf("failed to create user %s: %w", user.TargetUser.Username, err)
		}

		log.Info().
			Str("username", createdUser.Username).
			Str("user_id", createdUser.UserID).
			Msg("User created successfully via KASM API")
		return createdUser.UserID, nil
	}

	// User exists; return the existing user ID
	log.Info().
		Str("username", user.TargetUser.Username).
		Str("user_id", user.UserID).
		Msg("User already exists in KASM API")
	return user.UserID, nil
}

// isUserNotFoundError checks if the error returned by GetUser indicates that the user was not found.
// You may need to adjust this function based on how your API communicates "not found" errors.
func isUserNotFoundError(err error) bool {
	// Example: Check if the error message contains "not found"
	return fmt.Sprintf("%v", err) == "user not found" // Adjust condition as per your API's error messages
}