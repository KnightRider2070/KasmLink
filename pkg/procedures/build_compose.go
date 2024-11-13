package procedures

import (
	"fmt"
	"io/fs"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
)

// InitFolder initializes a specified folder by copying embedded templates or Dockerfiles.
func InitFolder(folderPath, subfolder, sourcePath string, embeddedFS fs.FS) error {
	targetFolder := filepath.Join(folderPath, subfolder)
	log.Info().Str("folderPath", folderPath).Str("subfolder", subfolder).Msg("Initializing folder path")

	// Create the target folder if it doesn’t exist
	if err := os.MkdirAll(targetFolder, os.ModePerm); err != nil {
		log.Error().Err(err).Str("path", targetFolder).Msg("Failed to create target folder")
		return fmt.Errorf("failed to create folder %s: %w", targetFolder, err)
	}

	// Copy files from embedded FS to target folder
	if err := copyEmbeddedFiles(embeddedFS, sourcePath, targetFolder); err != nil {
		log.Error().Err(err).Str("subfolder", subfolder).Msg("Error during folder initialization")
		return fmt.Errorf("error initializing folder %s: %w", subfolder, err)
	}

	log.Info().Str("folderPath", targetFolder).Msg("Folder initialization completed successfully")
	return nil
}

// copyEmbeddedFiles copies files from an embedded file system to a target directory.
func copyEmbeddedFiles(embeddedFS fs.FS, sourcePath, targetFolder string) error {
	return fs.WalkDir(embeddedFS, sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking through embedded directory")
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		// Prepare target file path and create necessary directories
		relativePath := strings.TrimPrefix(path, sourcePath+"/")
		targetPath := filepath.Join(targetFolder, relativePath)
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to create directory for file")
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Copy file content from embedded FS to target path
		content, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to read embedded file")
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to write file to target path")
			return fmt.Errorf("failed to write file to %s: %w", targetPath, err)
		}
		log.Info().Str("file", targetPath).Msg("File initialized successfully")
		return nil
	})
}

// InitTemplatesFolder initializes the templates folder with embedded templates.
func InitTemplatesFolder(folderPath string) error {
	return InitFolder(folderPath, "services", "services", embedfiles.EmbeddedServicesFS)
}

// InitDockerfilesFolder initializes the Dockerfiles folder with embedded Dockerfile templates.
func InitDockerfilesFolder(folderPath string) error {
	return InitFolder(folderPath, "dockerfiles", "dockerfiles", embedfiles.EmbeddedDockerImagesDirectory)
}

// PopulateComposeWithTemplate populates a Docker Compose file with instances of a specified template.
func PopulateComposeWithTemplate(composeFile *dockercompose.ComposeFile, folderPath, templateName string, count int, serviceNames map[int]string) error {
	if !strings.HasSuffix(templateName, ".yaml") {
		templateName += ".yaml"
	}
	templatePath := filepath.Join(folderPath, "services", templateName)
	log.Info().Str("templateName", templateName).Int("count", count).Msg("Starting template population")

	// Load and parse the template
	tmpl, err := loadTemplate(templatePath, templateName)
	if err != nil {
		return fmt.Errorf("error loading template: %w", err)
	}

	// Ensure networks and volumes are initialized in the Compose file
	ensureNetworksAndVolumes(composeFile)

	// Loop through each instance to create services
	for i := 1; i <= count; i++ {
		serviceName := generateServiceName(serviceNames, i, templateName)
		log.Debug().Str("serviceName", serviceName).Msg("Creating service instance")

		// Define and render the service structure
		service := createServiceConfig(serviceName)
		if err := renderServiceToCompose(tmpl, service, serviceName, composeFile); err != nil {
			log.Error().Err(err).Str("serviceName", serviceName).Msg("Failed to render service to compose file")
			return fmt.Errorf("failed to render service %s to compose file: %w", serviceName, err)
		}
	}

	log.Info().Str("templateName", templateName).Int("count", count).Msg("Template population completed successfully")
	return nil
}

