// shadowscp/shadowscp.go
package shadowscp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	sshmanager "kasmlink/pkg/sshmanager"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// ShadowCopyFile copies a local file to a remote node via SSH using a custom SCP implementation.
// It includes retry logic and structured logging.
func ShadowCopyFile(ctx context.Context, localFilePath, remoteDir string, sshConfig *sshmanager.SSHConfig) error {
	log.Info().
		Str("username", sshConfig.Username).
		Str("host", sshConfig.Host).
		Int("port", sshConfig.Port).
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting file copy to remote node via SSH using custom SCP")

	retries := 3
	delay := 2 * time.Second

	for attempt := 1; attempt <= retries; attempt++ {
		err := performCopy(ctx, localFilePath, remoteDir, sshConfig)
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
				// Continue to next retry
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

// performCopy handles the actual file copy using a custom SCP implementation.
func performCopy(ctx context.Context, localFilePath, remoteDir string, sshConfig *sshmanager.SSHConfig) error {
	// Establish SSH connection using sshmanager
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

	// Retrieve the underlying ssh.Client
	client := sshClient.GetClient()
	if client == nil {
		return fmt.Errorf("SSH client is nil")
	}

	// Create a new SSH session
	session, err := client.NewSession()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create SSH session")
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		if cerr := session.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close SSH session")
		}
	}()

	// Prepare the remote SCP command
	// -t indicates the server is ready to receive files
	cmd := fmt.Sprintf("scp -t %s", remoteDir)
	if err := session.Start(cmd); err != nil {
		log.Error().
			Err(err).
			Str("command", cmd).
			Msg("Failed to start SCP command")
		return fmt.Errorf("failed to start SCP command: %w", err)
	}

	// Get the standard input and output of the session
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get stdin pipe")
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get stdout pipe")
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Create a buffered reader for stdout to read server responses
	reader := bufio.NewReader(stdout)

	// Start a goroutine to read server responses
	responseChan := make(chan error, 1)
	go func() {
		resp, err := reader.ReadByte()
		if err != nil {
			responseChan <- fmt.Errorf("failed to read server response: %w", err)
			return
		}
		if resp != 0 {
			// Read the error message
			errMsg, err := reader.ReadString('\n')
			if err != nil {
				responseChan <- fmt.Errorf("failed to read error message: %w", err)
				return
			}
			responseChan <- fmt.Errorf("server error: %s", errMsg)
			return
		}
		responseChan <- nil
	}()

	// Send file metadata
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("local_file", localFilePath).
			Msg("Failed to stat local file")
		return fmt.Errorf("could not stat local file: %w", err)
	}

	// Prepare the SCP protocol header
	// C0644 <size> <filename>\n
	fileMode := fileInfo.Mode().Perm()
	fileSize := fileInfo.Size()
	fileName := filepath.Base(localFilePath)
	header := fmt.Sprintf("C%04o %d %s\n", fileMode, fileSize, fileName)

	if _, err := io.WriteString(stdin, header); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send file metadata to SCP")
		return fmt.Errorf("failed to send file metadata: %w", err)
	}

	// Wait for server acknowledgment
	if err := <-responseChan; err != nil {
		log.Error().
			Err(err).
			Msg("Server responded with an error after sending metadata")
		return err
	}

	// Open the local file for reading
	file, err := os.Open(localFilePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("local_file", localFilePath).
			Msg("Failed to open local file for SCP")
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Copy the file content to stdin
	if _, err := io.Copy(stdin, file); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send file content to SCP")
		return fmt.Errorf("failed to send file content: %w", err)
	}

	// Send null byte to indicate end of file transfer
	if _, err := stdin.Write([]byte{0}); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send end of file indicator to SCP")
		return fmt.Errorf("failed to send end of file indicator: %w", err)
	}

	// Read server acknowledgment for file content
	resp, err := reader.ReadByte()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to read server response after file content")
		return fmt.Errorf("failed to read server response: %w", err)
	}
	if resp != 0 {
		// Read the error message
		errMsg, err := reader.ReadString('\n')
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to read error message after file content")
			return fmt.Errorf("failed to read error message: %w", err)
		}
		log.Error().
			Str("error_message", errMsg).
			Msg("Server responded with an error after sending file content")
		return fmt.Errorf("server error after file content: %s", errMsg)
	}

	// Send end of transfer
	if _, err := io.WriteString(stdin, "E\n"); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send end of transfer indicator to SCP")
		return fmt.Errorf("failed to send end of transfer indicator: %w", err)
	}

	// Wait for final server acknowledgment
	finalResp, err := reader.ReadByte()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to read final server response")
		return fmt.Errorf("failed to read final server response: %w", err)
	}
	if finalResp != 0 {
		// Read the error message
		errMsg, err := reader.ReadString('\n')
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to read final error message from SCP server")
			return fmt.Errorf("failed to read final error message: %w", err)
		}
		log.Error().
			Str("error_message", errMsg).
			Msg("Server responded with an error at end of transfer")
		return fmt.Errorf("server error at end of transfer: %s", errMsg)
	}

	// Wait for the SCP command to complete
	if err := session.Wait(); err != nil {
		log.Error().
			Err(err).
			Msg("SCP session encountered an error")
		return fmt.Errorf("SCP session error: %w", err)
	}

	log.Debug().
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("File transferred successfully via SCP")

	return nil
}
