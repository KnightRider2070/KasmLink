package ssh

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient encapsulates the SSH configuration and session management
type SSHClient struct {
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
	Client     *ssh.Client
}

// NewSSHClient creates a new instance of SSHClient
func NewSSHClient(host string, port int, username, password, privateKey string) (*SSHClient, error) {
	client := &SSHClient{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		PrivateKey: privateKey,
	}

	// Try to establish a connection immediately.
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return client, nil
}

// Connect establishes an SSH connection using the provided credentials
func (c *SSHClient) Connect() error {
	var authMethods []ssh.AuthMethod

	// Using private key if provided
	if c.PrivateKey != "" {
		key, err := os.ReadFile(c.PrivateKey) // Updated from ioutil.ReadFile to os.ReadFile
		if err != nil {
			return fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("unable to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Using password if provided
	if c.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.Password))
	}

	// Configuring SSH client configuration
	config := &ssh.ClientConfig{
		User:            c.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // NOTE: Should be replaced with a proper callback for production.
		Timeout:         10 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", c.Host, c.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	c.Client = client
	return nil
}

// RunCommand executes a command on the remote SSH server
func (c *SSHClient) RunCommand(cmd string) (string, error) {
	if c.Client == nil {
		return "", fmt.Errorf("SSH client is not connected")
	}

	session, err := c.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Capture the output of the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// Disconnect closes the SSH connection
func (c *SSHClient) Disconnect() error {
	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			return fmt.Errorf("failed to close SSH connection: %v", err)
		}
		c.Client = nil
	}
	return nil
}
