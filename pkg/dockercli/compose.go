package dockercli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

// ComposeUp runs `docker-compose up` to bring up the services defined in a compose file.
func ComposeUp(composeFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Starting Docker Compose services")
	_, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "up", "-d")
	return err
}

// ComposeDown runs `docker-compose down` to stop and remove services defined in a compose file.
func ComposeDown(composeFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Stopping and removing Docker Compose services")
	_, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "down")
	return err
}

// ComposeStart runs `docker-compose start` to start existing services defined in a compose file.
func ComposeStart(composeFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Starting existing Docker Compose services")
	_, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "start")
	return err
}

// ComposeStop runs `docker-compose stop` to stop running services defined in a compose file.
func ComposeStop(composeFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Stopping Docker Compose services")
	_, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "stop")
	return err
}

// ComposeBuild runs `docker-compose build` to build the services defined in a compose file.
func ComposeBuild(composeFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Building Docker Compose services")
	_, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "build")
	return err
}

// ComposeLogs runs `docker-compose logs` to show logs from services defined in a compose file.
func ComposeLogs(composeFilePath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Fetching Docker Compose logs")
	output, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "logs")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ListComposeServices lists all services and their corresponding container names and IDs in the Docker Compose setup.
func ListComposeServices(composeFilePath string) (map[string]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("compose_file", composeFilePath).Msg("Listing Docker Compose services")
	output, err := executeDockerCommand(ctx, 3, "docker-compose", "-f", composeFilePath, "ps", "--format", "json")
	if err != nil {
		return nil, err
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	log.Info().Str("container_id", containerID).Msg("Inspecting Docker container")
	output, err := executeDockerCommand(ctx, 3, "docker", "inspect", containerID)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
