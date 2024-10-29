package shadowscp

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

// ShadowCopyFile copies a local file to a remote node via SSH.
func ShadowCopyFile(agentName, secretKey, nodeAddress, localFilePath, remoteDir string) error {
	log.Info().
		Str("agent_name", agentName).
		Str("node_address", nodeAddress).
		Str("local_file", localFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting file copy to remote node via SSH")

	// Create the SSH client configuration
	config := &ssh.ClientConfig{
		User: agentName,
		Auth: []ssh.AuthMethod{
			ssh.Password(secretKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the remote node
	shadowClient, err := ssh.Dial("tcp", nodeAddress, config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer shadowClient.Close()
	log.Debug().Msg("SSH connection established")

	// Create a new session for SCP
	shadowSession, err := shadowClient.NewSession()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create SSH session")
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer shadowSession.Close()
	log.Debug().Msg("SSH session created")

	// Open the local file
	localFile, err := os.Open(localFilePath)
	if err != nil {
		log.Error().Err(err).Str("local_file", localFilePath).Msg("Failed to open local file")
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Get the file info (to obtain size and permissions)
	fileInfo, err := localFile.Stat()
	if err != nil {
		log.Error().Err(err).Str("local_file", localFilePath).Msg("Failed to get file info")
		return fmt.Errorf("could not stat local file: %v", err)
	}
	log.Debug().Int64("file_size", fileInfo.Size()).Str("permissions", fileInfo.Mode().String()).Msg("Local file details retrieved")

	// Prepare the SCP command to receive the file on the remote node
	targetFileName := filepath.Base(localFilePath)
	command := fmt.Sprintf("scp -t %s/%s", remoteDir, targetFileName)

	// Set up stdin pipe to the session (for sending file metadata and contents)
	stdinPipe, err := shadowSession.StdinPipe()
	if err != nil {
		log.Error().Err(err).Msg("Failed to set up stdin pipe for SCP")
		return fmt.Errorf("failed to set up stdin for SCP: %v", err)
	}

	// Start the SCP session
	if err := shadowSession.Start(command); err != nil {
		log.Error().Err(err).Str("command", command).Msg("Failed to start SCP command")
		return fmt.Errorf("failed to start SCP command: %v", err)
	}
	log.Info().Str("command", command).Msg("SCP command started on remote node")

	// Send the file metadata (size and permissions)
	fmt.Fprintf(stdinPipe, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), targetFileName)
	log.Debug().Str("target_file_name", targetFileName).Msg("Sent file metadata")

	// Send the file contents
	_, err = io.Copy(stdinPipe, localFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy file contents")
		return fmt.Errorf("failed to copy file contents: %v", err)
	}
	log.Debug().Msg("File contents sent successfully")

	// Signal EOF to the SCP session and close stdin
	fmt.Fprint(stdinPipe, "\x00")
	stdinPipe.Close()
	log.Debug().Msg("EOF signal sent, closing stdin")

	// Wait for the session to finish
	if err := shadowSession.Wait(); err != nil {
		log.Error().Err(err).Msg("SCP session failed to complete")
		return fmt.Errorf("failed to complete SCP session: %v", err)
	}

	log.Info().
		Str("local_file", localFilePath).
		Str("node_address", nodeAddress).
		Str("remote_dir", remoteDir).
		Msg("File copied successfully to remote node")
	return nil
}
