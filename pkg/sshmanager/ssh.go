package shadowssh

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os/user"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHConfig holds SSH connection parameters.
type SSHConfig struct {
	Username          string
	Password          string
	Host              string
	Port              int
	KnownHostsFile    string
	ConnectionTimeout time.Duration
}

// SSHClient manages the SSH client connection.
type SSHClient struct {
	client *ssh.Client
	config SSHConfig
}

// NewSSHConfig initializes and validates an SSHConfig struct.
// It expands the tilde (~) in the KnownHostsFile path if present.
func NewSSHConfig(username, password, host string, port int, knownHostsFile string, timeout time.Duration) (*SSHConfig, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}

	// Expand ~ to home directory if present
	if len(knownHostsFile) >= 2 && knownHostsFile[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		knownHostsFile = filepath.Join(usr.HomeDir, knownHostsFile[2:])
	}

	return &SSHConfig{
		Username:          username,
		Password:          password,
		Host:              host,
		Port:              port,
		KnownHostsFile:    knownHostsFile,
		ConnectionTimeout: timeout,
	}, nil
}

// NewSSHClient establishes an SSH connection using the provided configuration.
// It returns an SSHClient instance or an error if the connection fails.
func NewSSHClient(ctx context.Context, config *SSHConfig) (*SSHClient, error) {
	if config == nil {
		return nil, fmt.Errorf("SSHConfig cannot be nil")
	}

	// Configure host key verification using the known_hosts file.
	hostKeyCallback, err := knownhosts.New(config.KnownHostsFile)
	if err != nil {
		log.Error().
			Err(err).
			Str("known_hosts_file", config.KnownHostsFile).
			Msg("Failed to load known hosts for SSH verification")
		return nil, fmt.Errorf("failed to load known hosts: %w", err)
	}

	// Set up SSH client configuration.
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Password)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         config.ConnectionTimeout,
	}

	// Build the network address with host and port.
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Establish SSH connection respecting the context.
	dialer := &netDialer{ctx: ctx, timeout: config.ConnectionTimeout}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		log.Error().
			Err(err).
			Str("address", address).
			Msg("Failed to dial SSH")
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	// Perform the SSH handshake.
	clientConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("address", address).
			Msg("Failed to establish SSH connection")
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	client := ssh.NewClient(clientConn, chans, reqs)

	log.Debug().
		Str("address", address).
		Msg("SSH connection established")

	return &SSHClient{
		client: client,
		config: *config,
	}, nil
}

// GetClient returns the underlying ssh.Client.
// This is useful for integrating with other SSH-based libraries.
func (c *SSHClient) GetClient() *ssh.Client {
	return c.client
}

// Close gracefully closes the SSH client connection.
// It logs any errors encountered during closure.
func (c *SSHClient) Close() error {
	if c.client != nil {
		err := c.client.Close()
		if err != nil {
			log.Error().
				Err(err).
				Str("address", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)).
				Msg("Failed to close SSH connection")
			return fmt.Errorf("failed to close SSH connection: %w", err)
		}
		log.Debug().
			Str("address", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)).
			Msg("SSH connection closed successfully")
	}
	return nil
}

