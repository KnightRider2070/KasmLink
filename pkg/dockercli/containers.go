package dockercli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// StartContainer starts a Docker container by name or ID.
func StartContainer(containerID string) error {
	log.Info().Str("container_id", containerID).Msg("Starting Docker container")
	cmd := exec.Command("docker", "start", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to start container")
		return fmt.Errorf("failed to start container: %w", err)
	}
	log.Info().Str("container_id", containerID).Msg("Container started successfully")
	return nil
}

// StopContainer stops a Docker container by name or ID.
func StopContainer(containerID string) error {
	log.Info().Str("container_id", containerID).Msg("Stopping Docker container")
	cmd := exec.Command("docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to stop container")
		return fmt.Errorf("failed to stop container: %w", err)
	}
	log.Info().Str("container_id", containerID).Msg("Container stopped successfully")
	return nil
}

// RemoveContainer removes a Docker container by name or ID.
func RemoveContainer(containerID string) error {
	log.Info().Str("container_id", containerID).Msg("Removing Docker container")
	cmd := exec.Command("docker", "rm", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to remove container")
		return fmt.Errorf("failed to remove container: %w", err)
	}
	log.Info().Str("container_id", containerID).Msg("Container removed successfully")
	return nil
}

// RestartContainer restarts a Docker container by name or ID.
func RestartContainer(containerID string) error {
	log.Info().Str("container_id", containerID).Msg("Restarting Docker container")
	cmd := exec.Command("docker", "restart", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to restart container")
		return fmt.Errorf("failed to restart container: %w", err)
	}
	log.Info().Str("container_id", containerID).Msg("Container restarted successfully")
	return nil
}

// CopyFromContainer copies files from a Docker container to the host.
func CopyFromContainer(containerID, containerPath, hostPath string) error {
	log.Info().Str("container_id", containerID).Str("container_path", containerPath).Str("host_path", hostPath).Msg("Copying files from container")
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", containerID, containerPath), hostPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to copy files from container")
		return fmt.Errorf("failed to copy files from container: %w", err)
	}
	log.Info().Str("container_id", containerID).Str("container_path", containerPath).Str("host_path", hostPath).Msg("Files copied successfully from container to host")
	return nil
}

// CopyToContainer copies files from the host to a Docker container.
func CopyToContainer(hostPath, containerID, containerPath string) error {
	log.Info().Str("host_path", hostPath).Str("container_id", containerID).Str("container_path", containerPath).Msg("Copying files to container")
	cmd := exec.Command("docker", "cp", hostPath, fmt.Sprintf("%s:%s", containerID, containerPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to copy files to container")
		return fmt.Errorf("failed to copy files to container: %w", err)
	}
	log.Info().Str("host_path", hostPath).Str("container_id", containerID).Str("container_path", containerPath).Msg("Files copied successfully from host to container")
	return nil
}

// ListContainers lists all running Docker containers and returns their names and IDs.
func ListContainers() (map[string]string, error) {
	log.Info().Msg("Listing all running Docker containers")
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to list containers")
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containers := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				containers[parts[0]] = parts[1]
			}
		}
	}

	log.Info().Int("container_count", len(containers)).Msg("Successfully listed running containers")
	return containers, nil
}

// GetContainerIDByName retrieves the container ID by its name.
func GetContainerIDByName(containerName string) (string, error) {
	log.Info().Str("container_name", containerName).Msg("Getting container ID by name")
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.ID}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to get container ID by name")
		return "", fmt.Errorf("failed to get container ID by name: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		log.Warn().Str("container_name", containerName).Msg("No container found with specified name")
		return "", fmt.Errorf("no container found with the name %s", containerName)
	}

	log.Info().Str("container_name", containerName).Str("container_id", containerID).Msg("Container ID retrieved successfully by name")
	return containerID, nil
}

// GetContainerNameByID retrieves the container name by its ID.
func GetContainerNameByID(containerID string) (string, error) {
	log.Info().Str("container_id", containerID).Msg("Getting container name by ID")
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("id=%s", containerID), "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to get container name by ID")
		return "", fmt.Errorf("failed to get container name by ID: %w", err)
	}

	containerName := strings.TrimSpace(string(output))
	if containerName == "" {
		log.Warn().Str("container_id", containerID).Msg("No container found with specified ID")
		return "", fmt.Errorf("no container found with the ID %s", containerID)
	}

	log.Info().Str("container_id", containerID).Str("container_name", containerName).Msg("Container name retrieved successfully by ID")
	return containerName, nil
}

// ExecInContainer executes a command inside a running Docker container.
func ExecInContainer(containerID string, command []string) (string, error) {
	log.Info().Str("container_id", containerID).Strs("command", command).Msg("Executing command in container")
	args := append([]string{"exec", containerID}, command...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("container_id", containerID).Str("output", string(output)).Msg("Failed to execute command in container")
		return "", fmt.Errorf("failed to execute command in container %s: %w", containerID, err)
	}

	log.Info().Str("container_id", containerID).Msg("Command executed successfully in container")
	return strings.TrimSpace(string(output)), nil
}
