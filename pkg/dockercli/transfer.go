package dockercli

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/shadowscp"
	"kasmlink/pkg/shadowssh"
	"os"
	"path/filepath"
)

// TransferImage transfers a Docker image to a remote server via SSH.
func (dc *DockerClient) TransferImage(ctx context.Context, image string, sshOptions *SSHOptions) error {
	log.Info().Str("image", image).Msg("Starting transfer of Docker image")

	// Create tarball of the image.
	tarPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.tar", filepath.Base(image)))
	if err := dc.SaveImage(ctx, image, tarPath); err != nil {
		log.Error().Err(err).Msg("Failed to save Docker image to tarball")
		return fmt.Errorf("failed to save Docker image %s: %w", image, err)
	}
	defer os.Remove(tarPath) // Clean up temporary tarball file.

	// Establish SSH connection and transfer tarball.
	sshClient, err := shadowssh.NewClient(ctx, &shadowssh.Config{
		Host: sshOptions.Host, Port: sshOptions.Port, Username: sshOptions.User, Password: sshOptions.Password})
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	remoteTarPath := fmt.Sprintf("/tmp/%s.tar", filepath.Base(image))
	if err := shadowscp.CopyFileToRemote(ctx, tarPath, remoteTarPath, &shadowssh.Config{
		Host: sshOptions.Host, Port: sshOptions.Port, Username: sshOptions.User, Password: sshOptions.Password}); err != nil {
		log.Error().Err(err).Msg("Failed to transfer image tarball to remote server")
		return fmt.Errorf("failed to transfer image tarball to remote server: %w", err)
	}

	// Load the image on the remote server.
	loadCommand := fmt.Sprintf("docker load -i %s", remoteTarPath)
	if output, err := sshClient.ExecuteCommand(ctx, loadCommand); err != nil {
		log.Error().Err(err).Str("output", output).Msg("Failed to load Docker image on remote server")
		return fmt.Errorf("failed to load Docker image on remote server: %w", err)
	}

	log.Info().Str("image", image).Msg("Docker image successfully transferred and loaded")
	return nil
}
