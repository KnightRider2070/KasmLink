package dockercli

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

// NetworkDriver defines a custom type for network drivers with only allowed values.
type NetworkDriver string

// Define constants for allowed network drivers.
const (
	DriverBridge  NetworkDriver = "bridge"
	DriverOverlay NetworkDriver = "overlay"
	DriverNone    NetworkDriver = "none"
)

// IsValid checks if the provided driver is a valid value.
func (d NetworkDriver) IsValid() bool {
	switch d {
	case DriverBridge, DriverOverlay, DriverNone:
		return true
	default:
		return false
	}
}

// NetworkOptions represents the customizable options for creating a Docker network.
type NetworkOptions struct {
	Name        string            // Name of the network
	Driver      NetworkDriver     // Driver type: "bridge", "overlay", or custom
	Attachable  bool              // Whether to allow manual container attachment (for overlay networks)
	Internal    bool              // Whether to restrict external access
	IPv6        bool              // Enable or disable IPv6
	Subnet      string            // Subnet in CIDR format
	Gateway     string            // Gateway for the network
	IPRange     string            // Allocate container IP from a sub-range
	IPAMDriver  string            // IP Address Management driver
	IPAMOptions map[string]string // IPAM driver specific options
	Labels      map[string]string // Labels for the network
	Options     map[string]string // Driver specific options
	Scope       string            // Scope of the network (local, swarm)
	RetryCount  int               // Retry attempts for Docker commands
	RetryDelay  time.Duration     // Delay between retry attempts
	Timeout     time.Duration     // Timeout for Docker commands
}

// default values for retrying Docker commands
const (
	defaultRetryCount     = 3
	defaultRetryDelay     = 2 * time.Second
	defaultCommandTimeout = 30 * time.Second
)

// CreateDockerNetwork creates a new Docker network with the provided options, with retry and timeout support.
func CreateDockerNetwork(opts NetworkOptions) error {
	// Validate driver
	if !opts.Driver.IsValid() {
		log.Error().Str("driver", string(opts.Driver)).Msg("Invalid driver specified")
		return errors.New("invalid driver: must be one of 'bridge', 'overlay', or 'none'")
	}

	// Set default retry count, delay, and timeout if not provided
	if opts.RetryCount <= 0 {
		opts.RetryCount = defaultRetryCount
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = defaultRetryDelay
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultCommandTimeout
	}

	log.Info().Str("network_name", opts.Name).Msg("Creating Docker network")

	// Construct the base command
	args := []string{"network", "create"}

	// Add driver option if provided
	if opts.Driver != "" {
		args = append(args, "--driver", string(opts.Driver))
	}

	// Add other options...
	if opts.Attachable {
		args = append(args, "--attachable")
	}
	if opts.Internal {
		args = append(args, "--internal")
	}
	if opts.IPv6 {
		args = append(args, "--ipv6")
	}
	if opts.Subnet != "" {
		args = append(args, "--subnet", opts.Subnet)
	}
	if opts.Gateway != "" {
		args = append(args, "--gateway", opts.Gateway)
	}
	if opts.IPRange != "" {
		args = append(args, "--ip-range", opts.IPRange)
	}
	if opts.IPAMDriver != "" {
		args = append(args, "--ipam-driver", opts.IPAMDriver)
	}
	if opts.IPAMOptions != nil {
		for key, value := range opts.IPAMOptions {
			args = append(args, "--ipam-opt", fmt.Sprintf("%s=%s", key, value))
		}
	}
	if opts.Labels != nil {
		for key, value := range opts.Labels {
			args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
		}
	}
	if opts.Options != nil {
		for key, value := range opts.Options {
			args = append(args, "--opt", fmt.Sprintf("%s=%s", key, value))
		}
	}
	if opts.Scope != "" {
		args = append(args, "--scope", opts.Scope)
	}

	// Append the network name
	args = append(args, opts.Name)

	// Execute the command with retry mechanism
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, opts.RetryCount, "docker", args...)
	if err != nil {
		return fmt.Errorf("failed to create network after %d attempts: %w", opts.RetryCount, err)
	}

	log.Info().Str("network_name", opts.Name).Str("output", string(output)).Msg("Docker network created successfully")
	return nil
}

// InspectNetwork inspects a Docker network by name or ID, with retry and timeout support.
func InspectNetwork(networkName string, retryCount int, retryDelay time.Duration, timeout time.Duration) (string, error) {
	if retryCount <= 0 {
		retryCount = defaultRetryCount
	}
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}

	log.Info().Str("network_name", networkName).Msg("Inspecting Docker network")

	// Execute the command with retry mechanism
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := executeDockerCommand(ctx, retryCount, "docker", "network", "inspect", networkName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect network after %d attempts: %w", retryCount, err)
	}

	log.Info().Str("network_name", networkName).Str("output", string(output)).Msg("Docker network inspected successfully")
	return string(output), nil
}

// RemoveNetwork removes a Docker network by name or ID, with retry and timeout support.
func RemoveNetwork(networkName string, retryCount int, retryDelay time.Duration, timeout time.Duration) error {
	if retryCount <= 0 {
		retryCount = defaultRetryCount
	}
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}

	log.Info().Str("network_name", networkName).Msg("Removing Docker network")

	// Execute the command with retry mechanism
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := executeDockerCommand(ctx, retryCount, "docker", "network", "rm", networkName)
	if err != nil {
		return fmt.Errorf("failed to remove network after %d attempts: %w", retryCount, err)
	}

	log.Info().Str("network_name", networkName).Msg("Docker network removed successfully")
	return nil
}
