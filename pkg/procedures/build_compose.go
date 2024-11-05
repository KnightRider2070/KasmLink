package procedures

import (
	_ "embed" // Required for embedding files
	"fmt"
	"github.com/rs/zerolog/log"
	"io/fs"
	"io/ioutil"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// InitTemplatesFolder initializes the templates folder by copying embedded default templates into a "services" subfolder.
func InitTemplatesFolder(folderPath string) error {
	servicesFolderPath := filepath.Join(folderPath, "services")
	log.Info().Str("folderPath", folderPath).Msg("Initialized folder path")
	log.Info().Str("servicesFolderPath", servicesFolderPath).Msg("Service folder path for templates")

	// Attempt to create the services folder
	if _, err := os.Stat(servicesFolderPath); os.IsNotExist(err) {
		log.Debug().Str("path", servicesFolderPath).Msg("Creating services folder")
		err = os.MkdirAll(servicesFolderPath, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Str("path", servicesFolderPath).Msg("Failed to create services folder")
			return fmt.Errorf("failed to create services folder: %v", err)
		}
	}

	err := fs.WalkDir(embedfiles.EmbeddedTemplateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking through embedded templates directory")
			return err
		}
		if d.IsDir() {
			log.Debug().Str("directory", path).Msg("Skipping directory")
			return nil
		}

		// Compute the relative path and the target path
		relativePath := strings.TrimPrefix(path, "templates/")
		targetPath := filepath.Join(servicesFolderPath, relativePath)
		targetDir := filepath.Dir(targetPath)

		// Log detailed debug information about path processing
		log.Debug().
			Str("source", path).
			Str("relativePath", relativePath).
			Str("targetPath", targetPath).
			Msg("Processing file for template initialization")

		// Create the target directory if it doesn’t exist
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			log.Debug().Str("path", targetDir).Msg("Creating target directory for file")
			if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
				log.Error().Err(err).Str("path", targetDir).Msg("Failed to create target directory")
				return fmt.Errorf("failed to create directory %s: %v", targetDir, err)
			}
		}

		// Read and write the template file content
		content, err := embedfiles.EmbeddedTemplateFS.ReadFile(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to read embedded template file")
			return fmt.Errorf("failed to read embedded template %s: %v", path, err)
		}

		err = ioutil.WriteFile(targetPath, content, 0644)
		if err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to write template file to target path")
			return fmt.Errorf("failed to write template to %s: %v", targetPath, err)
		}
		log.Info().Str("file", targetPath).Msg("Template file initialized successfully")
		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Error during template folder initialization")
		return fmt.Errorf("error initializing templates: %v", err)
	}

	log.Info().Str("servicesFolderPath", servicesFolderPath).Msg("Template initialization completed")
	return nil
}

