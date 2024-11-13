package cmd

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func CreateContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// HandleError handles an error by logging it and exiting the program if it's not nil.
func HandleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1) // Properly exit with an error code.
	}
}

// ParseInt helper function to parse string to int with a fallback default.
func ParseInt(arg string, fallback int) int {
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		log.Warn().Str("arg", arg).Msgf("Invalid number format, using default value %d", fallback)
		return fallback
	}
	return parsed
}

// ParseDuration helper function to parse string to time.Duration with a fallback default.
func ParseDuration(arg string, fallback time.Duration) time.Duration {
	parsed, err := time.ParseDuration(arg)
	if err != nil {
		log.Warn().Str("arg", arg).Msgf("Invalid duration format, using default value %v", fallback)
		return fallback
	}
	return parsed
}
