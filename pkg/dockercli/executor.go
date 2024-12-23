package dockercli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// CommandExecutor defines the interface for executing shell commands.
type CommandExecutor interface {
	Execute(ctx context.Context, command string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor is a default implementation of CommandExecutor.
type DefaultCommandExecutor struct{}

// Execute runs the specified command with arguments.
func (e *DefaultCommandExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	log.Debug().Str("command", command).Strs("args", args).Msg("Executing command")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Command execution failed")
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	log.Debug().Str("output", string(output)).Msg("Command executed successfully")

	// Example: Process logs if the command is `docker build`
	if command == "docker" && len(args) > 0 && args[0] == "build" {
		logReader := bytes.NewReader(output)
		if err := PrintBuildLogs(logReader); err != nil {
			return nil, fmt.Errorf("failed to process build logs: %w", err)
		}
	}

	return output, nil
}
