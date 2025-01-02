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

	// Step 1: Check if the image exists locally using DockerClient.
	options := dockercli.ListImagesOptions{Repository: imageName}
	images, err := dockerClient.ListImages(ctx, options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list Docker images locally.")
		return fmt.Errorf("failed to list Docker images locally: %w", err)
	}

	if len(images) > 0 {
		log.Info().Str("image_name", imageName).Msg("Image already exists locally.")
	} else {
		// Step 2: Attempt to pull the image locally.
		log.Warn().Str("image_name", imageName).Msg("Image not found locally. Attempting to pull.")
		if err := dockerClient.PullImage(ctx, imageName); err != nil {
			log.Error().Err(err).Msg("Failed to pull image locally.")
			return fmt.Errorf("failed to pull image locally: %w", err)
		}
		log.Info().Str("image_name", imageName).Msg("Image successfully pulled locally.")
	}

	// Step 3: Export the image to a tarball if necessary.
	imageTarPath := filepath.Join(tarDirectory, fmt.Sprintf("%s.tar", imageName))
	if _, err := os.Stat(imageTarPath); os.IsNotExist(err) {
		log.Info().Str("image_name", imageName).Msg("Exporting image to tarball.")
		if err := dockerClient.SaveImage(ctx, imageName, imageTarPath); err != nil {
			log.Error().Err(err).Msg("Failed to save image to tarball.")
			return fmt.Errorf("failed to save image to tarball: %w", err)
		}
		defer os.Remove(imageTarPath)
	} else {
		log.Info().Str("tar_path", imageTarPath).Msg("Image tarball already exists.")
	}

	// Step 4: Transfer the tarball to the remote node.
	log.Info().Str("tar_path", imageTarPath).Msg("Transferring tarball to remote node.")
	remoteTarPath := fmt.Sprintf("/tmp/%s.tar", imageName)
	if err := shadowscp.CopyFileToRemote(ctx, imageTarPath, remoteTarPath, sshConfig); err != nil {
		log.Error().Err(err).Msg("Failed to transfer tarball to remote node.")
		return fmt.Errorf("failed to transfer tarball to remote node: %w", err)
	}

	// Step 5: Load the image on the remote node.
	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection.")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	loadCommand := fmt.Sprintf("docker load -i %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, loadCommand); err != nil {
		log.Error().Err(err).Msg("Failed to load image on remote node.")
		return fmt.Errorf("failed to load image on remote node: %w", err)
	}

	log.Info().Str("image_name", imageName).Msg("Image successfully loaded on remote node.")

	// Step 6: Clean up remote tarball.
	removeCommand := fmt.Sprintf("rm -f %s", remoteTarPath)
	if _, err := sshClient.ExecuteCommand(ctx, removeCommand); err != nil {
		log.Warn().Err(err).Str("tar_path", remoteTarPath).Msg("Failed to remove tarball from remote node.")
	}

	log.Info().Str("image_name", imageName).Msg("Standalone deployment of image completed successfully.")
	return nil
}
