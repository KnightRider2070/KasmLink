package dockercompose

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateDockerComposeFile writes the Docker Compose file according to the template and structure.
func GenerateDockerComposeFile(tmpl *template.Template, composeFile ComposeFile, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	log.Info().Msgf("Generating Docker Compose file at: %s", outputPath)

	// Ensure the output directory exists.
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Error().Err(err).Msgf("Failed to create output directory: %s", outputDir)
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Render the template with the composeFile data.
	var renderedContent strings.Builder
	if err := tmpl.Execute(&renderedContent, composeFile); err != nil {
		log.Error().Err(err).Msg("Failed to apply template")
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Create the output file.
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create output file: %s", outputPath)
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer outputFile.Close()

	// Write the rendered content to the output file.
	if _, err := outputFile.WriteString(renderedContent.String()); err != nil {
		log.Error().Err(err).Msg("Failed to write to output file")
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	log.Info().Msgf("Docker Compose file generated successfully at %s", outputPath)
	return nil
}
