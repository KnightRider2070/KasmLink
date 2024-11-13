package shadowssh

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig holds SSH connection parameters provided via CLI
var (
	SshUser           = flag.String("ssh-user", "", "Username for SSH connection")
	SshPassword       = flag.String("ssh-password", "", "Password for SSH connection")
	TargetNodeAddr    = flag.String("target-node", "", "Target node address for SSH connection")
	KnownHostsFile    = flag.String("known-hosts", "~/.ssh/known_hosts", "Path to known_hosts file")
	ConnectionTimeout = flag.Duration("connection-timeout", 10*time.Second, "SSH connection timeout")
)

// SSHConfig holds the necessary information for SSH authentication and host verification.
type SSHConfig struct {
	Username          string
	Password          string
	NodeAddress       string
	KnownHostsFile    string
	ConnectionTimeout time.Duration
}

// NewSSHConfigFromFlags returns an SSHConfig struct populated with values from the CLI flags.
func NewSSHConfigFromFlags() *SSHConfig {
	return &SSHConfig{
		Username:          *SshUser,
		Password:          *SshPassword,
		NodeAddress:       *TargetNodeAddr,
		KnownHostsFile:    *KnownHostsFile,
		ConnectionTimeout: *ConnectionTimeout,
	}
}

// NewSSHClient establishes an SSH connection using the provided configuration.
func NewSSHClient(config *SSHConfig) (*ssh.Client, error) {
	// Configure host key verification from the known hosts file.
	hostKeyCallback, err := knownhosts.New(config.KnownHostsFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load known hosts for SSH verification")
		return nil, fmt.Errorf("failed to load known hosts: %v", err)
	}

	// Configure SSH client authentication and connection.
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         config.ConnectionTimeout,
	}

	// Connect to the SSH server.
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", config.NodeAddress), sshConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to establish SSH connection")
		return nil, fmt.Errorf("failed to dial SSH: %v", err)
	}

	log.Debug().Msg("SSH connection established")
	return client, nil
}

// ShadowExecuteCommandWithOutput executes a command over SSH and logs the output for a specified duration.
func ShadowExecuteCommandWithOutput(client *ssh.Client, command string, logDuration time.Duration) (string, error) {
	// Create a new session for the command.
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close SSH session")
		}
	}()

	// Get pipes for standard output and error.
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command.
	if err := session.Start(command); err != nil {
		return "", fmt.Errorf("failed to start command: %v", err)
	}

	// Channels for real-time logging and capturing output.
	outputChan := make(chan string)
	errChan := make(chan error)
	go func() {
		defer close(outputChan)

		// Rename the variable to avoid conflict with the `scanner` package
		outputScanner := bufio.NewScanner(io.MultiReader(stdoutPipe, stderrPipe))
		for outputScanner.Scan() {
			outputChan <- outputScanner.Text()
		}
		if err := outputScanner.Err(); err != nil && err != io.EOF {
			errChan <- fmt.Errorf("error reading output: %v", err)
		}
		close(errChan)
	}()

	// Log command output in real-time for the specified duration.
	var outputBuffer string
	logTimer := time.NewTimer(logDuration)
	log.Info().Str("command", command).Msg("Logging command output")

	for {
		select {
		case output, ok := <-outputChan:
			if !ok {
				log.Info().Msg("Command output completed")
				return outputBuffer, session.Wait()
			}
			log.Info().Str("output", output).Msg("Command output")
			outputBuffer += output + "\n"
		case err := <-errChan:
			if err != nil {
				log.Error().Err(err).Msg("Error reading command output")
				return "", err
			}
		case <-logTimer.C:
			log.Info().Msg("Logging duration expired; capturing remaining output")
			for output := range outputChan {
				outputBuffer += output + "\n"
			}
			return outputBuffer, session.Wait()
		}
	}
}

// ExecuteCommand connects to a remote node via SSH, executes a command, and returns the combined stdout and stderr output.
func ExecuteCommand(client *ssh.Client, command string) (string, error) {
	// Create a new session for the command
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close SSH session")
		}
	}()

	// Capture both stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Execute the command
	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("command execution failed: %v, stderr: %s", err, stderrBuf.String())
	}

	// Return combined output
	return stdoutBuf.String(), nil
}
