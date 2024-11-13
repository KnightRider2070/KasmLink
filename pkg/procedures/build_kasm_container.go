package procedures

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockerutils"
	shadowscp "kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/ssh"
	"os"
	"path/filepath"
	"time"
)

// BuildCoreImageKasm orchestrates the Docker image build using the embedded Dockerfile and base image.
func BuildCoreImageKasm(imageTag, baseImage string) error {
	if baseImage == "" {
		baseImage = "opensuse/leap:15.5"
	}
	log.Info().Str("imageTag", imageTag).Str("baseImage", baseImage).Msg("Building Docker image")

	// Create the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Create tar archive from embedded Dockerfile and build context
	buildContext, err := dockerutils.CreateTarFromEmbedded(embedfiles.EmbeddedKasmDirectory, "workspace-core-image")
	if err != nil {
		return fmt.Errorf("failed to create build context tar: %v", err)
	}

	// Define build arguments for Docker
	buildArgs := map[string]*string{"BASE_IMAGE": &baseImage}

	// Build Docker image
	return dockerutils.BuildDockerImage(cli, imageTag, "dockerfile-kasm-core-suse", buildContext, buildArgs)
}

// DeployKasmDockerImage builds, exports, and loads a Docker image on a remote node.
// If a localTarFilePath is provided, it will use that file instead of building a new image.
func DeployKasmDockerImage(imageTag, baseImage, dockerfilePath, targetNodePath, localTarFilePath string) error {
	var tarFilePath string
	var err error

	// Step 1: Check if a local tar file is provided and exists
	if localTarFilePath != "" {
		if _, err = os.Stat(localTarFilePath); err == nil {
			// Local tar file exists, use it instead of building a new image
			tarFilePath = localTarFilePath
			log.Info().Msg("Using existing local tar file for Docker image deployment")
		} else {
			return fmt.Errorf("local tar file specified but not found: %v", err)
		}
	} else {
		// Step 2: Build the Docker image if no local file is provided
		if err = BuildCoreImageKasm(imageTag, baseImage); err != nil {
			return fmt.Errorf("failed to build Docker image: %v", err)
		}

		// Step 3: Export image to temp file
		ctx := context.Background() // Creating a context object
		retries := 3                // Set the number of retries for the export
		tarFilePath, err = dockercli.ExportImageToTar(ctx, retries, imageTag, "")
		if err != nil {
			return fmt.Errorf("failed to export Docker image to tar: %v", err)
		}
		defer func() {
			if err := os.Remove(tarFilePath); err != nil {
				log.Error().Err(err).Msg("Failed to remove tar file")
			}
		}()
	}

	// Step 4: Establish SSH connection to target node
	sshConfig := &shadowssh.SSHConfig{
		Username:          *shadowssh.SshUser,
		Password:          *shadowssh.SshPassword,
		NodeAddress:       *shadowssh.TargetNodeAddr,
		KnownHostsFile:    *shadowssh.KnownHostsFile,
		ConnectionTimeout: *shadowssh.ConnectionTimeout,
	}

	sshClient, err := shadowssh.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %v", err)
	}
	defer func() {
		if err := sshClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close SSH client")
		}
	}()

	// Step 5: Copy and load the Docker image on the remote node
	err = ImportDockerImageToRemoteNode(sshConfig.Username, sshConfig.Password, sshConfig.NodeAddress, tarFilePath, targetNodePath)
	if err != nil {
		return fmt.Errorf("failed to import Docker image on remote node: %v", err)
	}

	log.Info().Msg("Docker image deployed and loaded successfully on target node")
	return nil
}

// DeployComposeFile uploads a specified Docker Compose file and deploys the services on the target node.
func DeployComposeFile(composeFilePath, targetNodePath string) error {
	// Step 1: Establish SSH connection to target node
	sshConfig := &shadowssh.SSHConfig{
		Username:          *shadowssh.SshUser,
		Password:          *shadowssh.SshPassword,
		NodeAddress:       *shadowssh.TargetNodeAddr,
		KnownHostsFile:    *shadowssh.KnownHostsFile,
		ConnectionTimeout: *shadowssh.ConnectionTimeout,
	}

	sshClient, err := shadowssh.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %v", err)
	}
	defer func() {
		if err := sshClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close SSH client")
		}
	}()

	// Step 2: Copy compose file onto node
	log.Info().
		Str("source", composeFilePath).
		Str("destination", targetNodePath).
		Msg("Starting to copy compose file onto remote node")

	err = shadowscp.ShadowCopyFile(composeFilePath, targetNodePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("nodeAddress", sshConfig.NodeAddress).
			Str("targetPath", targetNodePath).
			Msg("Failed to copy compose file onto remote node")
		return fmt.Errorf("failed to copy compose file onto remote node: %v", err)
	}

	log.Info().
		Str("nodeAddress", sshConfig.NodeAddress).
		Str("composeFile", targetNodePath).
		Msg("Compose file copied successfully")

	// Step 3: Start compose backend
	targetNodeComposeFilePath := filepath.Join(targetNodePath, filepath.Base(composeFilePath))
	dockerComposeUpCommand := fmt.Sprintf("docker compose -f %s up -d", targetNodeComposeFilePath)

	log.Info().
		Str("command", dockerComposeUpCommand).
		Str("nodeAddress", sshConfig.NodeAddress).
		Msg("Starting docker compose on the remote node")

	// 1 Minute of log output
	output, err := shadowssh.ShadowExecuteCommandWithOutput(sshClient, dockerComposeUpCommand, 1*time.Minute)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
			Str("command", dockerComposeUpCommand).
			Str("output", output).
			Msg("Failed to start backend on remote node using compose")
		return fmt.Errorf("failed to start backend on remote node using compose: %v", err)
	}

	log.Info().
		Str("nodeAddress", sshConfig.NodeAddress).
		Msg("Backend compose deployed successfully on target node")
	return nil
}
