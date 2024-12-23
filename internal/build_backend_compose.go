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
	if err := validateFileExists(backendComposePath); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

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

	missingImages, err := dockerClient.ListImages(ctx, dockercli.ListImagesOptions{SSH: &dockercli.SSHOptions{
		Host: sshConfig.Host, Port: sshConfig.Port, User: sshConfig.Username, Password: sshConfig.Password}})
	if err != nil {
		log.Error().Err(err).Msg("Failed to list Docker images")
		return fmt.Errorf("failed to list Docker images: %w", err)
	}

	if len(missingImages) > 0 {
		if err := transferMissingImages(ctx, missingImages, dockerClient, sshConfig); err != nil {
			return fmt.Errorf("error during image transfer: %w", err)
		}
	}

	if err := deployCompose(ctx, backendComposePath, sshClient); err != nil {
		return fmt.Errorf("deployment error: %w", err)
	}

	log.Info().Msg("Deployment completed successfully")
	return nil
}

func validateFileExists(path string) error {
	log.Info().Str("path", path).Msg("Validating file existence")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error().Err(err).Msgf("File does not exist: %s", path)
		return fmt.Errorf("file does not exist: %s", path)
	}
	return nil
}

func transferMissingImages(ctx context.Context, missingImages []dockercli.DockerImage, dockerClient *dockercli.DockerClient, sshConfig *shadowssh.Config) error {
	log.Info().Int("count", len(missingImages)).Msg("Starting transfer of missing images")
	for _, image := range missingImages {
		if err := dockerClient.TransferImage(ctx, image.Repository, &dockercli.SSHOptions{
			Host: sshConfig.Host, Port: sshConfig.Port, User: sshConfig.Username, Password: sshConfig.Password}); err != nil {
			log.Error().Err(err).Str("image", image.Repository).Msg("Failed to transfer image")
			return fmt.Errorf("failed to transfer image %s: %w", image.Repository, err)
		}
		log.Info().Str("image", image.Repository).Msg("Image transferred successfully")
	}
	return nil
}

func deployCompose(ctx context.Context, composePath string, sshClient *shadowssh.Client) error {
	remoteDir := "/composefiles"
	log.Info().Str("compose_path", composePath).Msg("Transferring Docker Compose file to remote host")
	if err := shadowscp.CopyFileToRemote(ctx, composePath, remoteDir, &shadowssh.Config{
		Host: sshClient.Config().Host, Port: sshClient.Config().Port, Username: sshClient.Config().Username}); err != nil {
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
