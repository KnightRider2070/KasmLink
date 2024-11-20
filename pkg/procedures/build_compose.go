package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/fs"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"strings"
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

func MergeComposeFiles(file1, file2 dockercompose.ComposeFile) (dockercompose.ComposeFile, error) {
	// Check if versions are compatible
	if file1.Version != "" && file2.Version != "" && file1.Version != file2.Version {
		log.Error().
			Str("file1_version", file1.Version).
			Str("file2_version", file2.Version).
			Msg("incompatible compose file versions")
		return dockercompose.ComposeFile{}, fmt.Errorf("incompatible compose file versions: %s and %s", file1.Version, file2.Version)
	}

	// Use file2's version if file1's version is empty
	version := file1.Version
	if version == "" {
		version = file2.Version
	}

	// Initialize the merged ComposeFile
	merged := dockercompose.ComposeFile{
		Version:  version,
		Services: make(map[string]dockercompose.Service),
		Networks: make(map[string]dockercompose.Network),
		Volumes:  make(map[string]dockercompose.Volume),
		Configs:  make(map[string]dockercompose.Config),
		Secrets:  make(map[string]dockercompose.Secret),
	}

	// Merge services
	log.Debug().
		Interface("file1_services", file1.Services).
		Interface("file2_services", file2.Services).
		Msg("merging services")
	for name, service := range file1.Services {
		merged.Services[name] = service
	}
	for name, service := range file2.Services {
		if existingService, exists := merged.Services[name]; exists {
			log.Debug().Str("service_name", name).Msg("merging existing service")
			merged.Services[name] = existingService
		} else {
			merged.Services[name] = service
		}
	}

	// Merge networks
	log.Debug().
		Interface("file1_networks", file1.Networks).
		Interface("file2_networks", file2.Networks).
		Msg("merging networks")
	for name, network := range file1.Networks {
		merged.Networks[name] = network
	}
	for name, network := range file2.Networks {
		if _, exists := merged.Networks[name]; !exists {
			merged.Networks[name] = network
		}
	}

	// Merge volumes
	log.Debug().
		Interface("file1_volumes", file1.Volumes).
		Interface("file2_volumes", file2.Volumes).
		Msg("merging volumes")
	for name, volume := range file1.Volumes {
		merged.Volumes[name] = volume
	}
	for name, volume := range file2.Volumes {
		if _, exists := merged.Volumes[name]; !exists {
			merged.Volumes[name] = volume
		}
	}

	// Merge configs
	for name, config := range file1.Configs {
		merged.Configs[name] = config
	}
	for name, config := range file2.Configs {
		if _, exists := merged.Configs[name]; !exists {
			merged.Configs[name] = config
		}
	}

	// Merge secrets
	for name, secret := range file1.Secrets {
		merged.Secrets[name] = secret
	}
	for name, secret := range file2.Secrets {
		if _, exists := merged.Secrets[name]; !exists {
			merged.Secrets[name] = secret
		}
	}

	log.Debug().Interface("merged_compose_file", merged).Msg("merge completed")
	return merged, nil
}

// CreateServiceReplicas creates replicas of a service with modified names in a Compose file.
func CreateServiceReplicas(composeFile *dockercompose.ComposeFile, replicas int, inputNames []string) error {
	// Ensure there is exactly one service in the compose file
	if len(composeFile.Services) != 1 {
		return fmt.Errorf("expected exactly one service in the compose file, found %d", len(composeFile.Services))
	}

	// Retrieve the single service
	var originalServiceName string
	var originalService dockercompose.Service
	for name, service := range composeFile.Services {
		originalServiceName = name
		originalService = service
		break
	}

	// Validate input names
	if len(inputNames) == 0 {
		return fmt.Errorf("no input names provided")
	}

	// Generate names for replicas
	replicaNames := make([]string, replicas)
	if len(inputNames) == 1 {
		// If only one name is provided, generate names using name-i format
		baseName := inputNames[0]
		for i := 0; i < replicas; i++ {
			replicaNames[i] = fmt.Sprintf("%s-%d", baseName, i+1)
		}
	} else if len(inputNames) == replicas {
		// If enough names are provided, use them directly
		copy(replicaNames, inputNames)
	} else {
		return fmt.Errorf("number of input names (%d) does not match the number of replicas (%d)", len(inputNames), replicas)
	}

	// Create replicas in the compose file
	for _, replicaName := range replicaNames {
		newService := originalService
		newService.ContainerName = replicaName // Update the container name
		composeFile.Services[replicaName] = newService
	}

	// Remove the original service
	delete(composeFile.Services, originalServiceName)

	return nil
}

// WriteComposeFile writes the provided ComposeFile object to a specified file path.
func WriteComposeFile(composeFile *dockercompose.ComposeFile, filePath string) error {
	// Open the file for writing (create or truncate)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Encode the ComposeFile as YAML
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(composeFile); err != nil {
		return fmt.Errorf("failed to write compose file to %s: %w", filePath, err)
	}

	return nil
}
