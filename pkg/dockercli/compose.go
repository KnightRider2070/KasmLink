package dockercli

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// ComposeUp runs `docker-compose up` to bring up the services defined in a compose file.
func ComposeUp(composeFilePath string) error {
	log.Info().Str("compose_file", composeFilePath).Msg("Starting Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to run docker-compose up")
		return fmt.Errorf("failed to run docker-compose up: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose started services successfully")
	return nil
}

// ComposeDown runs `docker-compose down` to stop and remove services defined in a compose file.
func ComposeDown(composeFilePath string) error {
	log.Info().Str("compose_file", composeFilePath).Msg("Stopping and removing Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to run docker-compose down")
		return fmt.Errorf("failed to run docker-compose down: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose stopped and removed services successfully")
	return nil
}

// ComposeStart runs `docker-compose start` to start existing services defined in a compose file.
func ComposeStart(composeFilePath string) error {
	log.Info().Str("compose_file", composeFilePath).Msg("Starting existing Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to start Docker Compose services")
		return fmt.Errorf("failed to start services: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose started services successfully")
	return nil
}

// ComposeStop runs `docker-compose stop` to stop running services defined in a compose file.
func ComposeStop(composeFilePath string) error {
	log.Info().Str("compose_file", composeFilePath).Msg("Stopping Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "stop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to stop Docker Compose services")
		return fmt.Errorf("failed to stop services: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose stopped services successfully")
	return nil
}

// ComposeBuild runs `docker-compose build` to build the services defined in a compose file.
func ComposeBuild(composeFilePath string) error {
	log.Info().Str("compose_file", composeFilePath).Msg("Building Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to build Docker Compose services")
		return fmt.Errorf("failed to build services: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose built services successfully")
	return nil
}

// ComposeLogs runs `docker-compose logs` to show logs from services defined in a compose file.
func ComposeLogs(composeFilePath string) (string, error) {
	log.Info().Str("compose_file", composeFilePath).Msg("Fetching Docker Compose logs")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "logs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to fetch Docker Compose logs")
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	log.Info().Str("compose_file", composeFilePath).Msg("Docker Compose logs fetched successfully")
	return string(output), nil
}

// ListComposeServices lists all services and their corresponding container names and IDs in the Docker Compose setup.
func ListComposeServices(composeFilePath string) (map[string]map[string]string, error) {
	log.Info().Str("compose_file", composeFilePath).Msg("Listing Docker Compose services")

	cmd := exec.Command("docker-compose", "-f", composeFilePath, "ps", "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to list Docker Compose services")
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []struct {
		Service     string `json:"Service"`
		ContainerID string `json:"ID"`
		Name        string `json:"Name"`
	}

	err = json.Unmarshal(output, &services)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to parse JSON output from docker-compose ps")
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	serviceMap := make(map[string]map[string]string)
	for _, service := range services {
		serviceMap[service.Service] = map[string]string{
			"ContainerID":   service.ContainerID,
			"ContainerName": service.Name,
		}
	}

	log.Info().Str("compose_file", composeFilePath).Int("service_count", len(serviceMap)).Msg("Docker Compose services listed successfully")
	return serviceMap, nil
}

// InspectComposeService inspects the details of a service's corresponding container.
func InspectComposeService(containerID string) (string, error) {
	log.Info().Str("container_id", containerID).Msg("Inspecting Docker container")

	cmd := exec.Command("docker", "inspect", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to inspect Docker container")
		return "", fmt.Errorf("failed to inspect service: %w", err)
	}

	log.Info().Str("container_id", containerID).Msg("Docker container inspected successfully")
	return string(output), nil
}
