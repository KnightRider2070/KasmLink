package dockercli

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
	"os/exec"
)

// Constants for backoff strategy
const (
	initialRetryDelay = 2 * time.Second
	backoffMultiplier = 2
	maxRetryDelay     = 16 * time.Second
	jitterFactor      = 0.1 // 10% jitter
)

// executeDockerCommand executes a Docker command with retry and timeout mechanisms.
// It employs exponential backoff with jitter to handle transient errors gracefully.
func executeDockerCommand(ctx context.Context, retries int, command string, args ...string) ([]byte, error) {
	var lastErr error
	retryDelay := initialRetryDelay

	for attempt := 1; attempt <= retries; attempt++ {
		// Check if context is done before executing
		select {
		case <-ctx.Done():
			// If the context has been cancelled, log and return the context error.
			log.Error().
				Int("attempt", attempt).
				Str("command", command).
				Strs("args", args).
				Err(ctx.Err()).
				Msg("Command execution aborted due to context cancellation")
			return nil, fmt.Errorf("command execution aborted due to context cancellation: %w", ctx.Err())
		default:
			// Continue with execution
		}

		// Create the command with context
		cmd := exec.CommandContext(ctx, command, args...)
		startTime := time.Now()

		// Log the execution attempt
		log.Debug().
			Str("command", command).
			Strs("args", args).
			Int("attempt", attempt).
			Msg("Executing Docker command")

		// Execute the command and capture combined output
		output, err := cmd.CombinedOutput()
		duration := time.Since(startTime)

		// Determine if the context caused the command to fail
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
			// Wrap the error with attempt and output information
			lastErr = fmt.Errorf("attempt %d: command failed: %w, output: %s", attempt, err, string(output))
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

		// If not successful and not the last attempt, wait before retrying
		if attempt < retries {
			// Calculate delay with jitter
			jitter := time.Duration(float64(retryDelay) * jitterFactor * (rand.Float64()*2 - 1)) // +/- jitterFactor * retryDelay
			sleepDuration := retryDelay + jitter
			if sleepDuration < 0 {
				sleepDuration = 0
			}

			log.Warn().
				Int("attempt", attempt).
				Dur("retry_delay", sleepDuration).
				Msg("Retrying command after delay")

			// Wait for the calculated duration or until context is canceled
			select {
			case <-time.After(sleepDuration):
				// Continue to next attempt
			case <-ctx.Done():
				log.Error().
					Int("attempt", attempt).
					Str("command", command).
					Strs("args", args).
					Err(ctx.Err()).
					Msg("Command execution aborted during retry delay due to context cancellation")
				return nil, fmt.Errorf("command execution aborted during retry delay: %w", ctx.Err())
			}

			// Exponential backoff
			retryDelay *= time.Duration(backoffMultiplier)
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
		}
	}

	return nil, fmt.Errorf("command failed after %d attempts: %w", retries, lastErr)
}
