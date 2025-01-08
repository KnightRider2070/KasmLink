package dockerRegistry

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/shadowscp"
	sshmanager "kasmlink/pkg/shadowssh"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RegistryConfig defines the configuration for a Docker registry.
type RegistryConfig struct {
	ContainerName string
	RegistryImage string
	URL           string
	Port          int
	User          string
	Password      string
}

// NewRegistryConfig creates a new RegistryConfig with default values.
func NewRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		ContainerName: "registry",
		RegistryImage: "registry:latest",
		URL:           "http://localhost:5000",
		Port:          5000,
		User:          "neo",
		Password:      "redpill42",
	}
}

// SetupRegistry sets up a Docker registry on a remote server via SSH.
func SetupRegistry(ctx context.Context, sshConfig *sshmanager.Config, registryConfig *RegistryConfig, dockerImagesTarPath string) error {
	client, err := sshmanager.NewClient(ctx, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to target node via SSH: %w", err)
	}
	defer client.Close()

	// Check if the registry image exists on the remote server.
	imageExists, err := checkRemoteImageExists(ctx, client, registryConfig.RegistryImage)
	if err != nil {
		return fmt.Errorf("failed to check remote images: %w", err)
	}

	if !imageExists {
		log.Info().Str("image", registryConfig.RegistryImage).Msg("Registry image not found on remote server. Preparing to transfer.")
		err := handleImageTransfer(ctx, client, registryConfig.RegistryImage, dockerImagesTarPath, sshConfig)
		if err != nil {
			return fmt.Errorf("failed to handle image transfer: %w", err)
		}
	}

	// Start the registry container.
	err = startRegistryContainer(ctx, client, registryConfig)
	if err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	log.Info().Msg("Registry container setup successfully")
	log.Info().Msgf("Registry URL: http://%s:%d", sshConfig.Host, registryConfig.Port)
	return nil
}

// checkRemoteImageExists checks if a Docker image exists on the remote server.
func checkRemoteImageExists(ctx context.Context, client *sshmanager.Client, image string) (bool, error) {
	log.Debug().Msg("Checking if registry image exists on remote server.")
	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	output, err := client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return false, fmt.Errorf("failed to execute image check command: %w", err)
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == image {
			return true, nil
		}
	}
	return false, nil
}

// handleImageTransfer ensures the Docker image is available on the remote server.
func handleImageTransfer(ctx context.Context, client *sshmanager.Client, image, dockerImagesTarPath string, sshConfig *sshmanager.Config) error {
	localTarPath := filepath.Join(dockerImagesTarPath, fmt.Sprintf("%s.tar", strings.ReplaceAll(image, ":", "_")))
	remoteTarPath := fmt.Sprintf("/tmp/%s.tar", strings.ReplaceAll(image, ":", "_"))

	// Check if the tar file exists locally; if not, pull and save the image.
	if _, err := os.Stat(localTarPath); os.IsNotExist(err) {
		log.Info().Str("image", image).Msg("Local tar file not found. Pulling and saving image locally.")
		if err := pullAndSaveImageLocally(image, localTarPath); err != nil {
			return fmt.Errorf("failed to pull and save image locally: %w", err)
		}
	}

	// Transfer the tar file to the remote server.
	log.Info().Msg("Transferring image tar to remote server.")
	err := shadowscp.CopyFileToRemote(ctx, localTarPath, "/tmp", sshConfig)
	if err != nil {
		return fmt.Errorf("failed to upload tar file to remote node: %w", err)
	}

	// Load the image on the remote server.
	log.Info().Msg("Loading tar file on remote server.")
	cmd := fmt.Sprintf("docker load < %s", remoteTarPath)
	_, err = client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load image on remote node: %w", err)
	}

	return nil
}

// pullAndSaveImageLocally pulls a Docker image and saves it as a tar file.
func pullAndSaveImageLocally(image, tarPath string) error {
	log.Info().Str("image", image).Msg("Pulling image locally.")
	cmd := exec.Command("docker", "pull", image)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull image: %s, %w", string(output), err)
	}

	log.Info().Str("tarPath", tarPath).Msg("Saving image to tar file.")
	cmd = exec.Command("docker", "save", "-o", tarPath, image)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to save image to tar: %s, %w", string(output), err)
	}
	return nil
}

// startRegistryContainer starts the Docker registry container on the remote server.
func startRegistryContainer(ctx context.Context, client *sshmanager.Client, registryConfig *RegistryConfig) error {
	log.Info().Str("container", registryConfig.ContainerName).Msg("Starting registry container on remote server.")
	cmd := fmt.Sprintf(
		"docker run -d --name %s -p %d:%d -e REGISTRY_HTTP_ADDR=0.0.0.0:%d %s",
		registryConfig.ContainerName,
		registryConfig.Port,
		registryConfig.Port,
		registryConfig.Port,
		registryConfig.RegistryImage,
	)
	_, err := client.ExecuteCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}
	return nil
}
