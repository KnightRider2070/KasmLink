package dockercli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
)

// CreateVolume creates a Docker volume with the specified name.
func CreateVolume(volumeName string) error {
	log.Info().Str("volume_name", volumeName).Msg("Creating Docker volume")

	cmd := exec.Command("docker", "volume", "create", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Str("output", string(output)).Msg("Failed to create Docker volume")
		return fmt.Errorf("failed to create volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume created successfully")
	return nil
}

// InspectVolume inspects a Docker volume by name.
func InspectVolume(volumeName string) (string, error) {
	log.Info().Str("volume_name", volumeName).Msg("Inspecting Docker volume")

	cmd := exec.Command("docker", "volume", "inspect", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Str("output", string(output)).Msg("Failed to inspect Docker volume")
		return "", fmt.Errorf("failed to inspect volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume inspected successfully")
	return string(output), nil
}

// RemoveVolume removes a Docker volume by name.
func RemoveVolume(volumeName string) error {
	log.Info().Str("volume_name", volumeName).Msg("Removing Docker volume")

	cmd := exec.Command("docker", "volume", "rm", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("volume_name", volumeName).Str("output", string(output)).Msg("Failed to remove Docker volume")
		return fmt.Errorf("failed to remove volume: %w", err)
	}

	log.Info().Str("volume_name", volumeName).Msg("Docker volume removed successfully")
	return nil
}
