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

// Config holds the configuration for SSH connections.
type Config struct {
	Username          string
	Password          string
	Host              string
	Port              int
	KnownHostsPath    string
	ConnectionTimeout time.Duration
}

// Client manages an SSH connection.
type Client struct {
	client *ssh.Client
	config Config
}

func (c *Client) Client() *ssh.Client {
	return c.client
}

func (c *Client) Config() Config {
	return c.config
}

// NewConfig initializes and validates an SSH configuration.
func NewConfig(username, password, host string, port int, knownHostsPath string, timeout time.Duration) (*Config, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}

	// Expand tilde (~) to home directory if present.
	if len(knownHostsPath) >= 2 && knownHostsPath[:2] == "~/" {
		userInfo, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
		knownHostsPath = filepath.Join(userInfo.HomeDir, knownHostsPath[2:])
	}

	return &Config{
		Username:          username,
		Password:          password,
		Host:              host,
		Port:              port,
		KnownHostsPath:    knownHostsPath,
		ConnectionTimeout: timeout,
	}, nil
}

// NewClient establishes an SSH connection and returns a Client.
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("SSH configuration cannot be nil")
	}

	hostKeyCallback, err := knownhosts.New(config.KnownHostsPath)
	if err != nil {
		log.Error().Err(err).Str("known_hosts_path", config.KnownHostsPath).Msg("Failed to load known hosts")
		return nil, fmt.Errorf("failed to load known hosts: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(config.Password)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         config.ConnectionTimeout,
	}

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := dialWithContext(ctx, "tcp", address, config.ConnectionTimeout)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Failed to connect to SSH server")
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Failed to establish SSH connection")
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	log.Info().Str("address", address).Msg("SSH connection established")
	return &Client{
		client: ssh.NewClient(clientConn, chans, reqs),
		config: *config,
	}, nil
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	if err := c.client.Close(); err != nil {
		log.Error().Err(err).Str("host", c.config.Host).Msg("Failed to close SSH connection")
		return fmt.Errorf("failed to close SSH connection: %w", err)
	}
	log.Info().Str("host", c.config.Host).Msg("SSH connection closed")
	return nil
}

// ExecuteCommand runs a command on the remote server and returns its output.
func (c *Client) ExecuteCommand(ctx context.Context, command string) (string, error) {
	session, err := c.createSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	output := stdout.String() + stderr.String()

	if err != nil {
		log.Error().Err(err).Str("command", command).Msg("Command execution failed")
		return output, fmt.Errorf("command execution failed: %w", err)
	}

	log.Info().Str("command", command).Msg("Command executed successfully")
	return output, nil
}

// ExecuteCommandWithLogs runs a command and logs its output in real-time.
func (c *Client) ExecuteCommandWithLogs(ctx context.Context, command string, logDuration time.Duration) (string, error) {
	session, err := c.createSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	output, err := c.streamCommandOutput(ctx, stdoutPipe, stderrPipe, logDuration)
	if err != nil {
		return output, err
	}

	if err := session.Wait(); err != nil {
		return output, fmt.Errorf("command execution failed: %w", err)
	}
	return output, nil
}

// Helper methods

func (c *Client) createSession() (*ssh.Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create SSH session")
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	return session, nil
}

func (c *Client) streamCommandOutput(ctx context.Context, stdoutPipe, stderrPipe io.Reader, logDuration time.Duration) (string, error) {
	var output string
	logTimer := time.NewTimer(logDuration)
	defer logTimer.Stop()

	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdoutPipe, stderrPipe))
		for scanner.Scan() {
			line := scanner.Text()
			log.Info().Str("output", line).Msg("Command output")
			output += line + "\n"
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
			errChan <- err
		}
		close(outputChan)
		close(errChan)
	}()

	select {
	case <-logTimer.C:
		log.Warn().Msg("Logging duration expired; returning output")
		return output, nil
	case <-ctx.Done():
		log.Warn().Err(ctx.Err()).Msg("Context canceled during command execution")
		return output, ctx.Err()
	case err := <-errChan:
		if err != nil {
			return output, fmt.Errorf("error reading command output: %w", err)
		}
	}
	return output, nil
}

// Helper function for dialing with context
func dialWithContext(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{Timeout: timeout}
	return dialer.DialContext(ctx, network, address)
}
