package dockercli

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// CreateTarWithContext creates a tar archive of the specified directory.
// It returns a reader for the tar archive or an error if the operation fails.
func CreateTarWithContext(buildContextDir string) (io.Reader, error) {
	if buildContextDir == "" {
		return nil, errors.New("build context directory cannot be empty")
	}

	buffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buffer)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close tar writer")
		}
	}()

	log.Info().Str("directory", buildContextDir).Msg("Creating tar archive for build context")

	err := filepath.Walk(buildContextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing file during tar creation")
			return err
		}

		// Create tar header for the file or directory.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("error creating tar header for file %s: %w", path, err)
		}

		// Set the header name to the relative path in the archive.
		header.Name, err = filepath.Rel(buildContextDir, path)
		if err != nil {
			return fmt.Errorf("error determining relative path for file %s: %w", path, err)
		}

		// Write the header to the tar writer.
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header for file %s: %w", path, err)
		}

		// If the path is a directory, skip writing content.
		if info.IsDir() {
			return nil
		}

		// Add the file content to the tar archive.
		return writeFileToTar(tarWriter, path)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create tar archive: %w", err)
	}

	log.Info().Msg("Tar archive created successfully")
	return io.NopCloser(buffer), nil
}

// writeFileToTar writes the contents of a file to the tar writer.
// It ensures the file is opened and closed properly.
func writeFileToTar(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("file", path).Msg("Failed to open file for tar")
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error().Err(cerr).Str("file", path).Msg("Failed to close file after tar write")
		}
	}()

	// Copy the file content to the tar writer.
	_, err = io.Copy(tw, file)
	if err != nil {
		log.Error().Err(err).Str("file", path).Msg("Failed to copy file contents to tar")
		return fmt.Errorf("failed to copy file contents for %s: %w", path, err)
	}

	log.Debug().Str("file", path).Msg("File added to tar archive successfully")
	return nil
}
