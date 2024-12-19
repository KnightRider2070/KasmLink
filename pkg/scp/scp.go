package shadowscp

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	sshmanager "kasmlink/pkg/sshmanager"
)

// ShadowCopyFile copies a local file to a remote node via SFTP over SSH.
func ShadowCopyFile(ctx context.Context, localFilePath, remoteDir string, sshConfig *sshmanager.SSHConfig) error {
	log.Info().
		Str("username", sshConfig.Username).
		Str("host", sshConfig.Host).
		Int("port", sshConfig.Port).
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting file copy to remote node via SSH using SFTP")

	retries := 3
	delay := 2 * time.Second

	for attempt := 1; attempt <= retries; attempt++ {
		err := performSFTPCopy(ctx, localFilePath, remoteDir, sshConfig)
		if err == nil {
			log.Info().Msg("File copy completed successfully")
			return nil
		}

		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max_retries", retries).
			Dur("delay", delay).
			Msg("Failed to copy file, retrying")

		if attempt < retries {
			select {
			case <-time.After(delay):
				// Continue to the next retry
			case <-ctx.Done():
				log.Error().
					Err(ctx.Err()).
					Msg("File copy canceled due to context cancellation")
				return fmt.Errorf("file copy canceled: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("failed to copy file after %d retries", retries)
}

func performSFTPCopy(ctx context.Context, localFilePath, remoteDir string, sshConfig *sshmanager.SSHConfig) error {
	log.Debug().Msg("Establishing SSH connection")
	sshClient, err := sshmanager.NewSSHClient(ctx, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := sshClient.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close SSH client")
		}
	}()
	log.Debug().Msg("SSH connection established")

	client := sshClient.GetClient()
	if client == nil {
		return fmt.Errorf("SSH client is nil")
	}

	// Create SFTP client
	log.Debug().Msg("Creating SFTP client")
	sftpClient, err := sftp.NewClient(client)
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
	remoteFilePath := remoteDir + "/" + fileNameFromPath(localFilePath)

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
		Msg("Copying file via SFTP")

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	log.Info().
		Str("local_file", localFilePath).
		Str("remote_file", remoteFilePath).
		Msg("File transferred successfully via SFTP")
	return nil
}

func fileNameFromPath(path string) string {
	// Simple helper to extract filename from a path
	// without adding extra dependencies.
	i := len(path) - 1
	for i >= 0 && (path[i] == '/' || path[i] == '\\') {
		i--
	}
	if i < 0 {
		return ""
	}
	start := i
	for start >= 0 && path[start] != '/' && path[start] != '\\' {
		start--
	}
	return path[start+1:]
}
