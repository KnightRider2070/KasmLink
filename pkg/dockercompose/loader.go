package dockercompose

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// LoadComposeFile loads the configuration for services, volumes, networks, etc., from a YAML file into the ComposeFile struct.
func LoadComposeFile(configPath string) (*ComposeFile, error) {
	// Open the YAML configuration file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %v", configPath, err)
	}
	defer file.Close()

	// Create a ComposeFile instance to hold the data
	var composeFile ComposeFile

	// Decode the YAML file into the ComposeFile struct
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&composeFile); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	// Return the loaded configuration
	return &composeFile, nil
}
