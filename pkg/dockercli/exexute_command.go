package dockercli

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"time"
)

// Strategy for executing Docker commands:
// This function executes Docker commands with a retry mechanism and enforced timeout to ensure resilience.
// - Configurable Retries: Allows specifying the number of retries for handling transient issues.
// - Timeout Enforcement: Uses a context with a deadline to prevent commands from hanging indefinitely.
// - Detailed Logging: Logs details for each attempt, including command arguments, errors, and durations, to facilitate troubleshooting.
// - Exponential Backoff: Waits before retrying, with an increasing delay on each retry to handle temporary disruptions effectively.
// - Graceful Error Handling: If all attempts fail, a descriptive error is returned summarizing the command and error details.

const (
	initialRetryDelay = 2 * time.Second
	backoffMultiplier = 2
	maxRetryDelay     = 16 * time.Second
)

// Helper function to execute Docker Compose commands with retry and timeout
func executeDockerCommand(ctx context.Context, retries int, command string, args ...string) ([]byte, error) {
	var output []byte
	var err error
	retryDelay := initialRetryDelay

	for attempt := 1; attempt <= retries; attempt++ {
		select {
		case <-ctx.Done():
			// If the context has been cancelled, log and return the context error.
			log.Error().
				Int("attempt", attempt).
				Str("command", command).
				Strs("args", args).
				Msg("Command execution aborted due to context cancellation")
			return nil, fmt.Errorf("command execution aborted due to context cancellation: %w", ctx.Err())
		default:
			// Proceed with executing the command
			cmd := exec.CommandContext(ctx, command, args...)
			startTime := time.Now()

			log.Debug().
				Str("command", command).
				Strs("args", args).
				Int("attempt", attempt).
				Msg("Executing Docker command")

			output, err = cmd.CombinedOutput()
			duration := time.Since(startTime)

			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Error().
					Dur("duration", duration).
					Int("attempt", attempt).
					Msgf("Command timed out after %v: %s %v", duration, command, args)
			} else if err != nil {
				log.Error().
					Err(err).
					Str("output", string(output)).
					Int("attempt", attempt).
					Dur("duration", duration).
					Msgf("Failed to execute command: %s %v", command, args)
			} else {
				log.Info().
					Str("command", command).
					Strs("args", args).
					Dur("duration", duration).
					Int("attempt", attempt).
					Str("output", string(output)).
					Msg("Docker command executed successfully")
				return output, nil
			}

			// If not successful, wait before retrying, with exponential backoff
			log.Warn().
				Int("attempt", attempt).
				Dur("retry_delay", retryDelay).
				Msg("Retrying command after delay")

			time.Sleep(retryDelay)
			retryDelay *= time.Duration(backoffMultiplier)
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
		}
	}

	return nil, fmt.Errorf("command failed after %d attempts: %s %v, error: %w", retries, command, args, err)
}
