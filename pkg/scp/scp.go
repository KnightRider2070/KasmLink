package shadowscp

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"kasmlink/pkg/ssh" // Using the shadowssh package for enhanced SSH capabilities
	"os"
	"path/filepath"
	"time"
)

// ShadowCopyFile copies a local file to a remote node via SSH with enhanced features.
func ShadowCopyFile(localFilePath, remoteDir string) error {
	// Create SSHConfig using CLI flags or default values
	sshConfig := shadowssh.NewSSHConfigFromFlags()

	log.Info().
		Str("username", sshConfig.Username).
		Str("node_address", sshConfig.NodeAddress).
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting file copy to remote node via SSH")

	retries := 3
	var err error

	for retries > 0 {
		err = performCopy(sshConfig, localFilePath, remoteDir)
		if err == nil {
			log.Info().Msg("File copy completed successfully")
			return nil
		}

		retries--
		log.Warn().
			Err(err).
			Int("retries_left", retries).
			Msg("Failed to copy file, retrying")
		time.Sleep(2 * time.Second)
	}

	log.Error().Err(err).Msg("File copy failed after all retries")
	return fmt.Errorf("failed to copy file after retries: %w", err)
}

// performCopy performs the actual file copy to the remote node via SSH.
func performCopy(sshConfig *shadowssh.SSHConfig, localFilePath, remoteDir string) error {
	// Establish SSH connection
	client, err := shadowssh.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer client.Close()
	log.Debug().Msg("SSH connection established")

	// Create a new session for SCP
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()
	log.Debug().Msg("SSH session created")

	// Open the local file
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get the file info (to obtain size and permissions)
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("could not stat local file: %w", err)
	}
	log.Debug().
		Int64("file_size", fileInfo.Size()).
		Str("permissions", fileInfo.Mode().String()).
		Msg("Local file details retrieved")

	// Prepare the SCP command to receive the file on the remote node
	targetFileName := filepath.Base(localFilePath)
	command := fmt.Sprintf("scp -t %s/%s", remoteDir, targetFileName)

	// Set up stdin pipe to the session (for sending file metadata and contents)
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stdin for SCP: %w", err)
	}
	defer stdinPipe.Close()

	// Start the SCP session
	if err := session.Start(command); err != nil {
		return fmt.Errorf("failed to start SCP command: %w", err)
	}
	log.Info().Str("command", command).Msg("SCP command started on remote node")

	// Send the file metadata (size and permissions)
	fmt.Fprintf(stdinPipe, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), targetFileName)
	log.Debug().Str("target_file_name", targetFileName).Msg("Sent file metadata")

	// Send the file contents with progress logging
	buffer := make([]byte, 4096)
	totalBytesCopied := int64(0)
	for {
		n, readErr := localFile.Read(buffer)
		if n > 0 {
			if _, writeErr := stdinPipe.Write(buffer[:n]); writeErr != nil {
				log.Error().Err(writeErr).Msg("Failed to write file contents to stdin pipe")
				return fmt.Errorf("failed to write file contents: %w", writeErr)
			}
			totalBytesCopied += int64(n)
			log.Debug().
				Int64("bytes_copied", totalBytesCopied).
				Float64("progress", float64(totalBytesCopied)/float64(fileInfo.Size())*100).
				Msg("Copying file in progress")
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			log.Error().Err(readErr).Msg("Failed to read from local file")
			return fmt.Errorf("failed to read file contents: %w", readErr)
		}
	}

	// Signal EOF to the SCP session and close stdin
	fmt.Fprint(stdinPipe, "\x00")
	log.Debug().Msg("EOF signal sent")

	// Wait for the session to finish
	if err := session.Wait(); err != nil {
		return fmt.Errorf("failed to complete SCP session: %w", err)
	}

	log.Info().
		Str("local_file", localFilePath).
		Str("node_address", sshConfig.NodeAddress).
		Str("remote_dir", remoteDir).
		Msg("File copied successfully to remote node")
	return nil
}
