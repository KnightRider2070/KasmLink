package procedures

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockerutils"
	"os"
)

// buildCoreImageKasm orchestrates the Docker image build using the embedded Dockerfile and base image.
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
func DeployKasmDockerImage(imageTag, baseImage, dockerfilePath, targetNodeAddress, targetNodePath, sshUser, sshPassword string) error {
	// Step 1: Build the Docker image
	if err := BuildCoreImageKasm(imageTag, baseImage); err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	// Step 3: Export image to temp file
	tarFilePath, err := dockercli.ExportImageToTar(imageTag, "")
	if err != nil {
		return fmt.Errorf("failed to export Docker image to tar: %v", err)
	}
	defer os.Remove(tarFilePath) // Cleanup tar file after deployment

	// Step 3: Copy and load the Docker image on the remote node
	err = ImportDockerImageToRemoteNode(sshUser, sshPassword, targetNodeAddress, tarFilePath, targetNodePath)
	if err != nil {
		return fmt.Errorf("failed to import Docker image on remote node: %v", err)
	}

	log.Info().Msg("Docker image deployed and loaded successfully on target node")
	return nil
}
