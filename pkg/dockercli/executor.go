package dockercli

import (
	"bytes"
	"context"
	"fmt"
	"kasmlink/pkg/shadowssh"
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

// SSHCommandExecutor executes shell commands over SSH.
type SSHCommandExecutor struct {
	sshConfig *shadowssh.Config
}

// NewSSHCommandExecutor creates a new SSHCommandExecutor.
func NewSSHCommandExecutor(sshConfig *shadowssh.Config) *SSHCommandExecutor {
	if sshConfig == nil {
		panic("SSHConfig must not be nil")
	}
	return &SSHCommandExecutor{sshConfig: sshConfig}
}

// Execute runs the specified command with arguments over SSH.
func (e *SSHCommandExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	// Combine command and arguments into a single command string.
	fullCommand := command
	if len(args) > 0 {
		fullCommand = fmt.Sprintf("%s %s", command, joinArgs(args))
	}

	// Establish SSH connection.
	client, err := shadowssh.NewClient(ctx, e.sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer client.Close()

	// Execute command over SSH.
	output, err := client.ExecuteCommand(ctx, fullCommand)
	if err != nil {
		return nil, fmt.Errorf("SSH command execution failed: %w", err)
	}

	log.Debug().Str("command", fullCommand).Str("output", output).Msg("SSH command executed successfully")
	return []byte(output), nil
}

// Helper to join arguments into a space-separated string.
func joinArgs(args []string) string {
	// Convert []string to []interface{}
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg
	}

	// Use fmt.Sprint with the converted slice
	return fmt.Sprint(interfaceArgs...)
}
