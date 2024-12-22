// dockercli/dockercli.go
package dockercli

import (
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/fatih/color"
)

// DockerClient encapsulates the Docker client and retry configurations.
type DockerClient struct {
	cli               *client.Client
	retries           int
	initialRetryDelay time.Duration
	backoffMultiplier int
	maxRetryDelay     time.Duration
	jitterFactor      float64
	successColor      *color.Color
	errorColor        *color.Color

	// Mutex to protect any future mutable state
	mu sync.RWMutex
}

// NewDockerClient initializes and returns a new DockerClient.
// It sets default values if provided configurations are zero-valued.
// Parameters:
// - cli: The Docker client instance.
// - retries: Number of retry attempts for operations.
// - initialRetryDelay: Initial delay before retrying an operation.
// - backoffMultiplier: Multiplier for exponential backoff.
// - maxRetryDelay: Maximum delay between retries.
// - jitterFactor: Factor for adding jitter to retry delays.
func NewDockerClient(cli *client.Client, retries int, initialRetryDelay time.Duration, backoffMultiplier int, maxRetryDelay time.Duration, jitterFactor float64) *DockerClient {
	if retries <= 0 {
		retries = 3
	}
	if initialRetryDelay <= 0 {
		initialRetryDelay = 2 * time.Second
	}
	if backoffMultiplier <= 0 {
		backoffMultiplier = 2
	}
	if maxRetryDelay <= 0 {
		maxRetryDelay = 16 * time.Second
	}
	if jitterFactor <= 0 {
		jitterFactor = 0.1 // 10% jitter
	}

	return &DockerClient{
		cli:               cli,
		retries:           retries,
		initialRetryDelay: initialRetryDelay,
		backoffMultiplier: backoffMultiplier,
		maxRetryDelay:     maxRetryDelay,
		jitterFactor:      jitterFactor,
		successColor:      color.New(color.FgGreen),
		errorColor:        color.New(color.FgRed),
	}
}
