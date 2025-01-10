package dockercli

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

// RetryWithBackoff retries a function with exponential backoff and jitter.
func (dc *DockerClient) RetryWithBackoff(ctx context.Context, operation func() error) error {
	delay := dc.initialRetryDelay

	for attempt := 0; attempt < dc.retries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		log.Warn().Err(err).Int("attempt", attempt+1).Msg("Operation failed, retrying")

		if err := waitForRetry(ctx, delay); err != nil {
			return err
		}
		delay = applyBackoff(delay, dc.backoffMultiplier, int(dc.maxRetryDelay.Seconds()), dc.jitterFactor)
	}

	return errors.New("operation failed after retries")
}

func applyBackoff(delay time.Duration, multiplier, maxDelay int, jitter float64) time.Duration {
	newDelay := delay * time.Duration(multiplier)
	if newDelay > time.Duration(maxDelay)*time.Second {
		newDelay = time.Duration(maxDelay) * time.Second
	}
	return newDelay + jitterDuration(delay, jitter)
}

func jitterDuration(delay time.Duration, jitterFactor float64) time.Duration {
	jitter := time.Duration(float64(delay) * jitterFactor * (rand.Float64()*2 - 1))
	return jitter
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("retry aborted due to context cancellation: %w", ctx.Err())
	}
}
