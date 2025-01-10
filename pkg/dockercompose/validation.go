package dockercompose

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	embedfiles "kasmlink/embedded"
)

func ValidateDockerCompose(compose DockerCompose) error {
	schemaLoader := gojsonschema.NewBytesLoader(embedfiles.ComposeSpec)

	// Convert DockerCompose struct to JSON
	jsonData, err := json.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal DockerCompose to JSON: %w", err)
	}

	// Log JSON for debugging
	log.Debug().Msgf("Validating JSON data: %s", string(jsonData))

	// Validate JSON against schema
	documentLoader := gojsonschema.NewStringLoader(string(jsonData))
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var validationErrors string
		for _, err := range result.Errors() {
			validationErrors += fmt.Sprintf("- %s\n", err.String())
		}
		log.Error().Msgf("Validation failed: %s", validationErrors)
		return fmt.Errorf("validation failed:\n%s", validationErrors)
	}

	log.Info().Msg("Validation succeeded")
	return nil
}
