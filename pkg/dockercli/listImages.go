package dockercli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// DockerImage represents the structure of a Docker image as returned by `docker images`.
type DockerImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"id"`
	Size       string `json:"size"`
}

// ListImagesOptions defines the options for listing Docker images.
type ListImagesOptions struct {
	SSH *SSHOptions // Optional SSH configuration to list images remotely.
}

// ListImages lists all Docker images either locally or on a remote server.
func (dc *DockerClient) ListImages(ctx context.Context, options ListImagesOptions) ([]DockerImage, error) {
	if options.SSH != nil {
		return listImagesViaSSH(ctx, options.SSH)
	}
	return listImagesLocally(ctx, dc)
}

// listImagesLocally lists Docker images present locally.
func listImagesLocally(ctx context.Context, dc *DockerClient) ([]DockerImage, error) {
	log.Info().Msg("Listing Docker images locally")

	cmd := []string{"docker", "images", "--format", "{{json .}}"}

	output, err := dc.executor.Execute(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker images locally: %w", err)
	}

	return parseDockerImages(output)
}

// listImagesViaSSH lists Docker images on a remote server via SSH.
func listImagesViaSSH(ctx context.Context, sshOpts *SSHOptions) ([]DockerImage, error) {
	if sshOpts.Host == "" || sshOpts.Port == 0 || sshOpts.User == "" || sshOpts.PrivateKey == "" {
		return nil, fmt.Errorf("SSH options are incomplete")
	}

	log.Info().Str("host", sshOpts.Host).Msg("Listing Docker images via SSH")

	// Establish SSH connection.
	sshClient, err := newSSHClient(sshOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	cmd := "docker images --format '{{json .}}'"

	output, err := executeCommandOverSSH(ctx, sshClient, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker images via SSH: %w", err)
	}

	return parseDockerImages([]byte(output))
}

// parseDockerImages parses the output of the `docker images` command into a slice of DockerImage structs.
func parseDockerImages(output []byte) ([]DockerImage, error) {
	var images []DockerImage
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var image DockerImage
		if err := json.Unmarshal(line, &image); err != nil {
			log.Warn().Err(err).Msg("Failed to parse Docker image JSON")
			continue
		}
		images = append(images, image)
	}

	log.Info().Int("count", len(images)).Msg("Docker images listed successfully")
	return images, nil
}
