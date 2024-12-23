package dockercli

import (
	"context"
	"golang.org/x/crypto/ssh"
	"time"
)

// DockerClient encapsulates Docker-specific operations.
type DockerClient struct {
	executor          CommandExecutor
	sshClientFactory  func(ctx context.Context, opts *SSHOptions) (*ssh.Client, error)
	retries           int
	initialRetryDelay time.Duration
	backoffMultiplier int
	maxRetryDelay     time.Duration
	jitterFactor      float64
}

// NewDockerClient creates a new DockerClient with a given executor.
func NewDockerClient(executor CommandExecutor) *DockerClient {
	return &DockerClient{
		executor: executor,
		sshClientFactory: func(ctx context.Context, opts *SSHOptions) (*ssh.Client, error) {
			return newSSHClient(opts)
		},
		retries:           3,
		initialRetryDelay: 2 * time.Second,
		backoffMultiplier: 2,
		maxRetryDelay:     16 * time.Second,
		jitterFactor:      0.1,
	}
}
