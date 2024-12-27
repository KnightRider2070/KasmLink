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

type WorkspaceConfig struct {
	WorkspaceID string             `yaml:"workspace_id"` // Unique identifier for the workspace
	ImageConfig models.TargetImage `yaml:"image_config"` // Target image configuration
}

// DeploymentConfig represents the full YAML structure with shared workspaces and user details.
type DeploymentConfig struct {
	Workspaces []WorkspaceConfig `yaml:"workspaces"` // Shared workspace configurations
	Users      []UserDetails     `yaml:"users"`      // Users with references to workspaces
}

// UserDetails represents details for a specific user in the configuration.
type UserDetails struct {
	TargetUser    models.TargetUser `yaml:"target_user"`     // Target user details
	WorkspaceID   string            `yaml:"workspace_id"`    // Reference to a shared workspace configuration
	KasmSessionID string            `yaml:"kasm_session_id"` // Kasm session ID
	Environment   map[string]string `yaml:"environment"`     // User-specific environment variables
	VolumeMounts  map[string]string `yaml:"volume_mounts"`   // User-specific volume mounts
}

// UserParser is responsible for managing and updating configurations.
type UserParser struct {
	mutex sync.Mutex
}

// NewUserParser creates and returns a new UserParser instance.
func NewUserParser() *UserParser {
	return &UserParser{}
}

// LoadDeploymentConfig loads the YAML configuration from a file.
func (u *UserParser) LoadDeploymentConfig(path string) (*DeploymentConfig, error) {
	log.Debug().Str("path", path).Msg("Loading deployment configuration file")
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to open configuration file")
		return nil, fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer file.Close()

	var config DeploymentConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to decode YAML configuration")
		return nil, fmt.Errorf("failed to decode YAML configuration: %w", err)
	}

	log.Debug().Str("path", path).
		Int("workspace_count", len(config.Workspaces)).
		Int("user_count", len(config.Users)).
		Msg("Configuration loaded successfully")
	return &config, nil
}

// SaveDeploymentConfig saves the YAML configuration back to the file.
func (u *UserParser) SaveDeploymentConfig(path string, config *DeploymentConfig) error {
	log.Debug().Str("path", path).Msg("Saving updated configuration to file")

	data, err := yaml.Marshal(config)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to marshal updated configuration")
		return fmt.Errorf("failed to marshal updated configuration: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to write updated configuration to file")
		return fmt.Errorf("failed to write updated configuration to file: %w", err)
	}

	log.Info().Str("path", path).Msg("Configuration updated successfully")
	return nil
}

// UpdateUserDetails updates a specific user's configuration.
func (u *UserParser) UpdateUserDetails(path, username, newUserID, newKasmSessionID string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	log.Info().Str("username", username).Str("path", path).Msg("Updating user in configuration")

	config, err := u.LoadDeploymentConfig(path)
	if err != nil {
		return err
	}

	updated := false
	for i, user := range config.Users {
		if user.TargetUser.Username == username {
			config.Users[i].TargetUser.UserID = newUserID
			config.Users[i].KasmSessionID = newKasmSessionID
			log.Debug().
				Str("username", username).
				Str("user_id", newUserID).
				Str("session_id", newKasmSessionID).
				Msg("User details updated")
			updated = true
			break
		}
	}

	if !updated {
		log.Warn().Str("username", username).Msg("User not found in configuration")
		return fmt.Errorf("user %s not found in configuration", username)
	}

	if err := u.SaveDeploymentConfig(path, config); err != nil {
		return err
	}

	log.Info().Str("username", username).Msg("User updated successfully")
	return nil
}

// ValidateDeploymentConfig validates the entire deployment configuration.
func (u *UserParser) ValidateDeploymentConfig(config *DeploymentConfig) error {
	workspaceMap := make(map[string]WorkspaceConfig)
	for _, workspace := range config.Workspaces {
		workspaceMap[workspace.WorkspaceID] = workspace
	}

	for _, user := range config.Users {
		if user.TargetUser.Username == "" {
			log.Error().Msg("Invalid configuration: Missing username")
			return errors.New("invalid configuration: missing username")
		}
		if user.WorkspaceID == "" {
			log.Error().Msg("Invalid configuration: Missing workspace ID")
			return errors.New("invalid configuration: missing workspace ID")
		}
		if _, exists := workspaceMap[user.WorkspaceID]; !exists {
			log.Error().Str("workspace_id", user.WorkspaceID).Msg("Invalid configuration: Workspace ID not found")
			return fmt.Errorf("invalid configuration: workspace ID %s not found", user.WorkspaceID)
		}
	}

	log.Debug().Msg("Configuration validation passed")
	return nil
}

// FindUserByUsername searches for a user by username.
func (u *UserParser) FindUserByUsername(config *DeploymentConfig, username string) (*UserDetails, error) {
	for _, user := range config.Users {
		if user.TargetUser.Username == username {
			log.Debug().Str("username", username).Msg("User found in configuration")
			return &user, nil
		}
	}
	log.Warn().Str("username", username).Msg("User not found in configuration")
	return nil, fmt.Errorf("user %s not found in configuration", username)
}

// FindWorkspaceByID searches for a workspace by ID.
func (u *UserParser) FindWorkspaceByID(config *DeploymentConfig, workspaceID string) (*WorkspaceConfig, error) {
	for _, workspace := range config.Workspaces {
		if workspace.WorkspaceID == workspaceID {
			log.Debug().Str("workspace_id", workspaceID).Msg("Workspace found in configuration")
			return &workspace, nil
		}
	}
	log.Warn().Str("workspace_id", workspaceID).Msg("Workspace not found in configuration")
	return nil, fmt.Errorf("workspace %s not found in configuration", workspaceID)
}
