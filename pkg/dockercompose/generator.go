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
	if outputPath == "" {
		log.Error().Msg("Output path is empty. Please provide a valid path for the Docker Compose file.")
		return fmt.Errorf("output path cannot be empty")
	}

	outputDir := filepath.Dir(outputPath)
	log.Info().Str("outputPath", outputPath).Msg("Generating Docker Compose file")

	// Ensure the output directory exists.
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Error().Err(err).Str("outputDir", outputDir).Msg("Failed to create output directory")
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}
	log.Debug().Str("outputDir", outputDir).Msg("Output directory verified/created successfully")

	// Render the template with the composeFile data.
	var renderedContent strings.Builder
	if err := tmpl.Execute(&renderedContent, composeFile); err != nil {
		log.Error().Err(err).Msg("Failed to apply template")
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Create a temporary file in the same directory for atomic write.
	tempFile, err := os.CreateTemp(outputDir, "docker-compose-*.yaml")
	if err != nil {
		log.Error().Err(err).Str("outputDir", outputDir).Msg("Failed to create temporary file")
		return fmt.Errorf("failed to create temporary file in %s: %w", outputDir, err)
	}
	defer func() {
		if cerr := tempFile.Close(); cerr != nil {
			err = fmt.Errorf("failed to close temp file: %v", cerr)
		}
		defer func() {
			if err := os.Remove(tempFile.Name()); err != nil {
				log.Error().Err(err).Msg("Failed to remove temporary file")
			}
		}()
	}()
	log.Debug().Str("tempFile", tempFile.Name()).Msg("Temporary file created for atomic write")

	// Write the rendered content to the temporary file.
	if _, err := tempFile.WriteString(renderedContent.String()); err != nil {
		log.Error().Err(err).Str("tempFile", tempFile.Name()).Msg("Failed to write to temporary file")
		return fmt.Errorf("failed to write to temporary file %s: %w", tempFile.Name(), err)
	}
	log.Debug().Str("tempFile", tempFile.Name()).Msg("Rendered content written to temporary file successfully")

	// Close the temporary file before renaming
	if err := tempFile.Close(); err != nil {
		log.Error().Err(err).Str("tempFile", tempFile.Name()).Msg("Failed to close temporary file")
		return fmt.Errorf("failed to close temporary file %s: %w", tempFile.Name(), err)
	}

	// Rename the temporary file to the final output file (atomic operation).
	if err := os.Rename(tempFile.Name(), outputPath); err != nil {
		log.Error().Err(err).Str("tempFile", tempFile.Name()).Str("outputPath", outputPath).Msg("Failed to rename temporary file to final output file")
		return fmt.Errorf("failed to rename temporary file %s to output file %s: %w", tempFile.Name(), outputPath, err)
	}

	log.Info().Str("outputPath", outputPath).Msg("Docker Compose file generated successfully")
	return nil
}
