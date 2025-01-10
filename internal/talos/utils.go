package talos

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	embedfiles "kasmlink/embedded"
	programmConfig "kasmlink/internal/config"
	"os"
	"path/filepath"
	"strings"
)

// copyFilesFromClusterDirectory copies files from the embedded "cluster" directory to the "ClusterDirectory"
func copyFilesFromClusterDirectory(embedDir, targetDir string) error {
	files, err := embedfiles.ClusterDirectory.ReadDir(embedDir)
	if err != nil {
		return fmt.Errorf("failed to read embedded directory %s: %w", embedDir, err)
	}

	for _, file := range files {
		sourcePath := embedDir
		if embedDir != "" {
			sourcePath += "/" // Ensure forward slash for embedded paths
		}
		sourcePath += file.Name()                                // Path in the embedded FS
		destinationPath := filepath.Join(targetDir, file.Name()) // Path in the real FS

		log.Info().Msgf("Processing: Source: %s, Destination: %s", sourcePath, destinationPath)

		if file.IsDir() {
			// Ensure the destination directory exists
			if err := os.MkdirAll(destinationPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destinationPath, err)
			}

			// Recursively copy contents
			if err := copyFilesFromClusterDirectory(sourcePath, destinationPath); err != nil {
				return err
			}
		} else {
			// Read file content from the embedded FS
			content, err := embedfiles.ClusterDirectory.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to read embedded file %s: %w", sourcePath, err)
			}

			// Write file content to the target FS
			if err := os.WriteFile(destinationPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destinationPath, err)
			}

			log.Info().Msgf("File copied: %s to %s", sourcePath, destinationPath)
		}
	}

	return nil
}

// ReplaceInFile replaces a placeholder string in a file with the given value.
func ReplaceInFile(filePath, placeholder, value string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, placeholder) {
		return fmt.Errorf("placeholder %s not found in the file %s", placeholder, filePath)
	}

	updatedContent := strings.Replace(contentStr, placeholder, value, -1)

	if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated content to file %s: %w", filePath, err)
	}

	log.Info().Msgf("Successfully replaced placeholder %s with Age public key in %s", placeholder, filePath)
	return nil
}

// logAndReturnError logs the error with a message and returns the error for further handling.
func logAndReturnError(err error, msg string) error {
	log.Error().Err(err).Msg(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

// addIndentationToLines adds indentation to each line in the input content.
func addIndentationToLines(content []byte) []string {
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		lines[i] = "  " + line
	}
	return lines
}

// ensureDirExists ensures that the directory exists, creating it if necessary.
func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// CopyFile copies a file from source to destination. If replaceExisting is true, it will overwrite existing files in the destination.
func CopyFile(source, destination string, replaceExisting bool) error {
	// If replaceExisting is false, check if the destination file exists and skip if it does
	if !replaceExisting {
		if _, err := os.Stat(destination); err == nil {
			log.Info().Msgf("Skipping existing file: %s", destination)
			return nil
		} else if !os.IsNotExist(err) {
			// If there was another error (not just that the file doesn't exist), log and return
			log.Error().Err(err).Msgf("Error checking if file exists: %s", destination)
			return fmt.Errorf("failed to check if destination file exists: %w", err)
		}
	}

	// Open the source file
	sourceFile, err := os.Open(source)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open source file: %s", source)
		return fmt.Errorf("failed to open source file %s: %w", source, err)
	}
	defer sourceFile.Close()

	// Check if the source file is empty
	info, err := sourceFile.Stat()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get source file stats: %s", source)
		return fmt.Errorf("failed to get source file stats: %w", err)
	}

	// If the file is empty, log a warning but continue the operation
	if info.Size() == 0 {
		log.Warn().Msgf("Source file is empty: %s", source)
	}

	// Create the destination file (or truncate it if it already exists)
	destinationFile, err := os.Create(destination)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create destination file: %s", destination)
		return fmt.Errorf("failed to create destination file %s: %w", destination, err)
	}
	defer destinationFile.Close()

	// Copy contents from source file to destination file
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to copy data from source to destination")
		return fmt.Errorf("failed to copy data from %s to %s: %w", source, destination, err)
	}

	// Log the successful copy operation
	log.Info().Msgf("File copied from %s to %s", source, destination)
	return nil
}

// getKnownHostsEntry generates a known_hosts entry for the specified Git URL.
func getKnownHostsEntry(gitURL string) string {
	// For simplicity, assuming the Git URL is for GitHub, returning the known hosts entry for GitHub
	return fmt.Sprintf("github.com,192.30.255.112,192.30.255.113 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICN4F29hWlP/2V8A6O49CHBYbQw87vFFYYg1I7VJpgW2X")
}

// EnsureCacheDirsExist ensures the necessary cache directories exist.
func EnsureCacheDirsExist() error {
	dirs := []string{
		programmConfig.CacheDir,
		programmConfig.CachePaths.HelmChartCache,
		programmConfig.CachePaths.KubernetesCache,
		programmConfig.CachePaths.BaseConfigCache,
		programmConfig.CachePaths.TemporaryCache,
		programmConfig.CachePaths.PatchCache,
		programmConfig.CachePaths.DocumentationCache,
	}

	// Ensure each directory exists
	for _, dir := range dirs {
		if err := createDirIfNotExist(dir); err != nil {
			return err
		}
	}
	return nil
}

// EnsureClusterDirsExist ensures the necessary cluster-related directories exist.
func EnsureClusterDirsExist() error {
	dirs := []string{
		programmConfig.ClusterDirectory,
		programmConfig.ConfigPaths.TalosDir,
		programmConfig.ConfigPaths.KubernetesDir,
		programmConfig.ConfigPaths.GeneratedDir,
	}

	// Ensure each directory exists
	for _, dir := range dirs {
		if err := createDirIfNotExist(dir); err != nil {
			return err
		}
	}
	return nil
}

// createDirIfNotExist ensures that a given directory exists, if not, it creates it.
func createDirIfNotExist(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", directory, err)
		}
	}
	return nil
}
