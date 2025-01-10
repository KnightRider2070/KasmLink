package talos

import (
	"encoding/json"
	"fmt"
	"github.com/budimanjojo/talhelper/v3/pkg/config"
	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog/log"
	programmConfig "kasmlink/internal/config"
	"os"
	"path/filepath"
)

// GenSchema generates a JSON schema for the Talos configuration and saves it to the specified directory.
func GenSchema() error {
	cfg := config.TalhelperConfig{}
	reflector := new(jsonschema.Reflector)
	reflector.FieldNameTag = "yaml"
	reflector.RequiredFromJSONSchemaTags = true

	schemaDir := filepath.Join(programmConfig.ConfigPaths.TalosDir)
	if err := os.MkdirAll(schemaDir, os.ModePerm); err != nil {
		log.Error().Err(err).Msgf("Failed to create schema directory: %s", schemaDir)
		return fmt.Errorf("failed to create schema directory: %w", err)
	}

	schema := reflector.Reflect(&cfg)
	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal schema to JSON")
		return fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	schemaFilePath := filepath.Join(schemaDir, "talconfig.json")
	if err := os.WriteFile(schemaFilePath, schemaData, 0644); err != nil {
		log.Error().Err(err).Msgf("Failed to write schema file: %s", schemaFilePath)
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	log.Info().Msgf("Schema generated and saved successfully to: %s", schemaFilePath)
	return nil
}