// ensureNetworksAndVolumes adds default networks and volumes if missing.
func ensureNetworksAndVolumes(composeFile *dockercompose.ComposeFile) {
	if composeFile.Networks == nil {
		composeFile.Networks = make(map[string]dockercompose.Network)
	}
	if composeFile.Volumes == nil {
		composeFile.Volumes = make(map[string]dockercompose.Volume)
	}

	// Add example network
	if _, exists := composeFile.Networks["example_network"]; !exists {
		composeFile.Networks["example_network"] = dockercompose.Network{
			Driver: "bridge",
			DriverOpts: map[string]string{
				"subnet": "10.5.0.0/16",
			},
		}
		log.Info().Str("network", "example_network").Msg("Default network added to ComposeFile")
	}

	// Add example volume
	if _, exists := composeFile.Volumes["nfs_data"]; !exists {
		composeFile.Volumes["nfs_data"] = dockercompose.Volume{
			Driver: "local",
			DriverOpts: map[string]string{
				"type":   "none",
				"device": "/path/to/host/directory",
				"o":      "bind",
			},
		}
		log.Info().Str("volume", "nfs_data").Msg("Default volume added to ComposeFile")
	}
}

// loadTemplate loads and parses the specified template file.
func loadTemplate(templatePath, templateName string) (*template.Template, error) {
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Error().Err(err).Str("templatePath", templatePath).Msg("Failed to load template content")
		return nil, fmt.Errorf("failed to load template %s: %w", templateName, err)
	}
	tmpl, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		log.Error().Err(err).Str("templateName", templateName).Msg("Failed to parse template")
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}
	log.Info().Str("templateName", templateName).Msg("Template loaded and parsed successfully")
	return tmpl, nil
}

// generateServiceName generates a service name based on the provided name or template name.
func generateServiceName(serviceNames map[int]string, index int, templateName string) string {
	if name, exists := serviceNames[index]; exists && name != "" {
		return name
	}
	return fmt.Sprintf("%s-%d", strings.TrimSuffix(templateName, ".yaml"), index)
}

// createServiceConfig defines a new service configuration.
func createServiceConfig(serviceName string) dockercompose.Service {
	log.Debug().Str("serviceName", serviceName).Msg("Creating service configuration")
	return dockercompose.Service{
		ContainerName: fmt.Sprintf("%s_container", serviceName),
		Build: &dockercompose.BuildConfig{
			Context:    "./path/to/context",
			Dockerfile: "Dockerfile",
			Args: map[string]string{
				"ARG1": "value1",
				"ARG2": "value2",
			},
		},
		Healthcheck: &dockercompose.Healthcheck{
			Test:     []string{"CMD-SHELL", "echo 'healthy'"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
		Logging: &dockercompose.Logging{
			Driver: "json-file",
			Options: map[string]string{
				"max-size": "10m",
				"max-file": "3",
			},
		},
		RestartPolicy: "on-failure",
		Volumes:       []string{"nfs_data:/data"},
		NetworkConfig: dockercompose.NetworkConfig{
			Networks: []string{"example_network"},
		},
		Tty:     false,
		Command: []string{"echo", "Hello World"},
	}
}

// renderServiceToCompose renders the service template and adds it to the Compose file.
func renderServiceToCompose(tmpl *template.Template, service dockercompose.Service, serviceName string, composeFile *dockercompose.ComposeFile) error {
	var renderedService strings.Builder
	if err := tmpl.Execute(&renderedService, service); err != nil {
		log.Error().Err(err).Str("serviceName", serviceName).Msg("Failed to apply template")
		return fmt.Errorf("failed to apply template for service %s: %w", serviceName, err)
	}

	// Merge volumes and networks if already present
	existingService, exists := composeFile.Services[serviceName]
	if exists {
		// Merge volumes
		service.Volumes = mergeStringSlices(existingService.Volumes, service.Volumes)
		// Merge networks
		service.NetworkConfig.Networks = mergeStringSlices(existingService.NetworkConfig.Networks, service.NetworkConfig.Networks)
	}

	composeFile.Services[serviceName] = service
	log.Info().Str("serviceName", serviceName).Msg("Service added to ComposeFile successfully")
	return nil
}

// mergeStringSlices merges two slices of strings, avoiding duplicates.
func mergeStringSlices(slice1, slice2 []string) []string {
	set := make(map[string]struct{})
	for _, v := range slice1 {
		set[v] = struct{}{}
	}
	for _, v := range slice2 {
		set[v] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	return result
}
