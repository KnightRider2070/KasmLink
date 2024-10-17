package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// UploadFile uploads a local file to the remote server using SCP
func (c *SSHClient) UploadFile(localPath, remotePath string) error {
	if c.Client == nil {
		return fmt.Errorf("SSH client is not connected")
	}

	session, err := c.Client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Open the local file for reading
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	// Get file size and base filename
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}
	filename := filepath.Base(localPath)
	filesize := stat.Size()

	// Start SCP process on the remote server
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", filesize, filename) // Send file metadata
		io.Copy(w, file)                                    // Send the file data
		fmt.Fprint(w, "\x00")                               // Send transfer end signal
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("failed to run SCP command on remote server: %v", err)
	}

	return nil
}

// DownloadFile downloads a file from the remote server to the local machine using SCP
func (c *SSHClient) DownloadFile(remotePath, localPath string) error {
	if c.Client == nil {
		return fmt.Errorf("SSH client is not connected")
	}

	session, err := c.Client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Create the local file for writing
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer file.Close()

	// Start SCP process on the remote server
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprint(w, "\x00") // Send initial OK signal
	}()

	// Open the remote file and copy its contents to the local file
	r, _ := session.StdoutPipe()
	err = session.Start(fmt.Sprintf("scp -f %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to run SCP command on remote server: %v", err)
	}

	// Read the SCP protocol response and write to the local file
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read from remote file: %v", err)
		}
		if n == 0 {
			break
		}
		if _, err := file.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write to local file: %v", err)
		}
	}

	return nil
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
