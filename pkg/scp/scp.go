package shadowscp

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

// ShadowCopyFile copies a local file to a remote node via SSH
func ShadowCopyFile(agentName, secretKey, nodeAddress, localFilePath, remoteDir string) error {
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
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer shadowClient.Close()

	// Create a new session for SCP
	shadowSession, err := shadowClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer shadowSession.Close()

	// Open the local file
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Get the file info (to obtain size and permissions)
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("could not stat local file: %v", err)
	}

	// Prepare the SCP command to receive the file on the remote node
	targetFileName := filepath.Base(localFilePath)
	command := fmt.Sprintf("scp -t %s/%s", remoteDir, targetFileName)

	// Set up stdin pipe to the session (for sending file metadata and contents)
	stdinPipe, err := shadowSession.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stdin for SCP: %v", err)
	}

	// Start the SCP session
	if err := shadowSession.Start(command); err != nil {
		return fmt.Errorf("failed to start SCP command: %v", err)
	}

	// Send the file metadata (size and permissions)
	fmt.Fprintf(stdinPipe, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), targetFileName)

	// Send the file contents
	_, err = io.Copy(stdinPipe, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// Signal EOF to the SCP session and close stdin
	fmt.Fprint(stdinPipe, "\x00")
	stdinPipe.Close()

	// Wait for the session to finish
	if err := shadowSession.Wait(); err != nil {
		return fmt.Errorf("failed to complete SCP session: %v", err)
	}

	fmt.Printf("Successfully copied %s to %s:%s\n", localFilePath, nodeAddress, remoteDir)
	return nil
}
