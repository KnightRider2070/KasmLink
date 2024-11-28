package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockercompose"
	shadowscp "kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/sshmanager"
	"kasmlink/pkg/userParser"
	"os"
	"path/filepath"
	"strings"
)

// DeployBackendServices deploys backend services based on the provided Docker Compose file and SSH configuration.
func DeployBackendServices(backendComposePath string, sshConfig shadowssh.SSHConfig) error {
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
		Str("host", sshConfig.NodeAddress).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
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

	missingImages, err := checkRemoteImages(client, imageNames)
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
			tarPath := filepath.Join(buildTarsDir, fmt.Sprintf("%s.tar", sanitizeImageName(image)))
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

				// Build the image locally
				log.Info().
					Str("image", image).
					Str("dockerfile", dockerfilePath).
					Msg("Building Docker image locally")

				if err := dockercli.BuildDockerImage(dockerfilePath, image); err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to build Docker image")
					return fmt.Errorf("failed to build image %s: %w", image, err)
				}

				// Export the image to a tar file
				log.Info().
					Str("image", image).
					Str("tar_path", tarPath).
					Msg("Exporting Docker image to tar")

				exportedTar, err := dockercli.ExportImageToTar(context.Background(), 3, image, tarPath)
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

			// Step 3.5 & 3.4: Copy the tar onto the remote node into /tmp
			remoteTmpDir := "/tmp"
			log.Info().
				Str("tar_path", tarPath).
				Str("remote_dir", remoteTmpDir).
				Msg("Copying tar file to remote node")

			if err := shadowscp.ShadowCopyFile(tarPath, remoteTmpDir); err != nil {
				log.Error().
					Err(err).
					Str("tar_path", tarPath).
					Str("remote_dir", remoteTmpDir).
					Msg("Failed to copy tar file to remote node")
				return fmt.Errorf("failed to copy tar %s to remote: %w", tarPath, err)
			}

			// Step 3.6: Load the docker image on the remote node
			loadCmd := fmt.Sprintf("docker load -i %s/%s.tar", remoteTmpDir, sanitizeImageName(image))
			log.Info().
				Str("image", image).
				Str("command", loadCmd).
				Msg("Loading Docker image on remote node")

			_, err = shadowssh.ExecuteCommand(client, loadCmd)
			if err != nil {
				log.Error().
					Err(err).
					Str("image", image).
					Str("command", loadCmd).
					Msg("Failed to load Docker image on remote node")
				return fmt.Errorf("failed to load image %s on remote: %w", image, err)
			}

			log.Info().
				Str("image", image).
				Msg("Successfully loaded Docker image on remote node")
		}
	} else {
		log.Info().
			Msg("All Docker images are already present on the remote node")
	}

	// Step 3.7: Copy compose file onto the node to /composefiles
	remoteComposeDir := "/composefiles"
	log.Info().
		Str("compose_path", backendComposePath).
		Str("remote_dir", remoteComposeDir).
		Msg("Copying Docker Compose file to remote node")

	if err := shadowscp.ShadowCopyFile(backendComposePath, remoteComposeDir); err != nil {
		log.Error().
			Err(err).
			Str("compose_path", backendComposePath).
			Str("remote_dir", remoteComposeDir).
			Msg("Failed to copy Docker Compose file to remote node")
		return fmt.Errorf("failed to copy compose file to remote: %w", err)
	}

	// Step 3.8: Execute 'docker compose up' on the remote node
	composeUpCmd := fmt.Sprintf("cd %s && docker compose up -d", remoteComposeDir)
	log.Info().
		Str("command", composeUpCmd).
		Msg("Executing 'docker compose up' on remote node")

	_, err = shadowssh.ExecuteCommand(client, composeUpCmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", composeUpCmd).
			Msg("Failed to execute 'docker compose up' on remote node")
		return fmt.Errorf("failed to execute docker compose up: %w", err)
	}

	log.Info().
		Msg("Deployment completed successfully")
	return nil
}

// checkRemoteImages checks which images are missing on the remote node.
func checkRemoteImages(client *ssh.Client, images []string) ([]string, error) {
	log.Debug().
		Msg("Executing remote Docker images command to list available images")

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	output, err := shadowssh.ExecuteCommand(client, cmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", cmd).
			Msg("Failed to execute remote Docker images command")
		return nil, err
	}

	remoteImages := strings.Split(output, "\n")
	missing := []string{}

	imageSet := make(map[string]struct{})
	for _, img := range remoteImages {
		trimmedImg := strings.TrimSpace(img)
		if trimmedImg != "" {
			imageSet[trimmedImg] = struct{}{}
		}
	}

	for _, img := range images {
		if _, exists := imageSet[img]; !exists {
			missing = append(missing, img)
			log.Debug().
				Str("image", img).
				Msg("Image is missing on remote node")
		}
	}

	return missing, nil
}

