package dockercli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

// BuildImageOptions defines the options for building a Docker image.
type BuildImageOptions struct {
	ContextDir     string
	DockerfilePath string
	ImageTag       string
	BuildArgs      map[string]string
	SSH            *SSHOptions
}

// SSHOptions defines the configuration for executing commands over SSH.
type SSHOptions struct {
	Host     string
	Port     int
	User     string
	Password string
}

// BuildImage builds a Docker image either locally or via SSH.
func (dc *DockerClient) BuildImage(ctx context.Context, options BuildImageOptions) error {
	// Validate required options.
	if err := validateBuildOptions(options); err != nil {
		return err
	}

	log.Info().Str("imageTag", options.ImageTag).Msg("Starting Docker image build")

	// Create a tarball of the build context.
	tarballPath, err := createTarball(options.ContextDir)
	if err != nil {
		return fmt.Errorf("failed to create tarball of build context: %w", err)
	}
	defer os.Remove(tarballPath) // Cleanup after use.

	// Determine the build method based on the SSH configuration.
	if options.SSH != nil {
		return dc.buildImageViaSSH(ctx, tarballPath, options)
	}
	return dc.buildImageLocally(ctx, tarballPath, options)
}

// validateBuildOptions ensures all required options are provided.
func validateBuildOptions(options BuildImageOptions) error {
	if options.ContextDir == "" {
		return fmt.Errorf("context directory is required")
	}
	if options.DockerfilePath == "" {
		return fmt.Errorf("Dockerfile path is required")
	}
	if options.ImageTag == "" {
		return fmt.Errorf("image tag is required")
	}
	if options.SSH != nil {
		if options.SSH.Host == "" || options.SSH.Port == 0 || options.SSH.User == "" || options.SSH.PrivateKey == "" {
			return fmt.Errorf("incomplete SSH options")
		}
	}
	return nil
}

// buildImageLocally builds the Docker image locally.
func (dc *DockerClient) buildImageLocally(ctx context.Context, tarballPath string, options BuildImageOptions) error {
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
	output, err := dc.executor.Execute(ctx, buildCmd[0], buildCmd[1:]...)
	if err != nil {
		return fmt.Errorf("failed to build Docker image locally: %w", err)
	}

	// Use PrintBuildLogs to process and format logs
	logReader := bytes.NewReader(output)
	if err := PrintBuildLogs(logReader); err != nil {
		return fmt.Errorf("failed to process build logs: %w", err)
	}

	return nil
}

// buildImageViaSSH builds the Docker image on a remote server via SSH.
func (dc *DockerClient) buildImageViaSSH(ctx context.Context, tarballPath string, options BuildImageOptions) error {
	log.Info().Str("host", options.SSH.Host).Msg("Building Docker image via SSH")

	// Establish an SSH connection.
	sshClient, err := dc.sshClientFactory(ctx, options.SSH)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer sshClient.Close()

	// Transfer the tarball via SFTP.
	remoteTarballPath := "/tmp/build-context.tar"
	if err := transferFileViaSFTP(ctx, sshClient, tarballPath, remoteTarballPath); err != nil {
		return fmt.Errorf("failed to transfer tarball to remote server: %w", err)
	}

	// Execute the Docker build command on the remote server.
	buildCmd := fmt.Sprintf(
		"docker build -t %s -f %s /tmp",
		options.ImageTag, filepath.Base(options.DockerfilePath),
	)
	output, err := executeCommandOverSSH(ctx, sshClient, buildCmd)
	if err != nil {
		return fmt.Errorf("failed to build Docker image via SSH: %w", err)
	}

	// Use PrintBuildLogs to process and format logs
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

// executeCommandOverSSH executes a command on a remote server via SSH.
func executeCommandOverSSH(ctx context.Context, client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// newSSHClient creates a new SSH client connection.
func newSSHClient(opts *SSHOptions) (*ssh.Client, error) {
	key, err := os.ReadFile(opts.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: opts.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Change for production environments.
	}

	address := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	return client, nil
}

// createTarball uses CreateTarWithContext to generate a tarball for the build context and writes it to a temporary file.
func createTarball(contextDir string) (string, error) {
	if contextDir == "" {
		return "", fmt.Errorf("context directory cannot be empty")
	}

	tarReader, err := CreateTarWithContext(contextDir)
	if err != nil {
		return "", fmt.Errorf("failed to create tarball from build context: %w", err)
	}

	tempFile, err := os.CreateTemp("", "build-context-*.tar")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for tarball: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, tarReader); err != nil {
		return "", fmt.Errorf("failed to write tarball to file: %w", err)
	}

	return tempFile.Name(), nil
}
