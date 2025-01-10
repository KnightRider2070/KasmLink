package dockercompose_tests

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGenerateDockerComposeYAML(t *testing.T) {
	// Define a valid DockerCompose instance
	validComposeFile := dockercompose.DockerCompose{
		Version: "3.9",
		Services: map[string]dockercompose.ServiceDefinition{
			"web": {
				Image: "nginx",
				Ports: []dockercompose.PortMapping{
					{Target: 80, Published: 8080, Protocol: "tcp", Mode: "host"},
				},
			},
		},
	}

	t.Run("Positive Case - Valid Input", func(t *testing.T) {
		// Create a temporary directory for the output
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "docker-compose.yml")

		// Call the function
		err := dockercompose.GenerateDockerComposeYAML(validComposeFile, outputPath)

		// Assert no errors occurred
		assert.NoError(t, err)

		// Assert the output file exists
		_, err = os.Stat(outputPath)
		assert.NoError(t, err)

		// Read and verify the content
		data, err := ioutil.ReadFile(outputPath)
		assert.NoError(t, err)
		content := string(data)
		t.Logf("Generated YAML:\n%s", content)

		// Validate key content
		assert.Contains(t, content, "version: \"3.9\"")
		assert.Contains(t, content, "web:")
		assert.Contains(t, content, "image: nginx")
	})

	t.Run("Negative Case - Invalid Directory", func(t *testing.T) {
		// Use an invalid directory path that fails on both Windows and Linux
		var invalidDir string
		if runtime.GOOS == "windows" {
			invalidDir = filepath.Join("P:\\NonExistentDrive\\InvalidDirectory")
		} else {
			invalidDir = filepath.Join("/nonexistent", "invalid_directory")
		}
		outputPath := filepath.Join(invalidDir, "docker-compose.yml")

		// Call the function
		err := dockercompose.GenerateDockerComposeYAML(validComposeFile, outputPath)

		// Assert that an error occurred
		assert.Error(t, err, "Expected an error for an invalid directory")
		t.Logf("Error: %v", err)

		// Check if the error contains expected failure information
		assert.Contains(t, err.Error(), "directory does not exist", "Error message did not match expected substring")
	})

	t.Run("Negative Case - Invalid Struct Data", func(t *testing.T) {
		// Define an invalid DockerCompose struct (e.g., invalid protocol or mode)
		invalidComposeFile := dockercompose.DockerCompose{
			Version: "3.9",
			Services: map[string]dockercompose.ServiceDefinition{
				"web": {
					Image: "nginx",
					Ports: []dockercompose.PortMapping{
						{Target: 80, Published: 8080, Protocol: "unsupported", Mode: "invalid-mode"}, // Invalid values
					},
					PullPolicy: "invalid-policy", // Invalid value
				},
			},
		}

		// Create a temporary directory for the output
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "docker-compose.yml")

		// Call the function
		err := dockercompose.GenerateDockerComposeYAML(invalidComposeFile, outputPath)

		// Assert an error occurred
		assert.Error(t, err)
		t.Logf("Expected error: %v", err)

		// Check if the error contains validation failure information
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "must be one of the following")
		assert.Contains(t, err.Error(), "pull_policy")
	})
}
