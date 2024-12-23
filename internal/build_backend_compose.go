package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockercompose"
	shadowscp "kasmlink/pkg/shadowscp"
	shadowssh "kasmlink/pkg/shadowssh"

	"github.com/rs/zerolog/log"
)

// DeployBackendServices deploys backend services based on the provided Docker Compose file and SSH configuration.
func DeployBackendServices(ctx context.Context, backendComposePath string, sshConfig *shadowssh.SSHConfig) error {
	// Step 1: Check if the Docker Compose file exists locally
	log.Info().
		Str("path", backendComposePath).
		Msg("Checking existence of Docker Compose file")

	if _, err := os.Stat(backendComposePath); os.IsNotExist(err) {
		log.Error().
			Err(err).
			Str("path", backendComposePath).
			Msg("Compose file does not exist")
		return fmt.Errorf("compose file does not exist at path: %s", backendComposePath)
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

	// Step 3: Load compose into structs
	log.Info().
		Str("path", backendComposePath).
		Msg("Loading Docker Compose file")

	compose, err := dockercompose.LoadComposeFile(backendComposePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", backendComposePath).
			Msg("Failed to load Docker Compose file")
		return fmt.Errorf("failed to load compose file: %w", err)
	}

	// Step 3.2: Extract image names used
	imageNames := make([]string, 0)
	serviceNames := make([]string, 0)
	for serviceName, service := range compose.Services {
		serviceNames = append(serviceNames, serviceName)
		imageNames = append(imageNames, service.Image)
	}
	log.Debug().
		Int("service_count", len(serviceNames)).
		Int("image_count", len(imageNames)).
		Msg("Extracted service and image names from Compose file")

	// Step 3.3: Check which images are present on the remote node
	log.Info().
		Msg("Checking for missing Docker images on the remote node")

	missingImages, err := checkRemoteImages(ctx, client, imageNames)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check remote Docker images")
		return fmt.Errorf("failed to check remote images: %w", err)
	}

	if len(missingImages) > 0 {
		log.Info().
			Int("missing_images_count", len(missingImages)).
			Msg("Identified missing Docker images on remote node")

		// Ensure buildTars directory exists locally
		buildTarsDir := "./buildTars"
		if _, err := os.Stat(buildTarsDir); os.IsNotExist(err) {
			log.Info().
				Str("directory", buildTarsDir).
				Msg("Creating buildTars directory")

			err := os.Mkdir(buildTarsDir, 0755)
			if err != nil {
				log.Error().
					Err(err).
					Str("directory", buildTarsDir).
					Msg("Failed to create buildTars directory")
				return fmt.Errorf("failed to create buildTars directory: %w", err)
			}
		}

		for _, image := range missingImages {
			sanitizedImageName := sanitizeImageName(image)
			tarPath := filepath.Join(buildTarsDir, fmt.Sprintf("%s.tar", sanitizedImageName))
			log.Debug().
				Str("image", image).
				Str("tar_path", tarPath).
				Msg("Processing missing image")

			if _, err := os.Stat(tarPath); os.IsNotExist(err) {
				// Step 3.4: Check for Dockerfile and build image if necessary
				log.Info().
					Str("image", image).
					Msg("Docker image tar not found locally, searching for Dockerfile")

				dockerfilePath, err := findDockerfileForService(image)
				if err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to find Dockerfile for image")
					return fmt.Errorf("failed to find Dockerfile for image %s: %w", image, err)
				}

				// Step 3.5: Build the image locally
				log.Info().
					Str("image", image).
					Str("dockerfile", dockerfilePath).
					Msg("Building Docker image locally")

				// Define build context directory if required
				buildContextDir := "./buildContexts" // Adjust as needed
				if err := os.MkdirAll(buildContextDir, 0755); err != nil {
					log.Error().
						Err(err).
						Str("directory", buildContextDir).
						Msg("Failed to create build context directory")
					return fmt.Errorf("failed to create build context directory: %w", err)
				}

				if err := dockercli.BuildDockerImage(ctx, 3, dockerfilePath, image); err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to build Docker image")
					return fmt.Errorf("failed to build image %s: %w", image, err)
				}

				// Step 3.6: Export the image to a tar file
				log.Info().
					Str("image", image).
					Str("tar_path", tarPath).
					Msg("Exporting Docker image to tar")

				exportedTar, err := dockercli.ExportImageToTar(ctx, 3, image, tarPath)
				if err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Str("tar_path", tarPath).
						Msg("Failed to export Docker image to tar")
					return fmt.Errorf("failed to export image %s to tar: %w", image, err)
				}
				log.Info().
					Str("image", image).
					Str("tar_path", exportedTar).
					Msg("Successfully exported Docker image to tar")
			}

			// Step 3.7: Copy the tar onto the remote node into /tmp
			remoteTmpDir := "/tmp"
			log.Info().
				Str("tar_path", tarPath).
				Str("remote_dir", remoteTmpDir).
				Msg("Copying tar file to remote node")

			if err := shadowscp.ShadowCopyFile(ctx, tarPath, remoteTmpDir, sshConfig); err != nil {
				log.Error().
					Err(err).
					Str("tar_path", tarPath).
					Str("remote_dir", remoteTmpDir).
					Msg("Failed to copy tar file to remote node")
				return fmt.Errorf("failed to copy tar %s to remote: %w", tarPath, err)
			}

			// Step 3.8: Load the image on the remote node
			loadCmd := fmt.Sprintf("docker load -i %s/%s.tar", remoteTmpDir, sanitizedImageName)
			log.Info().
				Str("image", image).
				Str("command", loadCmd).
				Msg("Loading Docker image on remote node")

			output, err := client.ExecuteCommandWithOutput(ctx, loadCmd, 1*time.Minute)
			if err != nil {
				log.Error().
					Err(err).
					Str("image", image).
					Str("command", loadCmd).
					Str("output", output).
					Msg("Failed to load Docker image on remote node")
				return fmt.Errorf("failed to load image %s on remote: %w", image, err)
			}

			log.Info().
				Str("image", image).
				Msg("Successfully loaded Docker image on remote node")

			// Step 3.9: Remove the tar file from the remote node
			removeCmd := fmt.Sprintf("rm %s/%s.tar", remoteTmpDir, sanitizedImageName)
			log.Info().
				Str("command", removeCmd).
				Msg("Removing tar file from remote node")

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
		}
	} else {
		log.Info().
			Msg("All Docker images are already present on the remote node")
	}

	// Step 4: Copy compose file onto the node to /composefiles
	remoteComposeDir := "/composefiles"
	log.Info().
		Str("compose_path", backendComposePath).
		Str("remote_dir", remoteComposeDir).
		Msg("Copying Docker Compose file to remote node")

	if err := shadowscp.ShadowCopyFile(ctx, backendComposePath, remoteComposeDir, sshConfig); err != nil {
		log.Error().
			Err(err).
			Str("compose_path", backendComposePath).
			Str("remote_dir", remoteComposeDir).
			Msg("Failed to copy Docker Compose file to remote node")
		return fmt.Errorf("failed to copy compose file to remote: %w", err)
	}

	// Step 5: Execute 'docker compose up' on the remote node
	composeUpCmd := fmt.Sprintf("cd %s && docker compose up -d", remoteComposeDir)
	log.Info().
		Str("command", composeUpCmd).
		Msg("Executing 'docker compose up' on remote node")

	output, err := client.ExecuteCommandWithOutput(ctx, composeUpCmd, 2*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", composeUpCmd).
			Str("output", output).
			Msg("Failed to execute 'docker compose up' on remote node")
		return fmt.Errorf("failed to execute docker compose up: %w", err)
	}

	log.Info().
		Msg("Deployment completed successfully")
	return nil
}
