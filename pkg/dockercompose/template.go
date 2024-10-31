package dockercompose

import (
	"fmt"
	embedfiles "kasmlink/embedded"
	"text/template"
)

// LoadEmbeddedTemplate loads the embedded Docker Compose template from the embedded file system.
func LoadEmbeddedTemplate() (*template.Template, error) {
	// Read the template content from the embedded filesystem
	templateContent, err := embedfiles.EmbeddedTemplateFS.ReadFile("templates/docker-compose-template.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded template: %v", err)
	}

	// Parse the template content
	tmpl, err := template.New("docker-compose").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded template: %v", err)
	}
	return tmpl, nil
}
