package internal

import (
	"context"
	"fmt"
	"os"

	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowscp"
	"kasmlink/pkg/shadowssh"

	"github.com/rs/zerolog/log"
)

// DeployBackendServices deploys backend services based on the provided Docker Compose file and SSH configuration.
func DeployBackendServices(ctx context.Context, backendComposePath string, sshConfig *shadowssh.Config, dockerClient *dockercli.DockerClient) error {
	// Validate that the Docker Compose file exists.
	if err := validateFileExists(backendComposePath); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Establish an SSH connection.
	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := sshClient.Close(); cerr != nil {
			log.Warn().Err(cerr).Msg("Failed to close SSH connection gracefully")
		}
	}()

	// List missing images on the remote server.
	missingImages, err := listMissingImages(ctx, dockerClient, sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list missing Docker images")
		return fmt.Errorf("failed to list missing Docker images: %w", err)
	}

	// Transfer missing images to the remote server.
	if len(missingImages) > 0 {
		if err := transferMissingImages(ctx, dockerClient, missingImages, sshConfig); err != nil {
			return fmt.Errorf("error during image transfer: %w", err)
		}
	}

	// Deploy Docker Compose services on the remote server.
	if err := deployCompose(ctx, backendComposePath, sshClient, sshConfig); err != nil {
		return fmt.Errorf("deployment error: %w", err)
	}

	log.Info().Msg("Deployment completed successfully")
	return nil
}

// validateFileExists checks if the given file exists.
func validateFileExists(path string) error {
	log.Info().Str("path", path).Msg("Validating file existence")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error().Err(err).Msgf("File does not exist: %s", path)
		return fmt.Errorf("file does not exist: %s", path)
	}
	return nil
}

// listMissingImages determines which Docker images are missing on the remote server.
func listMissingImages(ctx context.Context, dockerClient *dockercli.DockerClient, sshConfig *shadowssh.Config) ([]dockercli.DockerImage, error) {
	log.Info().Msg("Checking for missing Docker images on remote host")

	// List images locally.
	localImages, err := dockerClient.ListImages(ctx, dockercli.ListImagesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list local Docker images: %w", err)
	}

	// List images on remote server using SSH.
	remoteExecutor := dockercli.NewSSHCommandExecutor(sshConfig)
	remoteDockerClient := dockercli.NewDockerClient(remoteExecutor, &dockercli.LocalFileSystem{})
	remoteImages, err := remoteDockerClient.ListImages(ctx, dockercli.ListImagesOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list remote Docker images: %w", err)
	}

	// Find missing images.
	return findMissingImages(localImages, remoteImages), nil
}

// findMissingImages compares local and remote images and returns missing images.
func findMissingImages(local, remote []dockercli.DockerImage) []dockercli.DockerImage {
	missing := []dockercli.DockerImage{}
	remoteMap := map[string]struct{}{}

	for _, img := range remote {
		key := fmt.Sprintf("%s:%s", img.Repository, img.Tag)
		remoteMap[key] = struct{}{}
	}

	for _, img := range local {
		key := fmt.Sprintf("%s:%s", img.Repository, img.Tag)
		if _, exists := remoteMap[key]; !exists {
			missing = append(missing, img)
		}
	}

	return missing
}

// transferMissingImages transfers missing Docker images to the remote server.
func transferMissingImages(ctx context.Context, dockerClient *dockercli.DockerClient, missingImages []dockercli.DockerImage, sshConfig *shadowssh.Config) error {
	log.Info().Int("count", len(missingImages)).Msg("Starting transfer of missing images")
	for _, image := range missingImages {
		if err := dockerClient.TransferImage(ctx, image.Repository, sshConfig); err != nil {
			log.Error().Err(err).Str("image", image.Repository).Msg("Failed to transfer image")
			return fmt.Errorf("failed to transfer image %s: %w", image.Repository, err)
		}
		log.Info().Str("image", image.Repository).Msg("Image transferred successfully")
	}
	return nil
}

// deployCompose transfers the Docker Compose file and runs `docker compose up` on the remote server.
func deployCompose(ctx context.Context, composePath string, sshClient *shadowssh.Client, sshConfig *shadowssh.Config) error {
	remoteDir := "/composefiles"
	log.Info().Str("compose_path", composePath).Msg("Transferring Docker Compose file to remote host")
	if err := shadowscp.CopyFileToRemote(ctx, composePath, remoteDir, sshConfig); err != nil {
		log.Error().Err(err).Msg("Failed to transfer Docker Compose file")
		return fmt.Errorf("failed to transfer Docker Compose file: %w", err)
	}

	log.Info().Msg("Executing Docker Compose up on remote host")
	command := fmt.Sprintf("cd %s && docker compose up -d", remoteDir)
	output, err := sshClient.ExecuteCommand(ctx, command)
	if err != nil {
		log.Error().Err(err).Str("output", output).Msg("Failed to execute Docker Compose up")
		return fmt.Errorf("failed to execute Docker Compose up: %w", err)
	}
	log.Info().Msg("Docker Compose up executed successfully")
	return nil
}
