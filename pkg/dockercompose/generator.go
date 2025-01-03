package dockercompose

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GenerateDockerComposeYAML generates a Docker Compose YAML file from the DockerCompose struct.
func GenerateDockerComposeYAML(composeFile DockerCompose, outputPath string) error {
	// Validate the input structure
	if err := ValidateDockerCompose(composeFile); err != nil {
		return fmt.Errorf("invalid DockerCompose structure: %w", err)
	}

	// Ensure the output directory is valid and writable
	outputDir := filepath.Dir(outputPath)
	if err := validateDirectory(outputDir); err != nil {
		return fmt.Errorf("invalid output directory %s: %w", outputDir, err)
	}

	// Convert the DockerCompose struct to YAML
	yamlData, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal DockerCompose to YAML: %w", err)
	}

	// Write the YAML data to the output file atomically
	if err := atomicWriteFile(outputPath, string(yamlData)); err != nil {
		return fmt.Errorf("failed to write Docker Compose YAML file: %w", err)
	}

	return nil
}

// validateDirectory checks if the directory exists and is writable.
func validateDirectory(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// Attempt to create the directory if it doesn't exist
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to access directory: %w", err)
		}
	}

	// Test if the directory is writable by creating a temp file
	tempFile, err := os.CreateTemp(dir, "test")
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())

	return nil
}

// atomicWriteFile writes content to a temporary file and then renames it to the target file.
func atomicWriteFile(filePath, content string) error {
	tmpFile := filePath + ".tmp"

	// Write to the temporary file
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tmpFile, err)
	}

	// Rename the temporary file to the target file
	if err := os.Rename(tmpFile, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary file %s to %s: %w", tmpFile, filePath, err)
	}

	return nil
}
