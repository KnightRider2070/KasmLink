package dockercli

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// ComposeUp runs `docker-compose up` to bring up the services defined in a compose file.
func ComposeUp(composeFilePath string) error {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run docker-compose up: %v - %s", err, string(output))
	}
	fmt.Printf("Docker Compose started services successfully.\n")
	return nil
}

// ComposeDown runs `docker-compose down` to stop and remove services defined in a compose file.
func ComposeDown(composeFilePath string) error {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run docker-compose down: %v - %s", err, string(output))
	}
	fmt.Printf("Docker Compose stopped and removed services successfully.\n")
	return nil
}

// ComposeStart runs `docker-compose start` to start existing services defined in a compose file.
func ComposeStart(composeFilePath string) error {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start services: %v - %s", err, string(output))
	}
	fmt.Printf("Docker Compose started services successfully.\n")
	return nil
}

// ComposeStop runs `docker-compose stop` to stop running services defined in a compose file.
func ComposeStop(composeFilePath string) error {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "stop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop services: %v - %s", err, string(output))
	}
	fmt.Printf("Docker Compose stopped services successfully.\n")
	return nil
}

// ComposeBuild runs `docker-compose build` to build the services defined in a compose file.
func ComposeBuild(composeFilePath string) error {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build services: %v - %s", err, string(output))
	}
	fmt.Printf("Docker Compose built services successfully.\n")
	return nil
}

// ComposeLogs runs `docker-compose logs` to show logs from services defined in a compose file.
func ComposeLogs(composeFilePath string) (string, error) {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "logs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v - %s", err, string(output))
	}
	return string(output), nil
}

// ListComposeServices lists all services and their corresponding container names and IDs in the Docker Compose setup.
func ListComposeServices(composeFilePath string) (map[string]map[string]string, error) {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "ps", "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v - %s", err, string(output))
	}

	// Parse the JSON output to map each service to its container name and ID
	var services []struct {
		Service     string `json:"Service"`
		ContainerID string `json:"ID"`
		Name        string `json:"Name"`
	}

	err = json.Unmarshal(output, &services)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %v", err)
	}

	// Create a map of services and their corresponding container IDs and names
	serviceMap := make(map[string]map[string]string)
	for _, service := range services {
		serviceMap[service.Service] = map[string]string{
			"ContainerID":   service.ContainerID,
			"ContainerName": service.Name,
		}
	}

	return serviceMap, nil
}

// InspectComposeService inspects the details of a service's corresponding container.
func InspectComposeService(containerID string) (string, error) {
	cmd := exec.Command("docker", "inspect", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to inspect service: %v - %s", err, string(output))
	}

	return string(output), nil
}
