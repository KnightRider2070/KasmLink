package userParser

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"kasmlink/pkg/api/models"
	"os"
	"sync"
)

// UsersConfig represents the top-level structure of the YAML configuration.
type UsersConfig struct {
	UserDetails []UserDetails `yaml:"user_details"`
}

// UserDetails represents details for a specific user in the configuration.
type UserDetails struct {
	TargetUser             models.TargetUser `yaml:"target_user"`
	Role                   string            `yaml:"role"`
	AssignedContainerTag   string            `yaml:"assigned_container_tag"`
	AssignedContainerId    string            `yaml:"assigned_container_id"`
	KasmSessionOfContainer string            `yaml:"kasm_session_of_container"`
	Network                string            `yaml:"network"`
	VolumeMounts           map[string]string `yaml:"volume-mounts"`
	EnvironmentArgs        map[string]string `yaml:"environment_args"`
}

// UserParser is responsible for managing and updating user configurations.
type UserParser struct {
	mutex sync.Mutex
}

// NewUserParser creates and returns a new UserParser instance.
func NewUserParser() *UserParser {
	return &UserParser{}
}

func (u *UserParser) LoadConfig(path string) (*UsersConfig, error) {
	log.Debug().Str("path", path).Msg("Loading configuration file")
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to open configuration file")
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer file.Close()

	// Decode YAML into UsersConfig
	var config UsersConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to decode YAML configuration")
		return nil, fmt.Errorf("failed to decode YAML configuration: %w", err)
	}

	// Ensure TargetUser fields are correctly handled
	for i := range config.UserDetails {
		user := &config.UserDetails[i]
		if user.TargetUser.UserID == "" {
			log.Warn().Str("username", user.TargetUser.Username).Msg("UserID is empty in the loaded configuration")
		}
	}

	log.Debug().Str("path", path).Int("user_count", len(config.UserDetails)).Msg("Configuration loaded successfully")
	return &config, nil
}

// Helper to map generic map to models.TargetUser
func mapToTargetUser(input map[string]interface{}) models.TargetUser {
	return models.TargetUser{
		Username: getStringFromMap(input, "username"),
		UserID:   getStringFromMap(input, "user_id"),
		// Add other fields here if needed
	}
}

// Helper to safely get string from map
func getStringFromMap(input map[string]interface{}, key string) string {
	if value, ok := input[key].(string); ok {
		return value
	}
	return ""
}

// Helper to convert map[string]interface{} to map[string]string
func mapStringInterfaceToStringString(input map[string]interface{}) map[string]string {
	output := make(map[string]string)
	for key, value := range input {
		if strValue, ok := value.(string); ok {
			output[key] = strValue
		}
	}
	return output
}

func (u *UserParser) SaveConfig(path string, config *UsersConfig) error {
	log.Debug().Str("path", path).Msg("Saving updated configuration to file")

	// Marshal UsersConfig to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to marshal updated configuration")
		return fmt.Errorf("failed to marshal updated configuration: %w", err)
	}

	// Write YAML to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to write updated configuration to file")
		return fmt.Errorf("failed to write updated configuration to file: %w", err)
	}

	log.Info().Str("path", path).Msg("Configuration updated successfully")
	return nil
}

func (u *UserParser) UpdateUserConfig(path, username, newUserID, newKasmSessionID, containerID string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	log.Info().Str("username", username).Str("path", path).Msg("Updating user in configuration")

	config, err := u.LoadConfig(path)
	if err != nil {
		return err
	}

	if !updateUserDetails(config, username, newUserID, newKasmSessionID, containerID) {
		log.Warn().Str("username", username).Msg("User not found in configuration")
		return fmt.Errorf("user %s not found in configuration", username)
	}

	if err := u.SaveConfig(path, config); err != nil {
		return err
	}

	log.Info().
		Str("username", username).
		Str("user_id", newUserID).
		Str("session_id", newKasmSessionID).
		Str("container_id", containerID).
		Msg("User updated successfully")
	return nil
}

// updateUserDetails modifies the user details in the configuration.
// Returns true if the user was found and updated, otherwise false.
func updateUserDetails(config *UsersConfig, username, newUserID, newKasmSessionID, containerID string) bool {
	for i, user := range config.UserDetails {
		if user.TargetUser.Username == username {
			config.UserDetails[i].AssignedContainerId = containerID
			config.UserDetails[i].TargetUser.UserID = newUserID
			config.UserDetails[i].KasmSessionOfContainer = newKasmSessionID
			log.Debug().
				Str("username", username).
				Str("user_id", newUserID).
				Str("session_id", newKasmSessionID).
				Str("container_id", containerID).
				Msg("User details updated")
			return true
		}
	}
	return false
}

// ValidateConfig validates the loaded configuration for required fields.
func (u *UserParser) ValidateConfig(config *UsersConfig) error {
	for _, user := range config.UserDetails {
		if user.TargetUser.Username == "" {
			log.Error().Msg("Invalid configuration: Missing username")
			return errors.New("invalid configuration: missing username")
		} else if user.TargetUser.UserID == "" {
			log.Error().Msg("Invalid configuration: Missing user ID")
			return errors.New("invalid configuration: missing user ID")
		}
	}
	log.Debug().Msg("Configuration validation passed")
	return nil
}
