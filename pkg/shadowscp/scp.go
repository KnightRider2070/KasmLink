package shadowscp

import (
	"context"
	"fmt"
	"io"
	"kasmlink/pkg/shadowssh"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
)

// CopyFileToRemote copies a local file to a remote server via SFTP over SSH.
func CopyFileToRemote(ctx context.Context, localFilePath, remoteDir string, sshConfig *shadowssh.Config) error {
	log.Info().
		Str("username", sshConfig.Username).
		Str("host", sshConfig.Host).
		Int("port", sshConfig.Port).
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting file transfer to remote server via SFTP")

	retries := 3
	delay := 2 * time.Second

	for attempt := 1; attempt <= retries; attempt++ {
		err := executeFileTransfer(ctx, localFilePath, remoteDir, sshConfig)
		if err == nil {
			log.Info().Msg("File transfer completed successfully")
			return nil
		}

		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max_retries", retries).
			Dur("delay", delay).
			Msg("File transfer failed, retrying")

		if attempt < retries {
			select {
			case <-time.After(delay):
				// Continue to the next retry
			case <-ctx.Done():
				log.Error().
					Err(ctx.Err()).
					Msg("File transfer canceled due to context cancellation")
				return fmt.Errorf("file transfer canceled: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("failed to transfer file after %d retries", retries)
}

func executeFileTransfer(ctx context.Context, localFilePath, remoteDir string, sshConfig *shadowssh.Config) error {
	log.Debug().Msg("Establishing SSH connection")
	sshClient, err := shadowssh.NewClient(ctx, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := sshClient.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close SSH client")
		}
	}()
	log.Debug().Msg("SSH connection established")

	// Retrieve the actual *ssh.Client from shadowssh.NewClient
	actualSSHClient := sshClient.Client() // Assuming Client is a method returning *ssh.Client
	if actualSSHClient == nil {
		return fmt.Errorf("failed to retrieve underlying SSH client")
	}

	// Use the actual *ssh.Client
	sftpClient, err := sftp.NewClient(actualSSHClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer func() {
		if cerr := sftpClient.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close SFTP client")
		}
	}()
	log.Debug().Msg("SFTP client created successfully")

	// Construct remote file path
	remoteFilePath := fmt.Sprintf("%s/%s", remoteDir, filepath.Base(localFilePath))

	// Open local file
	log.Debug().Str("file", localFilePath).Msg("Opening local file")
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Create (or overwrite) remote file
	log.Debug().Str("remote_file", remoteFilePath).Msg("Creating remote file")
	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy local file to remote file
	log.Debug().
		Str("local_file", localFilePath).
		Str("remote_file", remoteFilePath).
		Msg("Transferring file via SFTP")

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	log.Info().
		Str("local_file", localFilePath).
		Str("remote_file", remoteFilePath).
		Msg("File transferred successfully via SFTP")
	return nil
}
