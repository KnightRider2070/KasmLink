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
	WorkspaceID    string             `yaml:"workspace_id"`    // Unique identifier for the workspace
	ImageConfig    models.TargetImage `yaml:"image_config"`    // Target image configuration
	DockerFilePath string             `yaml:"dockerfile_path"` // Path to the Dockerfile defining the workspace
	TargetStage    string             `yaml:"target_stage"`    // Target stage for the Docker build if multi-stage
}

// DeploymentConfig represents the full YAML structure with shared workspaces and userService details.
type DeploymentConfig struct {
	Workspaces []WorkspaceConfig    `yaml:"workspaces"` // Shared workspace configurations
	Users      []UserDetails        `yaml:"users"`      // Users with references to workspaces
	Groups     []models.GroupStruct `yaml:"groups"`     // User groups
}

// UserDetails represents details for a specific userService in the configuration.
type UserDetails struct {
	TargetUser    models.TargetUser `yaml:"target_user"`     // Target userService details
	GroupName     string            `yaml:"group"`           // Group name
	KasmSessionID string            `yaml:"kasm_session_id"` // Kasm session ID
	Environment   map[string]string `yaml:"environment"`     // User-specific environment variables
	VolumeMounts  map[string]string `yaml:"volume_mounts"`   // User-specific volume mounts
}

// UserParser is responsible for managing and updating configurations.
type UserParser struct {
	mutex       sync.Mutex
	configCache map[string]*DeploymentConfig
}

// NewUserParser creates and returns a new UserParser instance.
func NewUserParser() *UserParser {
	return &UserParser{
		configCache: make(map[string]*DeploymentConfig),
	}
}

// normalizeDeploymentConfig ensures all fields are consistently initialized.
func normalizeDeploymentConfig(config *DeploymentConfig) {
	for i := range config.Workspaces {
		if config.Workspaces[i].ImageConfig.RestrictNetworkNames == nil {
			config.Workspaces[i].ImageConfig.RestrictNetworkNames = []string{}
		}
		if config.Workspaces[i].ImageConfig.Categories == nil {
			config.Workspaces[i].ImageConfig.Categories = []string{}
		}
		if config.Workspaces[i].ImageConfig.ExecConfig == nil {
			config.Workspaces[i].ImageConfig.ExecConfig = map[string]interface{}{}
		}
		if config.Workspaces[i].ImageConfig.LaunchConfig == nil {
			config.Workspaces[i].ImageConfig.LaunchConfig = map[string]interface{}{}
		}
	}

	for i := range config.Users {
		if config.Users[i].Environment == nil {
			config.Users[i].Environment = map[string]string{}
		}
		if config.Users[i].VolumeMounts == nil {
			config.Users[i].VolumeMounts = map[string]string{}
		}
	}
}

// LoadDeploymentConfig loads the YAML configuration from a file.
func (u *UserParser) LoadDeploymentConfig(path string) (*DeploymentConfig, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Return cached config if it exists
	if config, exists := u.configCache[path]; exists {
		log.Debug().Str("path", path).Msg("Returning cached configuration")
		return config, nil
	}

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

	normalizeDeploymentConfig(&config)
	u.configCache[path] = &config

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

	u.mutex.Lock()
	defer u.mutex.Unlock()

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to write updated configuration to file")
		return fmt.Errorf("failed to write updated configuration to file: %w", err)
	}

	// Update cache
	u.configCache[path] = config

	log.Info().Str("path", path).Msg("Configuration updated successfully")
	return nil
}

// UpdateUserDetails updates a specific userService's configuration.
func (u *UserParser) UpdateUserDetails(path, username, newUserID, newKasmSessionID string) error {
	log.Info().Str("username", username).Str("path", path).Msg("Updating userService in configuration")

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
		return fmt.Errorf("userService %s not found in configuration", username)
	}

	return u.SaveDeploymentConfig(path, config)
}

// ValidateDeploymentConfig validates the entire deployment configuration.
func (u *UserParser) ValidateDeploymentConfig(config *DeploymentConfig) error {
	workspaceIDs := make(map[string]bool)
	for _, workspace := range config.Workspaces {
		if workspace.WorkspaceID == "" {
			return errors.New("workspace ID cannot be empty")
		}
		if workspaceIDs[workspace.WorkspaceID] {
			return fmt.Errorf("duplicate workspace ID found: %s", workspace.WorkspaceID)
		}
		workspaceIDs[workspace.WorkspaceID] = true
	}

	for _, user := range config.Users {
		if user.TargetUser.Username == "" {
			return errors.New("userService must have a username")
		}
		if user.GroupName == "" {
			return errors.New("userService must have a group name")
		}
	}
	return nil
}

// FindUserByUsername searches for a userService by username.
func (u *UserParser) FindUserByUsername(config *DeploymentConfig, username string) (*UserDetails, error) {
	for _, user := range config.Users {
		if user.TargetUser.Username == username {
			log.Debug().Str("username", username).Msg("User found in configuration")
			return &user, nil
		}
	}
	log.Warn().Str("username", username).Msg("User not found in configuration")
	return nil, fmt.Errorf("userService %s not found in configuration", username)
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

// AddUser adds a new userService to the configuration.
func (u *UserParser) AddUser(config *DeploymentConfig, user UserDetails) error {
	for _, existingUser := range config.Users {
		if existingUser.TargetUser.Username == user.TargetUser.Username {
			return fmt.Errorf("userService %s already exists", user.TargetUser.Username)
		}
	}

	config.Users = append(config.Users, user)
	return nil
}
