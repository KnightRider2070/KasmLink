package dockercli

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

// CreateVolume creates a Docker volume with the specified name.
func CreateVolume(volumeName string) error {
	log.Info().Str("volume_name", volumeName).Msg("Creating Docker volume")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := executeDockerCommand(ctx, 3, "docker", "volume", "create", volumeName)
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Msg("Failed to create Docker volume")
		return fmt.Errorf("failed to create volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume created successfully")
	return nil
}

// InspectVolume inspects a Docker volume by name.
func InspectVolume(volumeName string) (string, error) {
	log.Info().Str("volume_name", volumeName).Msg("Inspecting Docker volume")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := executeDockerCommand(ctx, 3, "docker", "volume", "inspect", volumeName)
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Msg("Failed to inspect Docker volume")
		return "", fmt.Errorf("failed to inspect volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume inspected successfully")
	return string(output), nil
}

// RemoveVolume removes a Docker volume by name.
func RemoveVolume(volumeName string) error {
	log.Info().Str("volume_name", volumeName).Msg("Removing Docker volume")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := executeDockerCommand(ctx, 3, "docker", "volume", "rm", volumeName)
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Msg("Failed to remove Docker volume")
		return fmt.Errorf("failed to remove volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume removed successfully")
	return nil
}
