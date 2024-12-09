// procedures/procedures.go
package procedures

import (
	"context"
	"fmt"
	embedfiles "kasmlink/embedded"
	"os"
	"path/filepath"
	"time"

	"kasmlink/pkg/dockercli"
	shadowscp "kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/sshmanager"

	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

// Constants for default configurations.
const (
	DefaultBaseImage       = "opensuse/leap:15.5"
	DefaultBuildContextDir = "workspace-core-image"
)

// BuildCoreImageKasm orchestrates the Docker image build using the embedded Dockerfile and base image.
// It utilizes the dockercli package to create the build context and build the Docker image.
// Parameters:
// - imageTag: The tag to assign to the built Docker image (e.g., "kasm/core:latest").
// - baseImage: The base image to use for building. If empty, DefaultBaseImage is used.
// Returns:
// - An error if the build process fails.
func BuildCoreImageKasm(imageTag, baseImage string) error {
	if imageTag == "" {
		return fmt.Errorf("imageTag cannot be empty")
	}

	if baseImage == "" {
		baseImage = DefaultBaseImage
	}

	log.Info().
		Str("imageTag", imageTag).
		Str("baseImage", baseImage).
		Msg("Starting Docker image build")

	// Create the Docker client with environment variables and API version negotiation.
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create Docker client")
		return fmt.Errorf("could not create Docker client: %w", err)
	}
	defer func() {
		if cerr := cli.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("Failed to close Docker client")
		}
	}()

	// Create tar archive from embedded Dockerfile and build context.
	_, err = dockercli.CreateTarFromEmbedded(embedfiles.EmbeddedKasmDirectory, DefaultBuildContextDir)
	if err != nil {
		log.Error().
			Err(err).
			Str("embeddedDir", "Embeded KASM files").
			Msg("Failed to create build context tar")
		return fmt.Errorf("failed to create build context tar: %w", err)
	}

	// Define the number of retries, e.g., 3
	retries := 3

	// Corrected function call
	err = dockercli.BuildDockerImage(context.Background(), retries, "dockerfile-kasm-core-suse", imageTag)
	if err != nil {
		log.Error().
			Err(err).
			Str("imageTag", imageTag).
			Msg("Docker image build failed")
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	log.Info().
		Str("imageTag", imageTag).
		Msg("Docker image built successfully")
	return nil
}

// DeployKasmDockerImage builds, exports, and loads a Docker image on a remote node.
// If a localTarFilePath is provided, it will use that file instead of building a new image.
// Parameters:
// - imageTag: The Docker image tag to deploy.
// - baseImage: The base image to use for building (if building).
// - targetNodePath: The destination path on the remote node where the image will be loaded.
// - localTarFilePath: Optional local tar file path. If provided and exists, it will be used instead of building.
// Returns:
// - An error if any step in the deployment process fails.
func DeployKasmDockerImage(imageTag, baseImage, targetNodePath, localTarFilePath string) error {
	var tarFilePath string
	var err error

	// Step 1: Determine the tar file to use.
	if localTarFilePath != "" {
		if _, err = os.Stat(localTarFilePath); err == nil {
			// Local tar file exists, use it.
			tarFilePath = localTarFilePath
			log.Info().Msg("Using existing local tar file for Docker image deployment")
		} else {
			log.Error().
				Err(err).
				Str("localTarFilePath", localTarFilePath).
				Msg("Specified local tar file does not exist")
			return fmt.Errorf("local tar file specified but not found: %w", err)
		}
	} else {
		// Step 2: Build the Docker image if no local tar file is provided.
		if err = BuildCoreImageKasm(imageTag, baseImage); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to build Docker image")
			return fmt.Errorf("failed to build Docker image: %w", err)
		}

		// Step 3: Export image to tar file.
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		// Define the number of retries, e.g., 3
		retries := 3

		// Define the output file path using a temporary file
		tempFile, err := os.CreateTemp("", "docker-image-*.tar")
		if err != nil {
			log.Error().
				Err(err).
				Msg("Could not create temporary tar file")
			return fmt.Errorf("could not create temporary tar file: %w", err)
		}
		defer func() {
			if cerr := tempFile.Close(); cerr != nil {
				log.Error().
					Err(cerr).
					Str("tarFilePath", tempFile.Name()).
					Msg("Failed to close tar file")
			}
		}()

		outputFile := tempFile.Name()

		// Corrected function call
		tarFilePath, err = dockercli.ExportImageToTar(ctx, retries, imageTag, outputFile)
		if err != nil {
			log.Error().
				Err(err).
				Str("imageTag", imageTag).
				Msg("Failed to export Docker image to tar")
			return fmt.Errorf("failed to export Docker image to tar: %w", err)
		}
		defer func() {
			if cerr := os.Remove(tarFilePath); cerr != nil {
				log.Error().
					Err(cerr).
					Str("tarFilePath", tarFilePath).
					Msg("Failed to remove temporary tar file")
			}
		}()
	}

	// Step 4: Establish SSH connection to target node.
	sshConfig, err := configureSSH()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to configure SSH settings")
		return fmt.Errorf("failed to configure SSH settings: %w", err)
	}

	sshClient, err := shadowssh.NewSSHClient(context.Background(), sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to establish SSH connection to remote node")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := sshClient.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("Failed to close SSH client")
		}
	}()

	// Step 5: Copy the tar file to the remote node.
	log.Info().
		Str("localTarFilePath", tarFilePath).
		Str("remoteDir", targetNodePath).
		Msg("Starting file copy to remote node via SCP")

	err = shadowscp.ShadowCopyFile(context.Background(), tarFilePath, targetNodePath, sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("tarFilePath", tarFilePath).
			Str("remoteDir", targetNodePath).
			Msg("Failed to copy tar file to remote node")
		return fmt.Errorf("failed to copy tar file to remote node: %w", err)
	}

	log.Info().Msg("Tar file copied to remote node successfully")

	// Step 6: Import the Docker image on the remote node.
	importCommand := fmt.Sprintf("docker load -i %s/%s", targetNodePath, filepath.Base(tarFilePath))
	log.Info().
		Str("command", importCommand).
		Msg("Importing Docker image on remote node")

	output, err := sshClient.ExecuteCommandWithOutput(context.Background(), importCommand, 1*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", importCommand).
			Str("output", output).
			Msg("Failed to import Docker image on remote node")
		return fmt.Errorf("failed to import Docker image on remote node: %w", err)
	}

	log.Info().Msg("Docker image imported successfully on remote node")
	return nil
}