// findDockerfileForService searches for a Dockerfile in the ./dockerfiles/ directory that contains the serviceName.
func findDockerfileForService(serviceName string) (string, error) {
	log.Debug().
		Str("service_name", serviceName).
		Msg("Searching for Dockerfile matching service name")

	dockerfilesDir := "./dockerfiles"
	pattern := fmt.Sprintf("*%s*", serviceName)

	matchedFiles, err := filepath.Glob(filepath.Join(dockerfilesDir, pattern))
	if err != nil {
		log.Error().
			Err(err).
			Str("pattern", pattern).
			Msg("Failed to glob Dockerfiles")
		return "", fmt.Errorf("failed to glob dockerfiles: %w", err)
	}

	if len(matchedFiles) == 0 {
		log.Error().
			Str("service_name", serviceName).
			Str("directory", dockerfilesDir).
			Msg("No Dockerfile found containing service name")
		return "", fmt.Errorf("no Dockerfile found containing service name '%s' in %s", serviceName, dockerfilesDir)
	}

	if len(matchedFiles) > 1 {
		log.Error().
			Str("service_name", serviceName).
			Strs("matched_files", matchedFiles).
			Msg("Multiple Dockerfiles found for service name")
		return "", fmt.Errorf("multiple Dockerfiles found for service '%s' in %s: %v", serviceName, dockerfilesDir, matchedFiles)
	}

	log.Debug().
		Str("dockerfile", matchedFiles[0]).
		Msg("Found matching Dockerfile")

	return matchedFiles[0], nil
}

// sanitizeImageName sanitizes the image name to create valid filenames.
func sanitizeImageName(imageName string) string {
	// Replace '/' and ':' with underscores to prevent directory traversal or invalid filenames.
	return strings.ReplaceAll(strings.ReplaceAll(imageName, "/", "_"), ":", "_")
}

