package dockercompose

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"text/template"
)

const defaultTemplatePath = "templates/docker-compose-template.yaml"

// LoadEmbeddedTemplate loads the embedded Docker Compose template from the embedded file system.
func LoadEmbeddedTemplate(templatePath string) (*template.Template, error) {
	// Use the default template path if none is provided
	if templatePath == "" {
		templatePath = defaultTemplatePath
		log.Debug().Str("templatePath", templatePath).Msg("No templatePath specified, using default")
	}

	log.Info().Str("templatePath", templatePath).Msg("Attempting to load embedded template")

	// Attempt to read the template content from the embedded filesystem
	templateContent, err := embedfiles.EmbeddedTemplateFS.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to load embedded template")
		return nil, fmt.Errorf("failed to load embedded template at %s: %w", templatePath, err)
	}

	// Sanity check for empty template content
	if len(templateContent) == 0 {
		log.Error().Str("templatePath", templatePath).Msg("Template content is empty")
		return nil, errors.New("embedded template content is empty")
	}

	log.Info().Str("templatePath", templatePath).Msg("Embedded template loaded successfully")

	// Parse the template content
	tmpl, err := template.New("docker-compose").Parse(string(templateContent))
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to parse embedded template")
		return nil, fmt.Errorf("failed to parse embedded template at %s: %w", templatePath, err)
	}

	log.Info().Str("templatePath", templatePath).Msg("Embedded template parsed successfully")
	return tmpl, nil
}
