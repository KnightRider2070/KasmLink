package tests

import (
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndParseComposeFile(t *testing.T) {
	t.Run("Positive Case - Valid YAML File", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()
		validFilePath := filepath.Join(tempDir, "valid-compose.yml")

		// Write a valid Docker Compose YAML to the file
		validYAML := `
version: "3.9"
services:
  web:
    image: nginx
`
		err := os.WriteFile(validFilePath, []byte(validYAML), 0644)
		assert.NoError(t, err, "Failed to write valid test YAML file")

		// Call the function
		result, err := dockercompose.LoadAndParseComposeFile(validFilePath)

		// Assert no errors occurred
		assert.NoError(t, err)
		assert.Equal(t, "3.9", result.Version)
		assert.Contains(t, result.Services, "web")
		assert.Equal(t, "nginx", result.Services["web"].Image)
	})

	t.Run("Negative Case - Missing File", func(t *testing.T) {
		// Use a guaranteed non-existent file path
		missingFilePath := filepath.Join("nonexistent-dir", "missing-compose.yml")

		// Call the function
		_, err := dockercompose.LoadAndParseComposeFile(missingFilePath)

		// Assert an error occurred
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file does not exist")
	})

	t.Run("Negative Case - Invalid YAML File", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()
		invalidFilePath := filepath.Join(tempDir, "invalid-compose.yml")

		// Write an invalid YAML content to the file
		invalidYAML := `
version: "3.9"
services
  web:
    image: nginx
`
		err := os.WriteFile(invalidFilePath, []byte(invalidYAML), 0644)
		assert.NoError(t, err, "Failed to write invalid test YAML file")

		// Call the function
		_, err = dockercompose.LoadAndParseComposeFile(invalidFilePath)

		// Assert an error occurred
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode YAML config file")
	})

	t.Run("Negative Case - Validation Error", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()
		invalidFilePath := filepath.Join(tempDir, "invalid-structure.yml")

		// Write YAML with an invalid structure to the file
		invalidYAML := `
version: "3.9"
services:
  web:
    image: nginx
    invalid_field: "unexpected"
`
		err := os.WriteFile(invalidFilePath, []byte(invalidYAML), 0644)
		assert.NoError(t, err, "Failed to write invalid structure test YAML file")

		// Call the function
		_, err = dockercompose.LoadAndParseComposeFile(invalidFilePath)

		// Assert a validation error occurred
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "invalid_field is not allowed")
	})
}
