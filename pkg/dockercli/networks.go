package dockercli

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
}

// CreateDockerNetwork creates a new Docker network with the provided options.
func CreateDockerNetwork(opts NetworkOptions) error {
	// Validate driver
	if !opts.Driver.IsValid() {
		return errors.New("invalid driver: must be one of 'bridge', 'overlay', or 'none'")
	}

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

	// Run the command
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create network: %v - %s", err, string(output))
	}

	fmt.Printf("Network %s created successfully.\n", opts.Name)
	return nil
}

// InspectNetwork inspects a Docker network by name or ID.
func InspectNetwork(networkName string) (string, error) {
	cmd := exec.Command("docker", "network", "inspect", networkName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to inspect network: %v - %s", err, string(output))
	}
	return string(output), nil
}

// RemoveNetwork removes a Docker network by name or ID.
func RemoveNetwork(networkName string) error {
	cmd := exec.Command("docker", "network", "rm", networkName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove network: %v - %s", err, string(output))
	}
	fmt.Printf("Network %s removed successfully.\n", networkName)
	return nil
}

// ListComposeNetworks lists all networks associated with the Docker Compose project.
func ListComposeNetworks(composeFilePath string) (map[string]string, error) {
	cmd := exec.Command("docker-compose", "-f", composeFilePath, "config", "--services")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %v - %s", err, string(output))
	}

	// Parse the output into a map of network names
	lines := strings.Split(string(output), "\n")
	networks := make(map[string]string)
	for _, line := range lines {
		if line != "" {
			// Fetch the network details using Docker
			networkName := line
			netID, err := getNetworkID(networkName)
			if err == nil {
				networks[networkName] = netID
			}
		}
	}

	return networks, nil
}

// getNetworkID fetches the network ID based on the network name.
func getNetworkID(networkName string) (string, error) {
	cmd := exec.Command("docker", "network", "ls", "--filter", fmt.Sprintf("name=%s", networkName), "--format", "{{.ID}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get network ID: %v - %s", err, string(output))
	}

	networkID := strings.TrimSpace(string(output))
	if networkID == "" {
		return "", fmt.Errorf("no network found with the name %s", networkName)
	}

	return networkID, nil
}
