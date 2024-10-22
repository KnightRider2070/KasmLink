package shadowssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
)

// ShadowExecuteCommand connects to a remote node via SSH and executes the specified command
func ShadowExecuteCommand(agentName, secretKey, nodeAddress, shadowCommand string) error {
	// Define SSH configuration
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

	// Create a session
	shadowSession, err := shadowClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer shadowSession.Close()

	// Execute the command
	stdout, err := shadowSession.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Run the command on the remote node
	if err := shadowSession.Start(shadowCommand); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	// Print the output
	io.Copy(os.Stdout, stdout)

	// Wait for the command to finish
	if err := shadowSession.Wait(); err != nil {
		return fmt.Errorf("failed to wait for command to finish: %v", err)
	}

	fmt.Printf("Command '%s' executed successfully on %s\n", shadowCommand, nodeAddress)
	return nil
}
