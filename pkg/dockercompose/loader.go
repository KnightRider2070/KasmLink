package dockercompose

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// LoadComposeFile loads the Compose configuration from a YAML file.
func LoadComposeFile(configPath string) (*ComposeFile, error) {
	log.Info().Msgf("Loading Docker Compose configuration from file: %s", configPath)

	// Check if the file exists and is accessible
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Error().Msgf("Configuration file does not exist at path: %s", configPath)
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open config file: %s", configPath)
		return nil, fmt.Errorf("failed to open config file %s: %w", configPath, err)
	}
	defer file.Close()

	var composeFile ComposeFile
	decoder := yaml.NewDecoder(file)

	// Attempt to decode the YAML file into the ComposeFile struct
	if err := decoder.Decode(&composeFile); err != nil {
		log.Error().Err(err).Msg("Failed to decode YAML config file into ComposeFile struct")
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Initialize any nil maps to prevent runtime errors during usage
	ensureMaps(&composeFile)

	log.Info().Msg("Docker Compose configuration loaded successfully")
	return &composeFile, nil
}

// ensureMaps initializes any nil maps in the ComposeFile structure to avoid runtime errors.
func ensureMaps(composeFile *ComposeFile) {
	if composeFile.Services == nil {
		composeFile.Services = make(map[string]Service)
		log.Debug().Msg("Initialized Services map in ComposeFile")
	}
	if composeFile.Networks == nil {
		composeFile.Networks = make(map[string]Network)
		log.Debug().Msg("Initialized Networks map in ComposeFile")
	}
	if composeFile.Volumes == nil {
		composeFile.Volumes = make(map[string]Volume)
		log.Debug().Msg("Initialized Volumes map in ComposeFile")
	}
	if composeFile.Configs == nil {
		composeFile.Configs = make(map[string]Config)
		log.Debug().Msg("Initialized Configs map in ComposeFile")
	}
	if composeFile.Secrets == nil {
		composeFile.Secrets = make(map[string]Secret)
		log.Debug().Msg("Initialized Secrets map in ComposeFile")
	}
}