// DeployComposeFile uploads a specified Docker Compose file and deploys the services on the target node.
// Parameters:
// - composeFilePath: The local path to the Docker Compose YAML file.
// - targetNodePath: The destination directory on the remote node where the Compose file will be placed.
// Returns:
// - An error if any step in the deployment process fails.
func DeployComposeFile(composeFilePath, targetNodePath string) error {
	// Validate compose file existence.
	if _, err := os.Stat(composeFilePath); os.IsNotExist(err) {
		log.Error().
			Err(err).
			Str("composeFilePath", composeFilePath).
			Msg("Compose file does not exist")
		return fmt.Errorf("compose file does not exist at path %s: %w", composeFilePath, err)
	}

	// Step 1: Establish SSH connection to target node.
	sshConfig, err := configureSSH()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to configure SSH settings")
		return fmt.Errorf("failed to configure SSH settings: %w", err)
	}

	sshClient, err := shadowssh.NewSSHClient(context.Background(), sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to establish SSH connection to remote node")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := sshClient.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("Failed to close SSH client")
		}
	}()

	// Step 2: Copy compose file onto node.
	log.Info().
		Str("source", composeFilePath).
		Str("destination", targetNodePath).
		Msg("Starting to copy compose file onto remote node")

	err = shadowscp.ShadowCopyFile(context.Background(), composeFilePath, targetNodePath, sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("nodeAddress", sshConfig.Host).
			Str("targetPath", targetNodePath).
			Msg("Failed to copy compose file onto remote node")
		return fmt.Errorf("failed to copy compose file onto remote node: %w", err)
	}

	log.Info().
		Str("nodeAddress", sshConfig.Host).
		Str("composeFile", filepath.Join(targetNodePath, filepath.Base(composeFilePath))).
		Msg("Compose file copied successfully")

	// Step 3: Start Docker Compose on the remote node.
	targetNodeComposeFilePath := filepath.Join(targetNodePath, filepath.Base(composeFilePath))
	dockerComposeUpCommand := fmt.Sprintf("docker compose -f %s up -d", targetNodeComposeFilePath)

	log.Info().
		Str("command", dockerComposeUpCommand).
		Str("nodeAddress", sshConfig.Host).
		Msg("Starting Docker Compose on the remote node")

	output, err := sshClient.ExecuteCommandWithOutput(context.Background(), dockerComposeUpCommand, 1*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.Host).
			Str("command", dockerComposeUpCommand).
			Str("output", output).
			Msg("Failed to start Docker Compose on remote node")
		return fmt.Errorf("failed to start Docker Compose on remote node: %w", err)
	}

	log.Info().
		Str("nodeAddress", sshConfig.Host).
		Msg("Docker Compose deployed successfully on target node")
	return nil
}

// configureSSH sets up the SSH configuration based on environment variables or other sources.
// It returns an SSHConfig instance or an error if configuration fails.
func configureSSH() (*shadowssh.SSHConfig, error) {
	// Example: Fetch SSH configurations from environment variables.
	// Replace these with your actual configuration retrieval logic.
	username := os.Getenv("SSH_USERNAME")
	password := os.Getenv("SSH_PASSWORD")
	host := os.Getenv("SSH_HOST")
	portStr := os.Getenv("SSH_PORT")
	knownHostsFile := os.Getenv("SSH_KNOWN_HOSTS")
	connectionTimeoutStr := os.Getenv("SSH_CONNECTION_TIMEOUT")

	// Parse port.
	port := 22 // Default SSH port.
	if portStr != "" {
		_, err := fmt.Sscanf(portStr, "%d", &port)
		if err != nil {
			log.Warn().
				Err(err).
				Str("SSH_PORT", portStr).
				Msg("Invalid SSH port format; using default port 22")
			port = 22
		}
	}

	// Parse connection timeout.
	connectionTimeout := 10 * time.Second // Default timeout.
	if connectionTimeoutStr != "" {
		duration, err := time.ParseDuration(connectionTimeoutStr)
		if err != nil {
			log.Warn().
				Err(err).
				Str("SSH_CONNECTION_TIMEOUT", connectionTimeoutStr).
				Msg("Invalid connection timeout format; using default timeout of 10s")
			connectionTimeout = 10 * time.Second
		} else {
			connectionTimeout = duration
		}
	}

	// Validate required fields.
	if username == "" || host == "" {
		return nil, fmt.Errorf("SSH_USERNAME and SSH_HOST must be set")
	}

	// Initialize SSHConfig.
	sshConfig, err := shadowssh.NewSSHConfig(username, password, host, port, knownHostsFile, connectionTimeout)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize SSH configuration")
		return nil, fmt.Errorf("failed to initialize SSH configuration: %w", err)
	}

	return sshConfig, nil
}
