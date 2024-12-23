package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kasmlink/pkg/dockercli"
	shadowscp "kasmlink/pkg/shadowscp"
	shadowssh "kasmlink/pkg/shadowssh"

	"github.com/rs/zerolog/log"
)

// DeployImages deploys a Docker image to the remote node based on the provided Dockerfile path.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - dockerFilePath: Path to the local Dockerfile.
// - imageName: Name/tag of the Docker image to build and deploy.
// - sshConfig: SSH configuration for connecting to the remote node.
// Returns:
// - An error if any step in the deployment process fails.
func DeployImages(ctx context.Context, dockerFilePath string, imageName string, sshConfig *shadowssh.SSHConfig) error {
	// Step 1: Check if the Dockerfile exists locally
	log.Info().
		Str("dockerfile_path", dockerFilePath).
		Msg("Checking existence of Dockerfile")

	if _, err := os.Stat(dockerFilePath); os.IsNotExist(err) {
		log.Error().
			Err(err).
			Str("dockerfile_path", dockerFilePath).
			Msg("Dockerfile does not exist")
		return fmt.Errorf("Dockerfile does not exist at path: %s", dockerFilePath)
	}

	// Step 2: Establish SSH connection with remote node using sshConfig
	log.Info().
		Str("host", sshConfig.Host).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(ctx, sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.Host).
			Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().
				Err(cerr).
				Msg("Failed to close SSH connection gracefully")
		} else {
			log.Debug().
				Msg("SSH connection closed")
		}
	}()

	// Step 3: Check if the Docker image tar file exists locally
	imageTarExistsLocally, localTarPath := checkLocalImageTarExists(imageName)
	if !imageTarExistsLocally {
		// Step 3.1: Build the Docker image locally
		log.Info().
			Str("image", imageName).
			Str("dockerfile_path", dockerFilePath).
			Msg("Building Docker image locally")

		if err := dockercli.BuildDockerImage(ctx, 3, dockerFilePath, imageName); err != nil {
			log.Error().
				Err(err).
				Str("image", imageName).
				Msg("Failed to build Docker image")
			return fmt.Errorf("failed to build Docker image %s: %w", imageName, err)
		}
		log.Info().
			Str("image", imageName).
			Msg("Successfully built Docker image locally")

		// Step 3.2: Export the Docker image to a tar file
		log.Info().
			Str("image", imageName).
			Msg("Exporting Docker image to tar")

		buildTarsDir := "./tarfiles"
		if _, err := os.Stat(buildTarsDir); os.IsNotExist(err) {
			log.Info().
				Str("directory", buildTarsDir).
				Msg("Creating tarfiles directory")

			if err := os.MkdirAll(buildTarsDir, 0755); err != nil {
				log.Error().
					Err(err).
					Str("directory", buildTarsDir).
					Msg("Failed to create tarfiles directory")
				return fmt.Errorf("failed to create tarfiles directory: %w", err)
			}
		}

		// Define the tar file path
		localTarPath = filepath.Join(buildTarsDir, fmt.Sprintf("%s.tar", sanitizeImageName(imageName)))

		exportedTar, err := dockercli.ExportImageToTar(ctx, 3, imageName, localTarPath)
		if err != nil {
			log.Error().
				Err(err).
				Str("image", imageName).
				Str("tar_path", localTarPath).
				Msg("Failed to export Docker image to tar")
			return fmt.Errorf("failed to export Docker image %s to tar: %w", imageName, err)
		}
		log.Info().
			Str("image", imageName).
			Str("tar_path", exportedTar).
			Msg("Successfully exported Docker image to tar")
	} else {
		log.Info().
			Str("image", imageName).
			Str("tar_path", localTarPath).
			Msg("Image tar already exists locally. Skipping build and export.")
	}

	// Step 4: Copy the tar file onto the remote node into /tmp
	log.Info().
		Str("tar_path", localTarPath).
		Str("remote_dir", "/tmp").
		Msg("Copying tar file to remote node")

	if err := shadowscp.ShadowCopyFile(ctx, localTarPath, "/tmp", sshConfig); err != nil {
		log.Error().
			Err(err).
			Str("tar_path", localTarPath).
			Str("remote_dir", "/tmp").
			Msg("Failed to copy tar file to remote node")
		return fmt.Errorf("failed to copy tar %s to remote: %w", localTarPath, err)
	}
	log.Info().
		Str("tar_path", localTarPath).
		Str("remote_dir", "/tmp").
		Msg("Successfully copied tar file to remote node")

	// Step 5: Load the Docker image on the remote node
	log.Info().
		Str("image", imageName).
		Str("remote_tar_path", "/tmp").
		Msg("Loading Docker image on remote node")

	// Define the remote tar file path
	remoteTarPath := filepath.Join("/tmp", fmt.Sprintf("%s.tar", sanitizeImageName(imageName)))

	// Execute the docker load command on the remote node
	loadCmd := fmt.Sprintf("docker load -i %s", remoteTarPath)
	output, err := client.ExecuteCommandWithOutput(ctx, loadCmd, 1*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("image", imageName).
			Str("command", loadCmd).
			Str("output", output).
			Msg("Failed to load Docker image on remote node")
		return fmt.Errorf("failed to load Docker image %s on remote node: %w", imageName, err)
	}
	log.Info().
		Str("image", imageName).
		Msg("Successfully loaded Docker image on remote node")

	// Step 6: Remove the tar file from the remote node
	log.Info().
		Str("remote_tar_path", remoteTarPath).
		Msg("Removing tar file from remote node")

	removeCmd := fmt.Sprintf("rm %s", remoteTarPath)
	output, err = client.ExecuteCommandWithOutput(ctx, removeCmd, 30*time.Second)
	if err != nil {
		log.Warn().
			Err(err).
			Str("command", removeCmd).
			Str("output", output).
			Msg("Failed to remove tar file from remote node")
		// Not returning error as removal failure is non-critical
	} else {
		log.Info().
			Str("command", removeCmd).
			Msg("Successfully removed tar file from remote node")
	}

	log.Info().
		Str("image", imageName).
		Msg("Image deployment process completed successfully")

	return nil
}