// InitDockerfilesFolder initializes the Dockerfiles folder by copying embedded Dockerfile templates.
func InitDockerfilesFolder(folderPath string) error {
	dockerfilesFolderPath := filepath.Join(folderPath, "dockerfiles")
	log.Info().Str("folderPath", folderPath).Msg("Initialized dockerfiles folder path")
	log.Info().Str("dockerfilesFolderPath", dockerfilesFolderPath).Msg("Dockerfiles folder path for templates")

	// Attempt to create the dockerfiles folder
	if _, err := os.Stat(dockerfilesFolderPath); os.IsNotExist(err) {
		log.Debug().Str("path", dockerfilesFolderPath).Msg("Creating dockerfiles folder")
		err = os.MkdirAll(dockerfilesFolderPath, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Str("path", dockerfilesFolderPath).Msg("Failed to create dockerfiles folder")
			return fmt.Errorf("failed to create dockerfiles folder: %v", err)
		}
	}

	err := fs.WalkDir(embedfiles.EmbeddedDockerImagesDirectory, "dockerfiles", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking through embedded dockerfiles directory")
			return err
		}
		if d.IsDir() {
			log.Debug().Str("directory", path).Msg("Skipping directory")
			return nil
		}

		// Compute the relative path and the target path
		relativePath := strings.TrimPrefix(path, "dockerfiles/")
		targetPath := filepath.Join(dockerfilesFolderPath, relativePath)
		targetDir := filepath.Dir(targetPath)

		// Log detailed debug information about path processing
		log.Debug().
			Str("source", path).
			Str("relativePath", relativePath).
			Str("targetPath", targetPath).
			Msg("Processing file for dockerfile initialization")

		// Create the target directory if it doesn’t exist
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			log.Debug().Str("path", targetDir).Msg("Creating target directory for file")
			if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
				log.Error().Err(err).Str("path", targetDir).Msg("Failed to create target directory")
				return fmt.Errorf("failed to create directory %s: %v", targetDir, err)
			}
		}

		// Read and write the Dockerfile template content
		content, err := embedfiles.EmbeddedDockerImagesDirectory.ReadFile(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to read embedded dockerfile template")
			return fmt.Errorf("failed to read embedded Dockerfile template %s: %v", path, err)
		}

		err = ioutil.WriteFile(targetPath, content, 0644)
		if err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to write dockerfile to target path")
			return fmt.Errorf("failed to write Dockerfile to %s: %v", targetPath, err)
		}
		log.Info().Str("file", targetPath).Msg("Dockerfile template initialized successfully")
		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Error during dockerfile folder initialization")
		return fmt.Errorf("error initializing Dockerfiles: %v", err)
	}

	log.Info().Str("dockerfilesFolderPath", dockerfilesFolderPath).Msg("Dockerfile initialization completed")
	return nil
}

// PopulateComposeWithTemplate populates a docker-compose file with a user-specified template, service name, and count.
func PopulateComposeWithTemplate(composeFile *dockercompose.ComposeFile, folderPath, templateName string, count int, serviceNames map[int]string) error {
	// Ensure the template file has a .yaml extension
	if !strings.HasSuffix(templateName, ".yaml") {
		templateName += ".yaml"
	}
	log.Info().Str("templateName", templateName).Int("count", count).Msg("Initializing template population")

	// Path to the user-modified template (e.g., "templates/example-service.yaml")
	templatePath := filepath.Join(folderPath, templateName)
	log.Debug().Str("templatePath", templatePath).Msg("Computed template path")

	// Load the modified template content from the specified folder
	tmplContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to load template content")
		return fmt.Errorf("failed to load template %s: %v", templateName, err)
	}
	log.Info().Str("templatePath", templatePath).Msg("Template loaded successfully")

	// Parse the template content
	tmpl, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		log.Error().Err(err).Str("templateName", templateName).Msg("Failed to parse template content")
		return fmt.Errorf("failed to parse template: %v", err)
	}
	log.Info().Str("templateName", templateName).Msg("Template parsed successfully")

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
			serviceName = fmt.Sprintf("%s-%d", strings.TrimSuffix(templateName, ".yaml"), i)
		}
		log.Debug().Str("serviceName", serviceName).Int("instance", i).Msg("Generated service name")

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
		log.Info().Str("serviceName", serviceName).Msg("Service input structure created")

		// Render the modified template into the service configuration
		var renderedService strings.Builder
		err = tmpl.Execute(&renderedService, service)
		if err != nil {
			log.Error().Err(err).Str("serviceName", serviceName).Msg("Failed to apply template to service configuration")
			return fmt.Errorf("failed to apply template for service %s: %v", serviceName, err)
		}
		log.Info().Str("serviceName", serviceName).Msg("Template successfully rendered for service")

		// Append the rendered service to the ComposeFile's Services
		composeFile.Services = append(composeFile.Services, service)
		log.Debug().Str("serviceName", serviceName).Msg("Service appended to ComposeFile")
	}

	log.Info().Str("templateName", templateName).Int("count", count).Msg("Template population completed successfully")
	return nil
}
