package dockercompose

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/rs/zerolog/log"
)

// LoadComposeFile loads the Compose configuration from a YAML file.
func LoadComposeFile(configPath string) (*ComposeFile, error) {
	log.Info().Str("configPath", configPath).Msg("Loading Docker Compose configuration from file")

	// Check if the file exists and is accessible
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error().Str("configPath", configPath).Msg("Configuration file does not exist")
			return nil, fmt.Errorf("configuration file does not exist at path: %s", configPath)
		}
		log.Error().Err(err).Str("configPath", configPath).Msg("Error accessing the configuration file")
		return nil, fmt.Errorf("error accessing configuration file at %s: %w", configPath, err)
	}

	// Check file permissions to ensure it's readable
	if fileInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		log.Error().Str("configPath", configPath).Msg("Configuration file is not readable")
		return nil, fmt.Errorf("configuration file is not readable: %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Error().Err(err).Str("configPath", configPath).Msg("Failed to open configuration file")
		return nil, fmt.Errorf("failed to open configuration file %s: %w", configPath, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = fmt.Errorf("failed to close configuration file: %v", cerr)
		}
	}()

	var composeFile ComposeFile
	decoder := yaml.NewDecoder(file)

	// Attempt to decode the YAML file into the ComposeFile struct
	if err := decoder.Decode(&composeFile); err != nil {
		log.Error().Err(err).Str("configPath", configPath).Msg("Failed to decode YAML configuration file into ComposeFile struct")
		return nil, fmt.Errorf("failed to decode configuration file %s: %w", configPath, err)
	}

	// Initialize any nil maps to prevent runtime errors during usage
	ensureMaps(&composeFile)

	log.Info().Str("configPath", configPath).Msg("Docker Compose configuration loaded successfully")
	return &composeFile, nil
}

// ensureMaps initializes any nil maps in the ComposeFile structure to avoid runtime errors.
func ensureMaps(composeFile *ComposeFile) {
	// Ensure each map is initialized to prevent runtime errors
	if composeFile.Services == nil {
		composeFile.Services = make(map[string]Service)
		log.Debug().Str("mapType", "Services").Msg("Initialized Services map in ComposeFile")
	}
	if composeFile.Networks == nil {
		composeFile.Networks = make(map[string]Network)
		log.Debug().Str("mapType", "Networks").Msg("Initialized Networks map in ComposeFile")
	}
	if composeFile.Volumes == nil {
		composeFile.Volumes = make(map[string]Volume)
		log.Debug().Str("mapType", "Volumes").Msg("Initialized Volumes map in ComposeFile")
	}
	if composeFile.Configs == nil {
		composeFile.Configs = make(map[string]Config)
		log.Debug().Str("mapType", "Configs").Msg("Initialized Configs map in ComposeFile")
	}
	if composeFile.Secrets == nil {
		composeFile.Secrets = make(map[string]Secret)
		log.Debug().Str("mapType", "Secrets").Msg("Initialized Secrets map in ComposeFile")
	}
}
