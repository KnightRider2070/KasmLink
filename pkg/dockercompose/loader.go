package dockercompose

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	embedfiles "kasmlink/embedded"
	"os"
)

func LoadAndParseComposeFile(configPath string) (DockerCompose, error) {
	log.Info().Msgf("Loading and parsing Docker Compose configuration from file: %s", configPath)

	if err := ensureFileExists(configPath); err != nil {
		return DockerCompose{}, err
	}

	yamlContent, err := decodeYAMLFile(configPath)
	if err != nil {
		return DockerCompose{}, err
	}

	// Validate raw YAML against schema before converting to struct
	jsonData, err := json.Marshal(yamlContent)
	if err != nil {
		return DockerCompose{}, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(embedfiles.ComposeSpec)
	documentLoader := gojsonschema.NewStringLoader(string(jsonData))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return DockerCompose{}, fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var validationErrors string
		for _, err := range result.Errors() {
			validationErrors += fmt.Sprintf("- %s\n", err.String())
		}
		return DockerCompose{}, fmt.Errorf("validation failed:\n%s", validationErrors)
	}

	log.Info().Msg("Validation succeeded")

	// Convert validated JSON to DockerCompose struct
	var dockerCompose DockerCompose
	if err := json.Unmarshal(jsonData, &dockerCompose); err != nil {
		return DockerCompose{}, fmt.Errorf("failed to parse JSON into DockerCompose struct: %w", err)
	}

	return dockerCompose, nil
}

// ensureFileExists checks if the given file exists and is accessible.
func ensureFileExists(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Error().Msgf("Configuration file does not exist at path: %s", configPath)
		return fmt.Errorf("config file does not exist: %s", configPath)
	}
	return nil
}

// decodeYAMLFile reads and decodes the given YAML file into a generic map.
func decodeYAMLFile(configPath string) (map[string]interface{}, error) {
	file, err := os.Open(configPath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open config file: %s", configPath)
		return nil, fmt.Errorf("failed to open config file %s: %w", configPath, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close config file")
		}
	}()

	var yamlContent map[string]interface{}
	decoder := yaml.NewDecoder(file)

	if err := decoder.Decode(&yamlContent); err != nil {
		log.Error().Err(err).Msg("Failed to decode YAML config file")
		return nil, fmt.Errorf("failed to decode YAML config file: %w", err)
	}
	return yamlContent, nil
}

// convertToDockerCompose converts a generic map (from YAML) to a DockerCompose struct.
func convertToDockerCompose(yamlContent map[string]interface{}) (DockerCompose, error) {
	jsonData, err := json.Marshal(yamlContent)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert YAML to JSON")
		return DockerCompose{}, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	var dockerCompose DockerCompose
	if err := json.Unmarshal(jsonData, &dockerCompose); err != nil {
		log.Error().Err(err).Msg("Failed to parse JSON into DockerCompose struct")
		return DockerCompose{}, fmt.Errorf("failed to parse JSON into DockerCompose struct: %w", err)
	}
	return dockerCompose, nil
}
