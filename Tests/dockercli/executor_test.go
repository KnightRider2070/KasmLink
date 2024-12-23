package dockercli

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/dockercli"
)

func TestDefaultCommandExecutor_Execute(t *testing.T) {
	executor := dockercli.DefaultCommandExecutor{}

	t.Run("Successful Command Execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Use a simple command that should succeed on most systems
		output, err := executor.Execute(ctx, "echo", "hello, world")

		assert.NoError(t, err)
		assert.Equal(t, "hello, world\n", string(output)) // The newline is expected
	})

	t.Run("Command Execution Failure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Use an invalid command to simulate failure
		output, err := executor.Execute(ctx, "nonexistent_command")

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "command execution failed")
	})

	t.Run("Context Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Use a command that will hang if allowed (simulate timeout)
		output, err := executor.Execute(ctx, "sleep", "2")

		assert.Error(t, err)
		assert.Nil(t, output)
		var ctxErr *exec.ExitError
		assert.True(t, errors.As(err, &ctxErr) || ctx.Err() == context.DeadlineExceeded)
	})
}