// ExecuteCommandWithOutput executes a command over SSH and logs the output in real-time for a specified duration.
// It returns the combined output from stdout and stderr.
func (c *SSHClient) ExecuteCommandWithOutput(ctx context.Context, command string, logDuration time.Duration) (string, error) {
	// Create a new session for the command.
	session, err := c.client.NewSession()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to create SSH session")
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		if cerr := session.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Str("command", command).
				Msg("Failed to close SSH session")
		}
	}()

	// Get pipes for standard output and error.
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to get stdout pipe")
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to get stderr pipe")
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command.
	if err := session.Start(command); err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to start command")
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Channels for real-time logging and capturing output.
	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(outputChan)
		combinedReader := io.MultiReader(stdoutPipe, stderrPipe)
		scanner := bufio.NewScanner(combinedReader)
		for scanner.Scan() {
			outputChan <- scanner.Text()
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
			errChan <- fmt.Errorf("error reading output: %w", err)
		}
		close(errChan)
	}()

	// Log command output in real-time for the specified duration.
	var outputBuffer string
	logTimer := time.NewTimer(logDuration)
	defer logTimer.Stop()

	log.Info().
		Str("command", command).
		Dur("log_duration", logDuration).
		Msg("Logging command output")

	for {
		select {
		case output, ok := <-outputChan:
			if !ok {
				log.Info().Msg("Command output completed")
				if err := session.Wait(); err != nil {
					log.Error().
						Err(err).
						Str("command", command).
						Msg("Command execution failed")
					return outputBuffer, fmt.Errorf("command execution failed: %w", err)
				}
				return outputBuffer, nil
			}
			log.Info().
				Str("output", output).
				Msg("Command output")
			outputBuffer += output + "\n"
		case err := <-errChan:
			if err != nil {
				log.Error().
					Err(err).
					Str("command", command).
					Msg("Error reading command output")
				return outputBuffer, err
			}
		case <-logTimer.C:
			log.Info().Msg("Logging duration expired; capturing remaining output")
			for output := range outputChan {
				outputBuffer += output + "\n"
			}
			if err := session.Wait(); err != nil {
				log.Error().
					Err(err).
					Str("command", command).
					Msg("Command execution failed after log duration")
				return outputBuffer, fmt.Errorf("command execution failed: %w", err)
			}
			return outputBuffer, nil
		case <-ctx.Done():
			log.Warn().
				Err(ctx.Err()).
				Str("command", command).
				Msg("Context canceled; terminating command execution")
			if err := session.Signal(ssh.SIGINT); err != nil {
				log.Error().
					Err(err).
					Str("command", command).
					Msg("Failed to send interrupt signal to SSH session")
			}
			return outputBuffer, ctx.Err()
		}
	}
}

// ExecuteCommand connects to a remote node via SSH, executes a command, and returns the combined stdout and stderr output.
// It respects the provided context for cancellation and timeout.
func (c *SSHClient) ExecuteCommand(ctx context.Context, command string) (string, error) {
	// Create a new session for the command.
	session, err := c.client.NewSession()
	if err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to create SSH session")
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() {
		if cerr := session.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Str("command", command).
				Msg("Failed to close SSH session")
		}
	}()

	// Capture both stdout and stderr.
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Start the command.
	if err := session.Start(command); err != nil {
		log.Error().
			Err(err).
			Str("command", command).
			Msg("Failed to start command")
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Channel to wait for the command to finish.
	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context canceled or timed out.
		log.Warn().
			Err(ctx.Err()).
			Str("command", command).
			Msg("Context canceled; terminating command execution")
		if err := session.Signal(ssh.SIGINT); err != nil {
			log.Error().
				Err(err).
				Str("command", command).
				Msg("Failed to send interrupt signal to SSH session")
		}
		return "", ctx.Err()
	case err := <-done:
		// Command completed.
		if err != nil {
			log.Error().
				Err(err).
				Str("command", command).
				Str("stderr", stderrBuf.String()).
				Msg("Command execution failed")
			return stdoutBuf.String() + stderrBuf.String(), fmt.Errorf("command execution failed: %w, stderr: %s", err, stderrBuf.String())
		}
		log.Info().
			Str("command", command).
			Msg("Command executed successfully")
		return stdoutBuf.String(), nil
	}
}

// netDialer is a custom dialer that respects the context for SSH connections.
type netDialer struct {
	ctx     context.Context
	timeout time.Duration
}

// Dial establishes a network connection respecting the context's cancellation.
// It uses the provided timeout for the connection attempt.
func (d *netDialer) Dial(network, addr string) (net.Conn, error) {
	dialer := net.Dialer{
		Timeout: d.timeout,
	}
	return dialer.DialContext(d.ctx, network, addr)
}
