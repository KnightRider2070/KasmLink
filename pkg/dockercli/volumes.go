package dockercli

import (
	"fmt"
	"os/exec"
)

// CreateVolume creates a Docker volume with the specified name.
func CreateVolume(volumeName string) error {
	cmd := exec.Command("docker", "volume", "create", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create volume: %v - %s", err, string(output))
	}
	fmt.Printf("Volume %s created successfully.\n", volumeName)
	return nil
}

// InspectVolume inspects a Docker volume by name.
func InspectVolume(volumeName string) (string, error) {
	cmd := exec.Command("docker", "volume", "inspect", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to inspect volume: %v - %s", err, string(output))
	}
	return string(output), nil
}

// RemoveVolume removes a Docker volume by name.
func RemoveVolume(volumeName string) error {
	cmd := exec.Command("docker", "volume", "rm", volumeName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove volume: %v - %s", err, string(output))
	}
	fmt.Printf("Volume %s removed successfully.\n", volumeName)
	return nil
}
