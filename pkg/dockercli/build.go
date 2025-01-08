package dockercli

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"kasmlink/pkg/shadowssh"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
)

// BuildImageOptions defines the options for building a Docker image.
type BuildImageOptions struct {
	ContextDir     string
	DockerfilePath string
	ImageTag       string
	BuildArgs      map[string]string
	SSH            *shadowssh.Config
}

// validateBuildOptions ensures all required options are provided.
func validateBuildOptions(options BuildImageOptions) error {
	if options.ContextDir == "" {
		return fmt.Errorf("context directory is required")
	}
	if options.DockerfilePath == "" {
		return fmt.Errorf("dockerfile path is required")
	}
	if options.ImageTag == "" {
		return fmt.Errorf("image tag is required")
	}
	if options.SSH != nil {
		if options.SSH.Host == "" || options.SSH.Port == 0 || options.SSH.Username == "" || options.SSH.Password == "" {
			return fmt.Errorf("incomplete SSH options")
		}
	}
	return nil
}

// buildImageLocally builds the Docker image locally.
func buildImageLocally(ctx context.Context, client *DockerClient, tarballPath string, options BuildImageOptions) error {
	buildCmd := []string{
		"docker", "build",
		"-t", options.ImageTag,
		"-f", options.DockerfilePath,
		options.ContextDir,
	}

	// Append build arguments.
	for key, value := range options.BuildArgs {
		buildCmd = append(buildCmd, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Execute the Docker build command.
	output, err := client.executor.Execute(ctx, buildCmd[0], buildCmd[1:]...)
	if err != nil {
		return fmt.Errorf("failed to build Docker image locally: %w", err)
	}

	// Use PrintBuildLogs to process and format logs.
	logReader := bytes.NewReader(output)
	if err := PrintBuildLogs(logReader); err != nil {
		return fmt.Errorf("failed to process build logs: %w", err)
	}

	return nil
}

// buildImageViaSSH builds the Docker image on a remote server via SSH.
func buildImageViaSSH(ctx context.Context, client *DockerClient, tarballPath string, options BuildImageOptions) error {
	log.Info().Str("host", options.SSH.Host).Msg("Building Docker image via SSH")

	// Establish an SSH connection.
	sshClient, err := client.sshClientFactory(ctx, options.SSH)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	// Transfer the tarball via SFTP.
	remoteTarballPath := "/tmp/build-context.tar"
	if err := transferFileViaSFTP(ctx, sshClient.Client(), tarballPath, remoteTarballPath); err != nil {
		return fmt.Errorf("failed to transfer tarball to remote server: %w", err)
	}

	// Execute the Docker build command on the remote server.
	buildCmd := fmt.Sprintf(
		"docker build -t %s -f %s /tmp",
		options.ImageTag, filepath.Base(options.DockerfilePath),
	)
	output, err := sshClient.ExecuteCommand(ctx, buildCmd)
	if err != nil {
		return fmt.Errorf("failed to build Docker image via SSH: %w", err)
	}

	// Use PrintBuildLogs to process and format logs.
	logReader := bytes.NewReader([]byte(output))
	if err := PrintBuildLogs(logReader); err != nil {
		return fmt.Errorf("failed to process build logs: %w", err)
	}

	return nil
}

// transferFileViaSFTP transfers a file to a remote server using SFTP.
func transferFileViaSFTP(ctx context.Context, client *ssh.Client, localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file to remote server: %w", err)
	}

	log.Info().Str("remote_file", remotePath).Msg("File transferred successfully via SFTP")
	return nil
}

// BuildImage builds a Docker image either locally or via SSH.
func BuildImage(ctx context.Context, client *DockerClient, options BuildImageOptions) error {
	// Validate required options.
	if err := validateBuildOptions(options); err != nil {
		return err
	}

	log.Info().Str("imageTag", options.ImageTag).Msg("Starting Docker image build")

	// Create a tarball of the build context.
	tarballReader, err := client.CreateTarWithContext(options.ContextDir)
	if err != nil {
		return fmt.Errorf("failed to create tarball of build context: %w", err)
	}
	tarballPath := ""
	if tarballReader != nil {
		tempFile, err := os.CreateTemp("", "build-context-*.tar")
		if err != nil {
			return fmt.Errorf("failed to create temporary tarball file: %w", err)
		}
		defer os.Remove(tempFile.Name())
		if _, err := io.Copy(tempFile, tarballReader); err != nil {
			return fmt.Errorf("failed to write tarball to temporary file: %w", err)
		}
		tarballPath = tempFile.Name()
	}

	// Determine the build method based on the SSH configuration.
	if options.SSH != nil {
		return buildImageViaSSH(ctx, client, tarballPath, options)
	}
	return buildImageLocally(ctx, client, tarballPath, options)
}
