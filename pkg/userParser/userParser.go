package userParser

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kasmlink/pkg/api"
)

type UserYaml struct {
	UserID          string            `yaml:"user_id"`
	UserName        string            `yaml:"username"`
	ContainerImage  string            `yaml:"container_image"`
	EnvironmentArgs map[string]string `yaml:"environment_args"`
}

// WriteUsersToYaml writes user details to a YAML file.
func WriteUsersToYaml(users []api.UserResponse, containerImage string, environmentArgs map[string]string, outputPath string) error {
	var usersYaml []UserYaml

	for _, user := range users {
		userYaml := UserYaml{
			UserID:          user.UserID,
			UserName:        user.Username,
			ContainerImage:  containerImage,
			EnvironmentArgs: environmentArgs,
		}
		usersYaml = append(usersYaml, userYaml)
	}

	yamlData, err := yaml.Marshal(usersYaml)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal users to YAML")
		return fmt.Errorf("failed to marshal users to YAML: %v", err)
	}

	if err := ioutil.WriteFile(outputPath, yamlData, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write YAML file")
		return fmt.Errorf("failed to write YAML file: %v", err)
	}

	log.Info().Str("outputPath", outputPath).Msg("YAML file written successfully")
	return nil
}

// LoadUsersFromYaml loads user details from a YAML file.
func LoadUsersFromYaml(inputPath string) ([]UserYaml, error) {
	var usersYaml []UserYaml

	yamlData, err := ioutil.ReadFile(inputPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read YAML file")
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	if err := yaml.Unmarshal(yamlData, &usersYaml); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal YAML data")
		return nil, fmt.Errorf("failed to unmarshal YAML data: %v", err)
	}

	log.Info().Str("inputPath", inputPath).Msg("YAML file loaded successfully")
	return usersYaml, nil
}
