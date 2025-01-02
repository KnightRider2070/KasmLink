package internal

import (
	"context"
	"errors"
	"fmt"
	"kasmlink/pkg/shadowscp"
	"os"
	"path/filepath"
	"sync"

	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockercompose"
	"kasmlink/pkg/shadowssh"

	"github.com/rs/zerolog/log"
)

// DeployBackendServices deploys backend services using Docker Compose on a remote server.
func DeployBackendServices(ctx context.Context, backendComposePath string, sshConfig *shadowssh.Config, dockerClient *dockercli.DockerClient) error {
	if err := validateFileExists(backendComposePath); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if err := validateSSHConfig(sshConfig); err != nil {
		return fmt.Errorf("invalid SSH configuration: %w", err)
	}

	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer closeSSHClient(sshClient)

	composeFile, err := dockercompose.LoadAndParseComposeFile(backendComposePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Docker Compose file")
		return fmt.Errorf("failed to load Docker Compose file: %w", err)
	}

	if err := dockercompose.ValidateDockerCompose(composeFile); err != nil {
		log.Error().Err(err).Msg("Failed to validate Docker Compose file")
		return fmt.Errorf("failed to validate Docker Compose file: %w", err)
	}

	requiredImages := extractRequiredImages(&composeFile)
	remoteImages, err := dockerClient.ListImages(ctx, dockercli.ListImagesOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list remote Docker images")
		return fmt.Errorf("failed to list remote Docker images: %w", err)
	}

	missingImages := findMissingImages(requiredImages, remoteImages)
	if len(missingImages) > 0 {
		if err := transferMissingImagesConcurrently(ctx, dockerClient, missingImages, sshConfig); err != nil {
			return fmt.Errorf("error during image transfer: %w", err)
		}
	}

	if err := deployCompose(ctx, backendComposePath, sshClient, sshConfig); err != nil {
		return fmt.Errorf("deployment error: %w", err)
	}

	log.Info().Msg("Deployment completed successfully")
	return nil
}

func validateSSHConfig(sshConfig *shadowssh.Config) error {
	if sshConfig == nil || sshConfig.Host == "" || sshConfig.Username == "" {
		return errors.New("missing required SSH configuration fields")
	}
	return nil
}

func closeSSHClient(client *shadowssh.Client) {
	if err := client.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close SSH connection gracefully")
	}
}

func extractRequiredImages(composeFile *dockercompose.DockerCompose) []dockercli.DockerImage {
	var images []dockercli.DockerImage
	for _, service := range composeFile.Services {
		images = append(images, dockercli.DockerImage{
			Repository: service.Image,
		})
	}
	return images
}

func transferMissingImagesConcurrently(ctx context.Context, dockerClient *dockercli.DockerClient, missingImages []dockercli.DockerImage, sshConfig *shadowssh.Config) error {
	log.Info().Int("count", len(missingImages)).Msg("Starting concurrent transfer of missing images")

	var wg sync.WaitGroup
	errCh := make(chan error, len(missingImages))
	for _, image := range missingImages {
		wg.Add(1)
		go func(img dockercli.DockerImage) {
			defer wg.Done()
			log.Info().Str("image", img.Repository).Msg("Transferring image")
			if err := dockerClient.TransferImage(ctx, img.Repository, sshConfig); err != nil {
				log.Error().Err(err).Str("image", img.Repository).Msg("Failed to transfer image")
				errCh <- fmt.Errorf("failed to transfer image %s: %w", img.Repository, err)
			}
		}(image)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}
	log.Info().Msg("All images transferred successfully")
	return nil
}

func validateFileExists(path string) error {
	log.Info().Str("path", path).Msg("Validating file existence")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error().Err(err).Msgf("File does not exist: %s", path)
		return fmt.Errorf("file does not exist: %s", path)
	}
	log.Info().Str("path", path).Msg("File exists")
	return nil
}

func deployCompose(ctx context.Context, composePath string, sshClient *shadowssh.Client, sshConfig *shadowssh.Config) error {
	remoteDir := "/composefiles"
	log.Info().Str("compose_path", composePath).Msg("Transferring Docker Compose file to remote host")

	remotePath := filepath.Join(remoteDir, filepath.Base(composePath))
	if err := shadowscp.CopyFileToRemote(ctx, composePath, remotePath, sshConfig); err != nil {
		log.Error().Err(err).Str("remote_path", remotePath).Msg("Failed to transfer Docker Compose file")
		return fmt.Errorf("failed to transfer Docker Compose file: %w", err)
	}

	log.Info().Str("remote_path", remotePath).Msg("Executing Docker Compose up on remote host")
	command := fmt.Sprintf("cd %s && docker compose up -d", remoteDir)
	output, err := sshClient.ExecuteCommand(ctx, command)
	if err != nil {
		log.Error().Err(err).Str("output", output).Msg("Failed to execute Docker Compose up")
		return fmt.Errorf("failed to execute Docker Compose up: %w", err)
	}
	log.Info().Msg("Docker Compose up executed successfully")
	return nil
}

func findMissingImages(local, remote []dockercli.DockerImage) []dockercli.DockerImage {
	remoteMap := make(map[string]struct{}, len(remote))
	for _, img := range remote {
		remoteMap[fmt.Sprintf("%s:%s", img.Repository, img.Tag)] = struct{}{}
	}

	var missing []dockercli.DockerImage
	for _, img := range local {
		if _, exists := remoteMap[fmt.Sprintf("%s:%s", img.Repository, img.Tag)]; !exists {
			missing = append(missing, img)
		}
	}
	return missing
}
