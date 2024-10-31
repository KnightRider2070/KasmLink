package procedures

import (
	_ "embed" // Required for embedding files
	"fmt"
	"io/fs"
	"io/ioutil"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// InitTemplatesFolder initializes the templates folder by copying embedded default templates.
func InitTemplatesFolder(folderPath string) error {
	// Create the target folder if it does not exist
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create templates folder: %v", err)
		}
	}

	// Walk through the embedded file system and write example templates to the target folder
	err := fs.WalkDir(embedfiles.EmbeddedTemplateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		fileName := filepath.Base(path)
		if fileName == "docker-compose-template.yaml" {
			// Skip copying the main template generator (docker-compose-template.yaml) to avoid confusion
			return nil
		}

		// Read the content of each example template file
		content, err := embedfiles.EmbeddedTemplateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded template %s: %v", path, err)
		}

		// Write the content to the target folder with the same file name
		targetPath := filepath.Join(folderPath, fileName)
		err = ioutil.WriteFile(targetPath, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write template to %s: %v", targetPath, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error initializing templates: %v", err)
	}

	fmt.Println("Example templates initialized in folder:", folderPath)
	return nil
}

func PopulateComposeWithTemplate(composeFile *dockercompose.ComposeFile, folderPath, templateName string, count int, serviceNames map[int]string) error {
	// Ensure template file has .yaml extension
	if !strings.HasSuffix(templateName, ".yaml") {
		templateName += ".yaml"
	}

	// Path to the user-modified template (e.g., "templates/example-service.yaml")
	templatePath := filepath.Join(folderPath, templateName)

	// Load the modified template content from the specified folder
	tmplContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %v", templateName, err)
	}

	// Parse the template content
	tmpl, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Loop to create service instances from the user-modified template
	for i := 1; i <= count; i++ {
		// Determine the service name
		var serviceName string
		if name, ok := serviceNames[i]; ok {
			serviceName = name
		} else if len(serviceNames) == 1 {
			for _, baseName := range serviceNames {
				serviceName = fmt.Sprintf("%s-%d", baseName, i)
			}
		} else {
			serviceName = fmt.Sprintf("%s-%d", templateName, i)
		}

		// Create a new ServiceInput based on the template
		service := dockercompose.ServiceInput{
			ServiceName:          serviceName,
			BuildContext:         "./path/to/context",
			ContainerName:        serviceName + "_container",
			ContainerIP:          "10.5.0.5",
			Command:              "echo Hello World",
			EnvironmentVariables: map[string]string{"ENV_VAR": "value"},
			HealthCheck: dockercompose.HealthCheck{
				Test:     []string{"CMD-SHELL", "echo 'healthy'"},
				Interval: "30s",
				Timeout:  "10s",
				Retries:  3,
			},
			Resources: dockercompose.ResourceLimits{
				MemoryLimit:       "512m",
				CPULimit:          "0.5",
				MemoryReservation: "256m",
			},
			Logging: dockercompose.LoggingConfig{
				Driver:  "json-file",
				MaxSize: "10m",
				MaxFile: "3",
			},
			Deploy: dockercompose.RestartPolicy{
				RestartCondition: "on-failure",
				MaxAttempts:      3,
			},
			Volumes: []string{"/data"},
		}

		// Render the modified template into the service configuration
		var renderedService strings.Builder
		err = tmpl.Execute(&renderedService, service)
		if err != nil {
			return fmt.Errorf("failed to apply template for service %s: %v", serviceName, err)
		}

		// Append the rendered service to the ComposeFile's Services
		composeFile.Services = append(composeFile.Services, service)
	}

	return nil
}
