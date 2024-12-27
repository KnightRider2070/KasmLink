package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/shadowscp"
	"kasmlink/pkg/shadowssh"

	"github.com/rs/zerolog/log"
)

// DeployImage deploys a Docker image to the remote node.
func DeployImage(ctx context.Context, imageName, tarDirectory string, dockerClient *dockercli.DockerClient, sshConfig *shadowssh.Config) error {
	log.Info().Str("image_name", imageName).Msg("Starting standalone deployment of image.")

	// Step 1: Establish SSH connection.
	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection.")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	// Step 2: Check if the image exists on the remote node.
	checkCommand := fmt.Sprintf("docker images -q %s", imageName)
	output, err := sshClient.ExecuteCommand(ctx, checkCommand)
	if err == nil && output != "" {
		log.Info().Str("image_name", imageName).Msg("Image already exists on remote node.")
		return nil
	}

	// Step 3: Attempt to pull the image on the remote node.
	pullCommand := fmt.Sprintf("docker pull %s", imageName)
	if _, err := sshClient.ExecuteCommand(ctx, pullCommand); err == nil {
		log.Info().Str("image_name", imageName).Msg("Image successfully pulled on remote node.")
		return nil
	}
	log.Warn().Str("image_name", imageName).Msg("Failed to pull image remotely. Checking local availability.")

	// Step 4: Check local availability of the image.
	imageTarPath := filepath.Join(tarDirectory, fmt.Sprintf("%s.tar", imageName))
	if _, err := os.Stat(imageTarPath); err == nil {
		log.Info().Str("tar_path", imageTarPath).Msg("Image tarball found locally.")
	} else {
		// Export image locally if tarball does not exist.
		log.Warn().Str("image_name", imageName).Msg("Image tarball not found. Exporting image locally.")
		if err := dockerClient.SaveImage(ctx, imageName, imageTarPath); err != nil {
			log.Error().Err(err).Msg("Failed to save image locally.")
			return fmt.Errorf("failed to save image locally: %w", err)
		}
		defer os.Remove(imageTarPath)
	}

	// Step 5: Transfer the tarball to the remote node.
	log.Info().Str("tar_path", imageTarPath).Msg("Transferring tarball to remote node.")
	remoteTarPath := fmt.Sprintf("/tmp/%s.tar", imageName)
	if err := shadowscp.CopyFileToRemote(ctx, imageTarPath, remoteTarPath, sshConfig); err != nil {
		log.Error().Err(err).Msg("Failed to transfer tarball to remote node.")
		return fmt.Errorf("failed to transfer tarball to remote node: %w", err)
	}

	// Step 6: Load the image on the remote node.
	loadCommand := fmt.Sprintf("docker load -i %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, loadCommand); err != nil {
		log.Error().Err(err).Msg("Failed to load image on remote node.")
		return fmt.Errorf("failed to load image on remote node: %w", err)
	}
	log.Info().Str("image_name", imageName).Msg("Image successfully loaded on remote node.")

	// Step 7: Clean up remote tarball.
	removeCommand := fmt.Sprintf("rm -f %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, removeCommand); err != nil {
		log.Warn().Err(err).Str("tar_path", remoteTarPath).Msg("Failed to remove tarball from remote node.")
	}

	log.Info().Str("image_name", imageName).Msg("Standalone deployment of image completed successfully.")
	return nil
}
