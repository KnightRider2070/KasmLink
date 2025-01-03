package dockercli_test

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
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

		// Adjust command and arguments based on the OS
		var command string
		var args []string
		if runtime.GOOS == "windows" {
			command = "cmd"
			args = []string{"/C", "echo hello, world"}
		} else {
			command = "echo"
			args = []string{"hello, world"}
		}

		output, err := executor.Execute(ctx, command, args...)

		expectedOutput := "hello, world\n"
		if runtime.GOOS == "windows" {
			expectedOutput = "hello, world\r\n" // Windows adds \r\n
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("Command Execution Failure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		output, err := executor.Execute(ctx, "nonexistent_command")

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "command execution failed")
	})

	t.Run("Context Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		var command string
		var args []string
		if runtime.GOOS == "windows" {
			command = "timeout"
			args = []string{"2"}
		} else {
			command = "sleep"
			args = []string{"2"}
		}

		output, err := executor.Execute(ctx, command, args...)

		assert.Error(t, err)
		assert.Nil(t, output)
		var ctxErr *exec.ExitError
		assert.True(t, errors.As(err, &ctxErr) || ctx.Err() == context.DeadlineExceeded)
	})
}
