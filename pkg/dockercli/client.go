package dockercli

import (
	"context"
	"fmt"
	"kasmlink/pkg/shadowssh"
	"time"
)

// DockerClient encapsulates Docker-specific operations.
type DockerClient struct {
	executor          CommandExecutor
	fs                FileSystem
	sshClientFactory  func(ctx context.Context, opts *shadowssh.Config) (*shadowssh.Client, error)
	retries           int
	initialRetryDelay time.Duration
	backoffMultiplier int
	maxRetryDelay     time.Duration
	jitterFactor      float64
}

// NewDockerClient creates a new DockerClient with a given executor and file system.
func NewDockerClient(executor CommandExecutor, fs FileSystem) *DockerClient {
	return &DockerClient{
		executor: executor,
		fs:       fs,
		sshClientFactory: func(ctx context.Context, opts *shadowssh.Config) (*shadowssh.Client, error) {
			config, err := shadowssh.NewConfig(opts.Username, opts.Password, opts.Host, opts.Port, opts.KnownHostsPath, opts.ConnectionTimeout)
			if err != nil {
				return nil, fmt.Errorf("failed to create SSH config: %w", err)
			}
			return shadowssh.NewClient(ctx, config)
		},
		retries:           3,
		initialRetryDelay: 2 * time.Second,
		backoffMultiplier: 2,
		maxRetryDelay:     16 * time.Second,
		jitterFactor:      0.1,
	}
}
