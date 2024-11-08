package dockercompose

import (
	"fmt"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"text/template"
)

// LoadEmbeddedTemplate loads the embedded Docker Compose template from the embedded file system.
func LoadEmbeddedTemplate(templatePath string) (*template.Template, error) {
	// Check if the template path is specified; otherwise, use the default
	defaultTemplatePath := "templates/docker-compose-template.yaml"
	if templatePath == "" {
		templatePath = defaultTemplatePath
		log.Debug().Str("templatePath", templatePath).Msg("No template path provided. Using default template path")
	} else {
		log.Debug().Str("templatePath", templatePath).Msg("Using provided template path")
	}

	// Attempt to read the template content from the embedded filesystem
	log.Info().Str("templatePath", templatePath).Msg("Attempting to load embedded template")
	templateContent, err := embedfiles.EmbeddedTemplateFS.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to load embedded template. Ensure the template path is correct and the file is available in the embedded filesystem")
		return nil, fmt.Errorf("failed to load embedded template at %s: %w", templatePath, err)
	}
	log.Debug().Str("templatePath", templatePath).Msg("Successfully loaded embedded template content")

	// Parse the template content
	tmpl, err := template.New("docker-compose").Parse(string(templateContent))
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to parse embedded template. Verify template syntax and content")
		return nil, fmt.Errorf("failed to parse embedded template at %s: %w", templatePath, err)
	}
	log.Info().Str("templatePath", templatePath).Msg("Embedded template parsed successfully")

	return tmpl, nil
}
