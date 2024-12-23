// procedures/procedures.go
package internal

import (
	"fmt"
	"io/fs"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Constants for folder initialization
const (
	DefaultFilePermission   = 0644
	DefaultFolderPermission = 0755
)

// Precompiled regular expressions for performance
var (
	urlRegex           = regexp.MustCompile(`https?://[^\s]+`)
	variablePatternStr = `(?m)^\s*%s\s*=\s*['"]?([^'"\s]+)['"]?`
)

// Mutex to ensure thread-safe operations if needed in future
var mu sync.RWMutex

// InitFolder initializes a specified folder by copying embedded templates or Dockerfiles.
// Parameters:
// - folderPath: The base directory where the subfolder will be created.
// - subfolder: The name of the subfolder to initialize.
// - sourcePath: The path within the embedded filesystem to copy files from.
// - embeddedFS: The embedded filesystem containing the source files.
// Returns:
// - An error if the initialization fails.
func InitFolder(folderPath, subfolder, sourcePath string, embeddedFS fs.FS) error {
	targetFolder := filepath.Join(folderPath, subfolder)
	log.Info().
		Str("folderPath", folderPath).
		Str("subfolder", subfolder).
		Msg("Initializing folder path")

	// Create the target folder if it doesn’t exist
	if err := os.MkdirAll(targetFolder, DefaultFolderPermission); err != nil {
		log.Error().
			Err(err).
			Str("path", targetFolder).
			Msg("Failed to create target folder")
		return fmt.Errorf("failed to create folder %s: %w", targetFolder, err)
	}

	// Copy files from embedded FS to target folder
	if err := copyEmbeddedFiles(embeddedFS, sourcePath, targetFolder); err != nil {
		log.Error().
			Err(err).
			Str("subfolder", subfolder).
			Msg("Error during folder initialization")
		return fmt.Errorf("error initializing folder %s: %w", subfolder, err)
	}

	log.Info().
		Str("folderPath", targetFolder).
		Msg("Folder initialization completed successfully")
	return nil
}

// copyEmbeddedFiles copies files from an embedded file system to a target directory.
// Parameters:
// - embeddedFS: The embedded filesystem to copy files from.
// - sourcePath: The source directory within the embedded filesystem.
// - targetFolder: The target directory on the local filesystem.
// Returns:
// - An error if the copying process fails.
func copyEmbeddedFiles(embeddedFS fs.FS, sourcePath, targetFolder string) error {
	return fs.WalkDir(embeddedFS, sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Error walking through embedded directory")
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		// Prepare target file path and create necessary directories
		relativePath := strings.TrimPrefix(path, sourcePath+"/")
		targetPath := filepath.Join(targetFolder, relativePath)
		if err := os.MkdirAll(filepath.Dir(targetPath), DefaultFolderPermission); err != nil {
			log.Error().
				Err(err).
				Str("path", targetPath).
				Msg("Failed to create directory for file")
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Copy file content from embedded FS to target path
		content, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Failed to read embedded file")
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}
		if err := os.WriteFile(targetPath, content, DefaultFilePermission); err != nil {
			log.Error().
				Err(err).
				Str("path", targetPath).
				Msg("Failed to write file to target path")
			return fmt.Errorf("failed to write file to %s: %w", targetPath, err)
		}
		log.Info().
			Str("file", targetPath).
			Msg("File initialized successfully")
		return nil
	})
}

// InitTemplatesFolder initializes the templates folder with embedded templates.
// Parameters:
// - folderPath: The base directory where the templates folder will be created.
// Returns:
// - An error if the initialization fails.
func InitTemplatesFolder(folderPath string) error {
	return InitFolder(folderPath, "templates", "templates", embedfiles.EmbeddedServicesFS)
}

// InitDockerfilesFolder initializes the Dockerfiles folder with embedded Dockerfile templates.
// Parameters:
// - folderPath: The base directory where the Dockerfiles folder will be created.
// Returns:
// - An error if the initialization fails.
func InitDockerfilesFolder(folderPath string) error {
	return InitFolder(folderPath, "dockerfiles", "dockerfiles", embedfiles.EmbeddedDockerImagesDirectory)
}

