package dockercli

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
)

// PullImage pulls a Docker image from a remote repository.
func (dc *DockerClient) PullImage(ctx context.Context, imageName string) error {
	log.Info().Str("image", imageName).Msg("Starting Docker image pull")

	// Build the Docker pull command.
	pullCommand := []string{"docker", "pull", imageName}

	// Execute the command using the executor.
	output, err := dc.executor.Execute(ctx, pullCommand[0], pullCommand[1:]...)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to pull Docker image")
		return fmt.Errorf("failed to pull Docker image %s: %w", imageName, err)
	}

	log.Info().Str("image", imageName).Msg("Docker image pulled successfully")
	return nil
}
