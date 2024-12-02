package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	shadowssh "kasmlink/pkg/sshmanager"
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
