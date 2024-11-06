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
	if templatePath == "" {
		templatePath = "templates/docker-compose-template.yaml"
	}
	log.Info().Str("templatePath", templatePath).Msg("Attempting to load embedded template")

	// Attempt to read the template content from the embedded filesystem
	templateContent, err := embedfiles.EmbeddedTemplateFS.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to load embedded template")
		return nil, fmt.Errorf("failed to load embedded template at %s: %v", templatePath, err)
	}
	log.Info().Str("templatePath", templatePath).Msg("Embedded template loaded successfully")

	// Parse the template content
	tmpl, err := template.New("docker-compose").Parse(string(templateContent))
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to parse embedded template")
		return nil, fmt.Errorf("failed to parse embedded template at %s: %v", templatePath, err)
	}
	log.Info().Str("templatePath", templatePath).Msg("Embedded template parsed successfully")

	return tmpl, nil
}
