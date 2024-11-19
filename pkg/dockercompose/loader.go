package dockercompose

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/rs/zerolog/log"
)

// LoadComposeFile loads the Docker Compose configuration from a YAML file.
func LoadComposeFile(configPath string) (*ComposeFile, error) {
	log.Info().Str("configPath", configPath).Msg("Loading Docker Compose configuration from file")

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Error().Err(err).Str("configPath", configPath).Msg("Failed to open configuration file")
		return nil, fmt.Errorf("failed to open configuration file %s: %w", configPath, err)
	}

	var composeFile ComposeFile
	err = yaml.Unmarshal(file, &composeFile)
	if err != nil {
		log.Error().Err(err).Str("configPath", configPath).Msg("Failed to decode YAML configuration file into ComposeFile struct")
		return nil, fmt.Errorf("failed to decode configuration file %s: %w", configPath, err)
	}

	log.Info().Str("configPath", configPath).Msg("Docker Compose configuration loaded successfully")
	return &composeFile, nil
}
