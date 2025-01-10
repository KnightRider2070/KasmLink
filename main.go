package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"kasmlink/cmd"
)

var Version = "dev"
var noColor = false

func LoadLogo(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to load logo from %s: %v", filename, err)
	}
	return string(content), nil
}

func main() {
	// Configure zerolog
	zerolog.DurationFieldUnit = time.Second

	// Set the global log level based on the LOGLEVEL environment variable
	var zerologLevel zerolog.Level
	switch os.Getenv("LOGLEVEL") {
	case "trace":
		zerologLevel = zerolog.TraceLevel
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "warn":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	case "fatal":
		zerologLevel = zerolog.FatalLevel
	case "panic":
		zerologLevel = zerolog.PanicLevel
	case "info":
		zerologLevel = zerolog.InfoLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	if os.Getenv("DEBUG") != "" {
		noColor = true
	}

	zerolog.SetGlobalLevel(zerologLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    noColor,
	})

	logo, err := LoadLogo("kasmlink.txt")
	if err != nil {
		log.Error().Msgf("Error loading logo: %v", err)
	} else {
		fmt.Printf("\n%s\n", logo)
	}
	fmt.Printf("---\nKasm Link CLI Version: %s\n---\n", Version)

	cmd.Execute()
}
