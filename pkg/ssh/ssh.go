package shadowssh

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
)

// ShadowExecuteCommand connects to a remote node via SSH and executes the specified command.
func ShadowExecuteCommand(agentName, secretKey, nodeAddress, shadowCommand string) error {
	log.Info().
		Str("agent_name", agentName).
		Str("node_address", nodeAddress).
		Str("command", shadowCommand).
		Msg("Starting remote command execution via SSH")

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
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer shadowClient.Close()
	log.Debug().Msg("SSH connection established")

	// Create a session
	shadowSession, err := shadowClient.NewSession()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create SSH session")
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer shadowSession.Close()
	log.Debug().Msg("SSH session created")

	// Obtain stdout pipe
	stdout, err := shadowSession.StdoutPipe()
	if err != nil {
		log.Error().Err(err).Msg("Failed to set up stdout pipe")
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// Run the command on the remote node
	if err := shadowSession.Start(shadowCommand); err != nil {
		log.Error().Err(err).Str("command", shadowCommand).Msg("Failed to start command execution")
		return fmt.Errorf("failed to run command: %v", err)
	}
	log.Info().Str("command", shadowCommand).Msg("Command execution started on remote node")

	// Print the command output to the console
	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		log.Error().Err(err).Msg("Failed to read command output")
		return fmt.Errorf("failed to read command output: %v", err)
	}
	log.Debug().Msg("Command output successfully copied to stdout")

	// Wait for the command to finish
	if err := shadowSession.Wait(); err != nil {
		log.Error().Err(err).Msg("Command execution did not complete successfully")
		return fmt.Errorf("failed to wait for command to finish: %v", err)
	}

	log.Info().
		Str("command", shadowCommand).
		Str("node_address", nodeAddress).
		Msg("Command executed successfully on remote node")
	return nil
}
