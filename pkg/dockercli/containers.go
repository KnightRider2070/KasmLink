package dockercli

import (
	"fmt"
	"os/exec"
	"strings"
)

// StartContainer starts a Docker container by name or ID.
func StartContainer(containerID string) error {
	cmd := exec.Command("docker", "start", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container: %v - %s", err, string(output))
	}
	fmt.Printf("Container %s started successfully.\n", containerID)
	return nil
}

// StopContainer stops a Docker container by name or ID.
func StopContainer(containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop container: %v - %s", err, string(output))
	}
	fmt.Printf("Container %s stopped successfully.\n", containerID)
	return nil
}

// RemoveContainer removes a Docker container by name or ID.
func RemoveContainer(containerID string) error {
	cmd := exec.Command("docker", "rm", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove container: %v - %s", err, string(output))
	}
	fmt.Printf("Container %s removed successfully.\n", containerID)
	return nil
}

// RestartContainer restarts a Docker container by name or ID.
func RestartContainer(containerID string) error {
	cmd := exec.Command("docker", "restart", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart container: %v - %s", err, string(output))
	}
	fmt.Printf("Container %s restarted successfully.\n", containerID)
	return nil
}

// CopyFromContainer copies files from a Docker container to the host.
func CopyFromContainer(containerID, containerPath, hostPath string) error {
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:%s", containerID, containerPath), hostPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy files from container: %v - %s", err, string(output))
	}
	fmt.Printf("Copied files from container %s:%s to host %s.\n", containerID, containerPath, hostPath)
	return nil
}

// CopyToContainer copies files from the host to a Docker container.
func CopyToContainer(hostPath, containerID, containerPath string) error {
	cmd := exec.Command("docker", "cp", hostPath, fmt.Sprintf("%s:%s", containerID, containerPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy files to container: %v - %s", err, string(output))
	}
	fmt.Printf("Copied files from host %s to container %s:%s.\n", hostPath, containerID, containerPath)
	return nil
}

// ListContainers lists all running Docker containers and returns their names and IDs.
func ListContainers() (map[string]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}} {{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v - %s", err, string(output))
	}

	// Parse the output into a map of container IDs to container names
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

	return containers, nil
}

// GetContainerIDByName retrieves the container ID by its name.
func GetContainerIDByName(containerName string) (string, error) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.ID}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container ID by name: %v - %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", fmt.Errorf("no container found with the name %s", containerName)
	}

	return containerID, nil
}

// GetContainerNameByID retrieves the container name by its ID.
func GetContainerNameByID(containerID string) (string, error) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("id=%s", containerID), "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container name by ID: %v - %s", err, string(output))
	}

	containerName := strings.TrimSpace(string(output))
	if containerName == "" {
		return "", fmt.Errorf("no container found with the ID %s", containerID)
	}

	return containerName, nil
}

// ExecInContainer executes a command inside a running Docker container.
func ExecInContainer(containerID string, command []string) (string, error) {
	// Build the exec command
	args := append([]string{"exec", containerID}, command...)

	// Execute the docker exec command
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command in container %s: %v - %s", containerID, err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
