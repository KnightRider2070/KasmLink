package dockercli_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"kasmlink/pkg/dockercli"
	"os"
	"path/filepath"
	"testing"
)

// MockExecutor is a mock implementation of the CommandExecutor interface.
type MockExecutor struct {
	mock.Mock
}

// Execute mocks the Execute method of the CommandExecutor interface.
func (m *MockExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	argsList := m.Called(ctx, command, args)
	return argsList.Get(0).([]byte), argsList.Error(1)
}

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

// Helper function to create a valid temporary SSH private key
func setupTestPrivateKey(t *testing.T) string {
	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Convert the private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create a temporary file to store the key
	tempFile, err := os.CreateTemp("", "test_private_key")
	assert.NoError(t, err)
	defer tempFile.Close()

	// Write the private key to the file
	_, err = tempFile.Write(privateKeyPEM)
	assert.NoError(t, err)

	return tempFile.Name()
}

func TestBuildImage(t *testing.T) {
	t.Run("Local Build", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := new(MockExecutor)

		// Set up the test context directory
		testContextDir := setupTestContextDir(t)

		// Simulate local build command execution with JSON-formatted logs.
		mockExecutor.On("Execute", ctx, "docker", mock.Anything).
			Return([]byte(`{"stream":"Step 1/3: FROM alpine:latest\n"}
{"stream":"Step 2/3: COPY . /app\n"}
{"stream":"Step 3/3: CMD [\"/bin/sh\"]\n"}`), nil)

		client := dockercli.NewDockerClient(mockExecutor)

		options := dockercli.BuildImageOptions{
			ContextDir:     testContextDir,
			DockerfilePath: filepath.Join(testContextDir, "Dockerfile"),
			ImageTag:       "test-image",
			BuildArgs:      map[string]string{"ARG1": "VALUE1"},
			SSH:            nil, // Local build
		}

		// Ensure the BuildImage method runs without error.
		err := client.BuildImage(ctx, options)
		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})
}
