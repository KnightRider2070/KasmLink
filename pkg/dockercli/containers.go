package dockercli

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// StartContainer starts a Docker container by name or ID.
func StartContainer(containerID string, retries int, timeout time.Duration) error {
	log.Info().Str("container_id", containerID).Msg("Starting Docker container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "start", containerID)
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to start container")
		return fmt.Errorf("failed to start container: %w", err)
	}

	log.Info().Str("container_id", containerID).Msg("Container started successfully")
	return nil
}

// StopContainer stops a Docker container by name or ID.
func StopContainer(containerID string, retries int, timeout time.Duration) error {
	log.Info().Str("container_id", containerID).Msg("Stopping Docker container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "stop", containerID)
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to stop container")
		return fmt.Errorf("failed to stop container: %w", err)
	}

	log.Info().Str("container_id", containerID).Msg("Container stopped successfully")
	return nil
}

// RemoveContainer removes a Docker container by name or ID.
func RemoveContainer(containerID string, retries int, timeout time.Duration) error {
	log.Info().Str("container_id", containerID).Msg("Removing Docker container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "rm", containerID)
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to remove container")
		return fmt.Errorf("failed to remove container: %w", err)
	}

	log.Info().Str("container_id", containerID).Msg("Container removed successfully")
	return nil
}

// RestartContainer restarts a Docker container by name or ID.
func RestartContainer(containerID string, retries int, timeout time.Duration) error {
	log.Info().Str("container_id", containerID).Msg("Restarting Docker container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "restart", containerID)
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to restart container")
		return fmt.Errorf("failed to restart container: %w", err)
	}

	log.Info().Str("container_id", containerID).Msg("Container restarted successfully")
	return nil
}

// CopyFromContainer copies files from a Docker container to the host.
func CopyFromContainer(containerID, containerPath, hostPath string, retries int, timeout time.Duration) error {
	log.Info().Str("container_id", containerID).Str("container_path", containerPath).Str("host_path", hostPath).Msg("Copying files from container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "cp", fmt.Sprintf("%s:%s", containerID, containerPath), hostPath)
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to copy files from container")
		return fmt.Errorf("failed to copy files from container: %w", err)
	}

	log.Info().Str("container_id", containerID).Str("container_path", containerPath).Str("host_path", hostPath).Msg("Files copied successfully from container to host")
	return nil
}

// CopyToContainer copies files from the host to a Docker container.
func CopyToContainer(hostPath, containerID, containerPath string, retries int, timeout time.Duration) error {
	log.Info().Str("host_path", hostPath).Str("container_id", containerID).Str("container_path", containerPath).Msg("Copying files to container")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retries, "docker", "cp", hostPath, fmt.Sprintf("%s:%s", containerID, containerPath))
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to copy files to container")
		return fmt.Errorf("failed to copy files to container: %w", err)
	}

	log.Info().Str("host_path", hostPath).Str("container_id", containerID).Str("container_path", containerPath).Msg("Files copied successfully from host to container")
	return nil
}
