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

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Convert the DockerCompose struct to YAML
	yamlData, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal DockerCompose to YAML: %w", err)
	}

	// Write the YAML data to the output file
	if err := writeFile(outputPath, string(yamlData)); err != nil {
		return err
	}

	return nil
}

// writeFile writes the given content to a file at the specified path.
func writeFile(filePath, content string) error {
	// Create the file
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", filePath, err)
	}
	defer outputFile.Close()

	// Write the content to the file
	if _, err := outputFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to output file %s: %w", filePath, err)
	}

	return nil
}
