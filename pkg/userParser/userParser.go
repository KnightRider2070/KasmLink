package userParser

import (
	"fmt"
	"io/ioutil"
	"kasmlink/pkg/api"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// UserDetails represents a user entry in the YAML file
type UserDetails struct {
	api.TargetUser
	Role                   string            `yaml:"Role"`
	AssignedContainerTag   string            `yaml:"AssignedContainerTag"`
	KasmSessionOfContainer string            `yaml:"KasmSessionOfContainer"`
	EnvironmentArgs        map[string]string `yaml:"environmentArgs"`
}

// UsersConfig represents the structure of the YAML configuration file
type UsersConfig struct {
	UserDetails []UserDetails `yaml:"UserDetails"`
}

// LoadConfig loads the configuration from a YAML file into a UsersConfig struct
func LoadConfig(yamlFilePath string) (*UsersConfig, error) {
	log.Info().Str("yamlFilePath", yamlFilePath).Msg("Loading configuration from YAML file")

	// Read the YAML file
	data, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read YAML file")
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse the YAML content
	var config UsersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Error().Err(err).Msg("Failed to parse YAML file")
		return nil, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	// Log details for debugging
	log.Info().Int("userCount", len(config.UserDetails)).Msg("Loaded user details successfully")

	return &config, nil
}

// UpdateUserConfig updates the userId and KasmSessionOfContainer for a specific user
func UpdateUserConfig(yamlFilePath string, username, userId, kasmSessionID string) error {
	log.Info().Str("yamlFilePath", yamlFilePath).Msg("Updating user configuration")

	// Load existing config
	config, err := LoadConfig(yamlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load configuration for update: %w", err)
	}

	// Find the user and update the fields
	updated := false
	for i, user := range config.UserDetails {
		if user.Username == username {
			log.Debug().Str("username", username).Msg("Updating user details")
			config.UserDetails[i].TargetUser.UserID = userId
			config.UserDetails[i].KasmSessionOfContainer = kasmSessionID
			updated = true
			break
		}
	}

	if !updated {
		log.Warn().Str("username", username).Msg("User not found in configuration")
		return fmt.Errorf("user %s not found in configuration", username)
	}

	// Marshal the updated config back to YAML
	updatedData, err := yaml.Marshal(config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal updated configuration")
		return fmt.Errorf("failed to marshal updated configuration: %w", err)
	}

	// Write the updated YAML back to the file
	if err := os.WriteFile(yamlFilePath, updatedData, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write updated configuration back to YAML file")
		return fmt.Errorf("failed to write updated configuration: %w", err)
	}

	log.Info().Str("username", username).Msg("User configuration updated successfully")
	return nil
}
