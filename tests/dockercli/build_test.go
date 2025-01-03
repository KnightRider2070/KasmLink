package dockercli_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"kasmlink/pkg/dockercli"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to set up a temporary test directory
func setupTestContextDir(t *testing.T) string {
	tempDir := t.TempDir()

	// Create a sample Dockerfile
	err := os.WriteFile(filepath.Join(tempDir, "Dockerfile"), []byte("FROM alpine:latest"), 0644)
	assert.NoError(t, err)

	// Create a sample file in the context directory
	err = os.WriteFile(filepath.Join(tempDir, "app.txt"), []byte("test content"), 0644)
	assert.NoError(t, err)

	return tempDir
}

func TestBuildImage(t *testing.T) {
	t.Run("Local Build", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := new(MockExecutor)
		mockFS := dockercli.NewLocalFileSystem()

		// Set up the test context directory
		testContextDir := setupTestContextDir(t)

		// Simulate local build command execution
		mockExecutor.On("Execute", ctx, "docker", mock.Anything).
			Return([]byte(`{"stream":"Step 1/3: FROM alpine:latest\n"}`), nil)

		client := dockercli.NewDockerClient(mockExecutor, mockFS)

		options := dockercli.BuildImageOptions{
			ContextDir:     testContextDir,
			DockerfilePath: filepath.Join(testContextDir, "Dockerfile"),
			ImageTag:       "test-image",
			BuildArgs:      map[string]string{"ARG1": "VALUE1"},
			SSH:            nil, // Local build
		}

		// Ensure the BuildImage method runs without error.
		err := dockercli.BuildImage(ctx, client, options)
		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})
}
