package dockerRegistry

import sshmanager "kasmlink/pkg/sshmanager"

type RegistryConfig struct {
	containerName       string
	registryImageToPull string
	url                 string
	port                int
	user                string
	password            string
}

// TODO: Implement setupRegistry
func setupRegistry(targetSSH *sshmanager.SSHConfig, registryConfig *RegistryConfig) error {

	// 1. Check if image is present
	// 2. If not present on Node check if present locally in .tar file
	// 2.1 If present locally load image to node from .tar file
	// 3. If not present locally pull from docker hub to node
	// 4. Start registry container with arguments
	return nil
}
