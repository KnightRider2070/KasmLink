package dockerRegistry

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	shadowscp "kasmlink/pkg/scp"
	sshmanager "kasmlink/pkg/sshmanager"
	"os"
	"os/exec"
	"strings"
)

type RegistryConfig struct {
	ContainerName       string
	RegistryImageToPull string
	URL                 string
	Port                int
	User                string
	Password            string
}

// NewRegistryConfig creates a new RegistryConfig with default values.
func NewRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		ContainerName:       "registry",
		RegistryImageToPull: "registry:latest",
		URL:                 "http://localhost:5000",
		Port:                5000,
		User:                "neo",
		Password:            "redpill42",
	}
}

func SetupRegistry(ctx context.Context, targetSSH *sshmanager.SSHConfig, registryConfig *RegistryConfig, dockerImagesTarPath string) error {
	client, err := sshmanager.NewSSHClient(ctx, targetSSH)
	if err != nil {
		return fmt.Errorf("failed to connect to target node via SSH: %w", err)
	}
	defer client.Close()

	// Check if image exists on the remote node
	imageCheckCommand := "docker images --format '{{.Repository}}:{{.Tag}}'"
	output, err := client.ExecuteCommand(ctx, imageCheckCommand)
	if err != nil {
		return fmt.Errorf("failed to check images on remote node: %w", err)
	}

	expectedImage := fmt.Sprintf("%s:latest", registryConfig.RegistryImageToPull)
	imageExists := false
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == expectedImage {
			imageExists = true
			break
		}
	}

	if !imageExists {
		localTarPath := fmt.Sprintf("%s/%s.tar", dockerImagesTarPath, strings.ReplaceAll(registryConfig.RegistryImageToPull, ":", "_"))
		remoteTarPath := fmt.Sprintf("/tmp/%s.tar", strings.ReplaceAll(registryConfig.RegistryImageToPull, ":", "_"))

		// Attempt to find or pull the image locally and upload
		if _, err := os.Stat(localTarPath); err == nil {
			log.Info().Msg("Local image tar found. Uploading to remote node...")
			if err := shadowscp.ShadowCopyFile(ctx, localTarPath, "/tmp", targetSSH); err != nil {
				return fmt.Errorf("failed to upload tar file to remote node: %w", err)
			}
		} else {
			log.Info().Msg("Local tar not found. Pulling image locally...")
			pullCommand := exec.Command("docker", "pull", registryConfig.RegistryImageToPull)
			if output, err := pullCommand.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to pull image locally: %s, %w", string(output), err)
			}

			log.Info().Msg("Exporting image to tar...")
			exportCommand := exec.Command("docker", "save", "-o", localTarPath, registryConfig.RegistryImageToPull)
			if output, err := exportCommand.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to export image to tar file: %s, %w", string(output), err)
			}

			if err := shadowscp.ShadowCopyFile(ctx, localTarPath, "/tmp", targetSSH); err != nil {
				return fmt.Errorf("failed to upload tar file to remote node: %w", err)
			}
		}

		log.Info().Msg("Loading tar file on remote node...")
		loadCommand := fmt.Sprintf("docker load < %s", remoteTarPath)
		if _, err := client.ExecuteCommand(ctx, loadCommand); err != nil {
			return fmt.Errorf("failed to load image on remote node: %w", err)
		}
	}

	// Start the registry container
	log.Info().Msg("Starting registry container...")
	runCommand := fmt.Sprintf("docker run -d --name %s -p %d:%d -e REGISTRY_HTTP_ADDR=0.0.0.0:%d %s",
		registryConfig.ContainerName,
		registryConfig.Port,
		registryConfig.Port,
		registryConfig.Port,
		registryConfig.RegistryImageToPull,
	)

	if _, err := client.ExecuteCommand(ctx, runCommand); err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	log.Info().Msg("Registry container setup successfully")
	return nil
}