// MergeComposeFiles merges two Docker Compose files into one.
// It ensures version compatibility and merges services, networks, volumes, configs, and secrets.
// Parameters:
// - file1: The first ComposeFile to merge.
// - file2: The second ComposeFile to merge.
// Returns:
// - The merged ComposeFile.
// - An error if merging fails due to incompatibilities.
func MergeComposeFiles(file1, file2 dockercompose.ComposeFile) (dockercompose.ComposeFile, error) {
	// Check if versions are compatible
	if file1.Version != "" && file2.Version != "" && file1.Version != file2.Version {
		log.Error().
			Str("file1_version", file1.Version).
			Str("file2_version", file2.Version).
			Msg("Incompatible compose file versions")
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
		Msg("Merging services")
	for name, service := range file1.Services {
		merged.Services[name] = service
	}
	for name, service := range file2.Services {
		if existingService, exists := merged.Services[name]; exists {
			log.Debug().
				Str("service_name", name).
				Msg("Merging existing service")
			merged.Services[name] = mergeServices(existingService, service)
		} else {
			merged.Services[name] = service
		}
	}

	// Merge networks
	log.Debug().
		Interface("file1_networks", file1.Networks).
		Interface("file2_networks", file2.Networks).
		Msg("Merging networks")
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
		Msg("Merging volumes")
	for name, volume := range file1.Volumes {
		merged.Volumes[name] = volume
	}
	for name, volume := range file2.Volumes {
		if _, exists := merged.Volumes[name]; !exists {
			merged.Volumes[name] = volume
		}
	}

	// Merge configs
	log.Debug().
		Interface("file1_configs", file1.Configs).
		Interface("file2_configs", file2.Configs).
		Msg("Merging configs")
	for name, config := range file1.Configs {
		merged.Configs[name] = config
	}
	for name, config := range file2.Configs {
		if _, exists := merged.Configs[name]; !exists {
			merged.Configs[name] = config
		}
	}

	// Merge secrets
	log.Debug().
		Interface("file1_secrets", file1.Secrets).
		Interface("file2_secrets", file2.Secrets).
		Msg("Merging secrets")
	for name, secret := range file1.Secrets {
		merged.Secrets[name] = secret
	}
	for name, secret := range file2.Secrets {
		if _, exists := merged.Secrets[name]; !exists {
			merged.Secrets[name] = secret
		}
	}

	log.Debug().
		Interface("merged_compose_file", merged).
		Msg("Merge completed")
	return merged, nil
}

// mergeServices merges two Docker Compose services into one.
// It can be extended to handle more complex merging logic.
// Parameters:
// - service1: The first service to merge.
// - service2: The second service to merge.
// Returns:
// - The merged service.
func mergeServices(service1, service2 dockercompose.Service) dockercompose.Service {
	// Placeholder for merging logic. Currently, service2 overrides service1.
	// Extend this function to handle specific merging rules as needed.
	return service2
}

// CreateServiceReplicas creates replicas of a service with modified names in a Compose file.
// Parameters:
// - composeFile: The ComposeFile object to modify.
// - replicas: The number of replicas to create.
// - inputNames: The names to assign to the replicas. If a single name is provided, it appends an index.
// Returns:
// - An error if replica creation fails.
func CreateServiceReplicas(composeFile *dockercompose.ComposeFile, replicas int, inputNames []string) error {
	mu.Lock()
	defer mu.Unlock()

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
		log.Info().
			Str("replica_name", replicaName).
			Msg("Created service replica")
	}

	// Remove the original service
	delete(composeFile.Services, originalServiceName)
	log.Info().
		Str("original_service", originalServiceName).
		Msg("Removed original service after creating replicas")

	return nil
}

// WriteComposeFile writes the provided ComposeFile object to a specified file path.
// Parameters:
// - composeFile: The ComposeFile object to write.
// - filePath: The destination file path.
// Returns:
// - An error if writing fails.
func WriteComposeFile(composeFile *dockercompose.ComposeFile, filePath string) error {
	// Open the file for writing (create or truncate)
	file, err := os.Create(filePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("filePath", filePath).
			Msg("Failed to create compose file")
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Str("filePath", filePath).
				Msg("Failed to close compose file")
		}
	}()

	// Encode the ComposeFile as YAML
	encoder := yaml.NewEncoder(file)
	defer func() {
		if cerr := encoder.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("Failed to close YAML encoder")
		}
	}()

	if err := encoder.Encode(composeFile); err != nil {
		log.Error().
			Err(err).
			Str("filePath", filePath).
			Msg("Failed to encode compose file as YAML")
		return fmt.Errorf("failed to write compose file to %s: %w", filePath, err)
	}

	log.Info().
		Str("filePath", filePath).
		Msg("Compose file written successfully")
	return nil
}
