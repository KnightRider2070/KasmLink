package userParser

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/webApi"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type UsersConfig struct {
	UserDetails []UserDetails `yaml:"user_details"`
}

type UserDetails struct {
	TargetUser             webApi.TargetUser `yaml:"target_user"`
	Role                   string            `yaml:"role"`
	AssignedContainerTag   string            `yaml:"assigned_container_tag"`
	KasmSessionOfContainer string            `yaml:"kasm_session_of_container"`
	EnvironmentArgs        map[string]string `yaml:"environment_args"`
}

type UserParser struct {
	mutex sync.Mutex
}

func NewUserParser() *UserParser {
	return &UserParser{}
}

// LoadConfig loads the configuration from the YAML file.
// Assumes that the caller holds the mutex.
func (u *UserParser) LoadConfig(path string) (*UsersConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config UsersConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateUserConfig updates the user configuration.
func (u *UserParser) UpdateUserConfig(path, username, newUserID, newKasmSessionID string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	log.Printf("Updating user %s in configuration file %s\n", username, path)

	config, err := u.LoadConfig(path)
	if err != nil {
		log.Printf("Error loading configuration: %v\n", err)
		return err
	}

	found := false
	for i, user := range config.UserDetails {
		if user.TargetUser.Username == username {
			config.UserDetails[i].TargetUser.UserID = newUserID
			config.UserDetails[i].KasmSessionOfContainer = newKasmSessionID
			found = true
			log.Printf("Updated user %s with new UserID %s and KasmSessionID %s\n", username, newUserID, newKasmSessionID)
			break
		}
	}

	if !found {
		log.Printf("User %s not found in configuration\n", username)
		return errors.New("user " + username + " not found in configuration")
	}

	// Marshal and write the updated configuration
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("Failed to marshal updated configuration: %v\n", err)
		return fmt.Errorf("failed to marshal updated configuration: %w", err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Printf("Failed to write updated configuration to YAML file: %v\n", err)
		return fmt.Errorf("failed to write updated configuration to YAML file: %w", err)
	}

	log.Printf("Successfully updated user %s\n", username)
	return nil
}
