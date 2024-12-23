package dockercli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"kasmlink/pkg/dockercli"
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

func TestListImages(t *testing.T) {
	t.Run("Local List Images", func(t *testing.T) {
		ctx := context.Background()
		mockExecutor := new(MockExecutor)

		client := dockercli.NewDockerClient(mockExecutor)

		// Simulate output for the `docker images` command
		mockOutput := `{"repository":"test-repo","tag":"latest","id":"123","size":"10MB"}
{"repository":"another-repo","tag":"v1.0","id":"456","size":"20MB"}
`
		mockExecutor.On("Execute", ctx, "docker", mock.Anything).Return([]byte(mockOutput), nil)

		images, err := client.ListImages(ctx, dockercli.ListImagesOptions{SSH: nil})

		assert.NoError(t, err)
		assert.Len(t, images, 2)
		assert.Equal(t, "test-repo", images[0].Repository)
		assert.Equal(t, "latest", images[0].Tag)
		mockExecutor.AssertExpectations(t)
	})
}
