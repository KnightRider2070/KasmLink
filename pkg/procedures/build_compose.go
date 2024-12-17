package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// findDockerfileForService searches for a Dockerfile in the ./dockerfiles/ directory that contains the serviceName.
func findDockerfileForService(serviceName string) (string, error) {
	log.Debug().
		Str("service_name", serviceName).
		Msg("Searching for Dockerfile matching service name")

	dockerfilesDir := "./dockerfiles"
	pattern := fmt.Sprintf("*%s*", serviceName)

	matchedFiles, err := filepath.Glob(filepath.Join(dockerfilesDir, pattern))
	if err != nil {
		log.Error().
			Err(err).
			Str("pattern", pattern).
			Msg("Failed to glob Dockerfiles")
		return "", fmt.Errorf("failed to glob dockerfiles: %w", err)
	}

	if len(matchedFiles) == 0 {
		log.Error().
			Str("service_name", serviceName).
			Str("directory", dockerfilesDir).
			Msg("No Dockerfile found containing service name")
		return "", fmt.Errorf("no Dockerfile found containing service name '%s' in %s", serviceName, dockerfilesDir)
	}

	if len(matchedFiles) > 1 {
		log.Error().
			Str("service_name", serviceName).
			Strs("matched_files", matchedFiles).
			Msg("Multiple Dockerfiles found for service name")
		return "", fmt.Errorf("multiple Dockerfiles found for service '%s' in %s: %v", serviceName, dockerfilesDir, matchedFiles)
	}

	log.Debug().
		Str("dockerfile", matchedFiles[0]).
		Msg("Found matching Dockerfile")

	return matchedFiles[0], nil
}

// sanitizeImageName sanitizes the image name to create valid filenames.
func sanitizeImageName(imageName string) string {
	// Replace '/' and ':' with underscores to prevent directory traversal or invalid filenames.
	return strings.ReplaceAll(strings.ReplaceAll(imageName, "/", "_"), ":", "_")
}

// checkLocalImageTarExists checks if the image tar file exists locally.
func checkLocalImageTarExists(imageName string) (bool, string) {
	sanitizedImageName := sanitizeImageName(imageName)
	localTarPath := filepath.Join("./tarfiles", fmt.Sprintf("%s.tar", sanitizedImageName))

	_, err := os.Stat(localTarPath)
	if err == nil {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar exists locally")
		return true, localTarPath
	}
	if os.IsNotExist(err) {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar does not exist locally")
		return false, localTarPath
	}
	// For other errors, log and treat as non-existent
	log.Error().
		Err(err).
		Str("local_tar_path", localTarPath).
		Msg("Error checking local tar file existence")
	return false, localTarPath
}