// DeployImages deploys a Docker image to the remote node based on the provided Dockerfile path.
func DeployImages(dockerFilePath string, imageName string, sshConfig shadowssh.SSHConfig) error {
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
		Str("host", sshConfig.NodeAddress).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
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

	// Step 2: Check if the Docker image tar file exists locally
	imageTarExistsLocally, localTarPath := checkLocalImageTarExists(imageName)
	if !imageTarExistsLocally {
		// Step 3: Build the Docker image locally
		log.Info().
			Str("image", imageName).
			Str("dockerfile_path", dockerFilePath).
			Msg("Building Docker image locally")

		if err := dockercli.BuildDockerImage(dockerFilePath, imageName); err != nil {
			log.Error().
				Err(err).
				Str("image", imageName).
				Msg("Failed to build Docker image")
			return fmt.Errorf("failed to build Docker image %s: %w", imageName, err)
		}
		log.Info().
			Str("image", imageName).
			Msg("Successfully built Docker image locally")

		// Step 4: Export the Docker image to a tar file
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

		exportedTar, err := dockercli.ExportImageToTar(context.Background(), 3, imageName, localTarPath)
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

	if err := shadowscp.ShadowCopyFile(localTarPath, "/tmp"); err != nil {
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

	loadCmd := fmt.Sprintf("docker load -i %s", "/tmp")
	_, err = shadowssh.ExecuteCommand(client, loadCmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("image", imageName).
			Str("command", loadCmd).
			Msg("Failed to load Docker image on remote node")
		return fmt.Errorf("failed to load Docker image %s on remote node: %w", imageName, err)
	}
	log.Info().
		Str("image", imageName).
		Msg("Successfully loaded Docker image on remote node")

	// Step 6: Remove the tar file from the remote node
	log.Info().
		Str("remote_tar_path", "/tmp").
		Msg("Removing tar file from remote node")

	removeCmd := fmt.Sprintf("rm %s", "/tmp")
	_, err = shadowssh.ExecuteCommand(client, removeCmd)
	if err != nil {
		log.Warn().
			Err(err).
			Str("command", removeCmd).
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

// checkLocalImageTarExists checks if the image tar file exists locally.
func checkLocalImageTarExists(imageName string) (bool, string) {
	sanitizedImageName := sanitizeImageName(imageName)
	localTarPath := filepath.Join("./tarfiles", fmt.Sprintf("%s.tar", sanitizedImageName))

	_, err := os.Stat(localTarPath)
	if err == nil {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar exists locally")
		return true, localTarPath
	}
	if os.IsNotExist(err) {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar does not exist locally")
		return false, localTarPath
	}
	// For other errors, log and treat as non-existent
	log.Error().
		Err(err).
		Str("local_tar_path", localTarPath).
		Msg("Error checking local tar file existence")
	return false, localTarPath
}

// DeployBackendServices deploys backend services based on the provided Docker Compose file and SSH configuration.
func DeployBackendServices(backendComposePath string, sshConfig shadowssh.SSHConfig) error {
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
		Str("host", sshConfig.NodeAddress).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
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

	missingImages, err := checkRemoteImages(client, imageNames)
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

				// Build the image locally
				log.Info().
					Str("image", image).
					Str("dockerfile", dockerfilePath).
					Msg("Building Docker image locally")

				if err := dockercli.BuildDockerImage(dockerfilePath, image); err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to build Docker image")
					return fmt.Errorf("failed to build image %s: %w", image, err)
				}

				// Export the image to a tar file
				log.Info().
					Str("image", image).
					Str("tar_path", tarPath).
					Msg("Exporting Docker image to tar")

				exportedTar, err := dockercli.ExportImageToTar(context.Background(), 3, image, tarPath)
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

			// Step 3.5 & 3.6: Copy the tar onto the remote node into /tmp and load the image
			remoteTmpDir := "/tmp"
			log.Info().
				Str("tar_path", tarPath).
				Str("remote_dir", remoteTmpDir).
				Msg("Copying tar file to remote node")

			if err := shadowscp.ShadowCopyFile(tarPath, remoteTmpDir, client); err != nil {
				log.Error().
					Err(err).
					Str("tar_path", tarPath).
					Str("remote_dir", remoteTmpDir).
					Msg("Failed to copy tar file to remote node")
				return fmt.Errorf("failed to copy tar %s to remote: %w", tarPath, err)
			}

			// Load the image on the remote node
			loadCmd := fmt.Sprintf("docker load -i %s/%s.tar", remoteTmpDir, sanitizeImageName(image))
			log.Info().
				Str("image", image).
				Str("command", loadCmd).
				Msg("Loading Docker image on remote node")

			_, err = shadowssh.ExecuteCommand(client, loadCmd)
			if err != nil {
				log.Error().
					Err(err).
					Str("image", image).
					Str("command", loadCmd).
					Msg("Failed to load Docker image on remote node")
				return fmt.Errorf("failed to load image %s on remote: %w", image, err)
			}

			log.Info().
				Str("image", image).
				Msg("Successfully loaded Docker image on remote node")

			// Step 3.7: Remove the tar file from the remote node
			removeCmd := fmt.Sprintf("rm %s/%s.tar", remoteTmpDir, sanitizeImageName(image))
			log.Info().
				Str("command", removeCmd).
				Msg("Removing tar file from remote node")

			_, err = shadowssh.ExecuteCommand(client, removeCmd)
			if err != nil {
				log.Warn().
					Err(err).
					Str("command", removeCmd).
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

	if err := shadowscp.ShadowCopyFile(backendComposePath, remoteComposeDir); err != nil {
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

	_, err = shadowssh.ExecuteCommand(client, composeUpCmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", composeUpCmd).
			Msg("Failed to execute 'docker compose up' on remote node")
		return fmt.Errorf("failed to execute docker compose up: %w", err)
	}

	log.Info().
		Msg("Deployment completed successfully")
	return nil
}

// CreateTestEnvironment creates a test environment based on the user configuration file.
func CreateTestEnvironment(userConfigurationFilePath string, sshConfig shadowssh.SSHConfig) error {
	// Step 1: Load user configuration from YAML file
	log.Info().
		Str("config_file", userConfigurationFilePath).
		Msg("Loading user configuration from YAML file")

	usersConfig, err := userParser.LoadConfig(userConfigurationFilePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("config_file", userConfigurationFilePath).
			Msg("Failed to load user configuration")
		return fmt.Errorf("failed to load user configuration: %w", err)
	}

	log.Info().
		Int("user_count", len(usersConfig.UserDetails)).
		Msg("Successfully loaded user configuration")

	// Step 2: Establish SSH connection with remote node using sshConfig
	log.Info().
		Str("host", sshConfig.NodeAddress).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
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

	// Iterate over each user in the configuration
	for _, user := range usersConfig.UserDetails {
		log.Info().
			Str("username", user.Username).
			Str("docker_image_tag", user.AssignedContainerTag).
			Msg("Processing user")

		// Step 3: Ensure that DockerImageTag exists on the remote node
		imageExists, err := checkRemoteImages(client, user.AssignedContainerTag)
		if err != nil {
			log.Error().
				Err(err).
				Str("image_tag", user.AssignedContainerTag).
				Msg("Error checking Docker image on remote node")
			return fmt.Errorf("error checking Docker image %s on remote node: %w", user.AssignedContainerTag, err)
		}

		if !imageExists {
			log.Error().
				Str("image_tag", user.AssignedContainerTag).
				Msg("Required Docker image tag does not exist on remote node")
			return fmt.Errorf("Docker image tag %s does not exist on remote node", user.AssignedContainerTag)
		}

		// Step 4: Create the user via KASM API
		log.Info().
			Str("username", user.Username).
			Msg("Creating user via KASM API")

		// Check if user already exists
		existingUser, err := api.GetUser("", user.Username)
		if err != nil {
			// If the error indicates that the user does not exist, proceed to create
			// Else, return the error
			if !strings.Contains(err.Error(), "not found") {
				log.Error().
					Err(err).
					Str("username", user.Username).
					Msg("Failed to retrieve user from KASM API")
				return fmt.Errorf("failed to retrieve user %s: %w", user.Username, err)
			}
			existingUser = nil
		}

		if existingUser != nil {
			log.Info().
				Str("username", user.Username).
				Str("user_id", existingUser.ID).
				Msg("User already exists in KASM")
			user.UserID = existingUser.ID
		} else {
			// Create user
			createdUser, err := api.CreateUser(api.TargetUser{
				Username: user.Username,
				ImageTag: user.AssignedContainerTag,
				// Add other necessary fields if required
			})
			if err != nil {
				log.Error().
					Err(err).
					Str("username", user.Username).
					Msg("Failed to create user via KASM API")
				return fmt.Errorf("failed to create user %s: %w", user.Username, err)
			}
			log.Info().
				Str("username", user.Username).
				Str("user_id", createdUser.ID).
				Msg("Successfully created user via KASM API")
			user.UserID = createdUser.ID
		}

		// Step 5: Add the user to the specified group
		log.Info().
			Str("username", user.Username).
			Str("role", user.Role).
			Msg("Adding user to the specified group via KASM API")

		groupID, err := api.GetGroupIDByName(user.Role)
		if err != nil {
			log.Error().
				Err(err).
				Str("role", user.Role).
				Msg("Failed to retrieve group ID from KASM API")
			return fmt.Errorf("failed to retrieve group ID for role %s: %w", user.Role, err)
		}

		err = api.AddUserToGroup(user.UserID, groupID)
		if err != nil {
			log.Error().
				Err(err).
				Str("user_id", user.UserID).
				Str("group_id", groupID).
				Msg("Failed to add user to group via KASM API")
			return fmt.Errorf("failed to add user %s to group %s: %w", user.Username, groupID, err)
		}
		log.Info().
			Str("user_id", user.UserID).
			Str("group_id", groupID).
			Msg("Successfully added user to group via KASM API")

		// Step 6: Update the YAML file with userId and KasmSessionOfContainer
		// Assume that KasmSessionOfContainer is retrieved or generated somehow
		// Replace the following line with actual logic to obtain the session ID
		kasmSessionOfContainer := "session-id-placeholder" // Replace with actual logic

		log.Info().
			Str("username", user.Username).
			Str("user_id", user.UserID).
			Str("kasm_session_of_container", kasmSessionOfContainer).
			Msg("Updating user configuration in YAML file")

		err = userParser.UpdateUserConfig(userConfigurationFilePath, user.Username, user.UserID, kasmSessionOfContainer)
		if err != nil {
			log.Error().
				Err(err).
				Str("username", user.Username).
				Msg("Failed to update user configuration in YAML file")
			return fmt.Errorf("failed to update user %s configuration: %w", user.Username, err)
		}
		log.Info().
			Str("username", user.Username).
			Msg("Successfully updated user configuration in YAML file")
	}

	log.Info().
		Msg("Test environment creation completed successfully")

	return nil
}
