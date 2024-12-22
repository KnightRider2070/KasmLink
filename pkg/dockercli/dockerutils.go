// dockercli/dockercli.go
package dockercli

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
	"math/rand"
)

// Constants for tar archive creation and export limits.
const (
	tarBufferSize = 1000 * 1024 * 1024 // 1GB buffer size for tar creation
	maxTarSize    = 100 << 30          // 100 GB maximum tar size
)

// BuildLog represents the structure of Docker build log messages.
type BuildLog struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

// BuildDockerImage builds a Docker image from a specified build context directory and Dockerfile.
// It streams the build output to the PrintBuildLogs method for real-time logging.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - imageTag: The tag to assign to the built image (e.g., "myapp:latest").
// - dockerfilePath: Path to the Dockerfile within the build context directory.
// - buildContextPath: Path to the build context directory.
// - buildArgs: Optional build arguments to pass to the Docker build.
// Returns:
// - An error if the build process fails or is aborted.
func (dc *DockerClient) BuildDockerImage(ctx context.Context, imageTag, dockerfilePath, buildContextPath string, buildArgs map[string]*string) error {
	log.Info().
		Str("imageTag", imageTag).
		Str("dockerfilePath", dockerfilePath).
		Str("buildContextPath", buildContextPath).
		Msg("Starting Docker image build")

	// Validate Dockerfile existence within the build context directory
	dockerfileFullPath := filepath.Join(buildContextPath, dockerfilePath)
	if _, err := os.Stat(dockerfileFullPath); os.IsNotExist(err) {
		log.Error().
			Str("dockerfilePath", dockerfileFullPath).
			Msg("Dockerfile does not exist in build context")
		return fmt.Errorf("dockerfile does not exist at path %s in build context", dockerfileFullPath)
	} else if err != nil {
		log.Error().
			Err(err).
			Str("dockerfilePath", dockerfileFullPath).
			Msg("Error accessing Dockerfile in build context")
		return fmt.Errorf("error accessing Dockerfile in build context: %w", err)
	}

	// Create a tar archive from the build context directory
	tarReader, err := CreateTarFromDirectory(buildContextPath)
	if err != nil {
		log.Error().
			Err(err).
			Str("buildContextPath", buildContextPath).
			Msg("Failed to create tar archive from build context")
		return fmt.Errorf("failed to create tar archive from build context: %w", err)
	}

	// Prepare build options
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: filepath.Base(dockerfilePath),
		Remove:     true, // Remove intermediate containers after a successful build
		BuildArgs:  buildArgs,
		// Set other build options as needed
	}

	// Attempt to build the image with retry logic
	var imageBuildResponse types.ImageBuildResponse

	retryDelay := dc.initialRetryDelay

	for attempt := 1; attempt <= dc.retries; attempt++ {
		// Check if context is done before attempting
		select {
		case <-ctx.Done():
			log.Error().
				Err(ctx.Err()).
				Msg("BuildDockerImage aborted due to context cancellation before attempting")
			return fmt.Errorf("build image aborted due to context cancellation: %w", ctx.Err())
		default:
			// Continue
		}

		// Start the image build
		imageBuildResponse, err = dc.cli.ImageBuild(ctx, tarReader, buildOptions)
		if err != nil {
			log.Error().
				Err(err).
				Str("imageTag", imageTag).
				Int("attempt", attempt).
				Msg("Failed to initiate Docker image build")

			// Categorize the error
			if isPermanentError(err) {
				log.Error().
					Err(err).
					Str("imageTag", imageTag).
					Msg("Permanent error encountered during Docker image build. Not retrying.")
				return fmt.Errorf("permanent error during Docker image build: %w", err)
			}

			// If not the last attempt, wait before retrying
			if attempt < dc.retries {
				// Calculate delay with jitter
				jitter := time.Duration(float64(retryDelay) * dc.jitterFactor * (rand.Float64()*2 - 1)) // +/- jitterFactor * retryDelay
				sleepDuration := retryDelay + jitter
				if sleepDuration < 0 {
					sleepDuration = 0
				}

				log.Warn().
					Int("attempt", attempt).
					Dur("retry_delay", sleepDuration).
					Msg("Retrying BuildDockerImage after delay")

				// Wait for the calculated duration or until context is canceled
				select {
				case <-time.After(sleepDuration):
					// Continue to next attempt
				case <-ctx.Done():
					log.Error().
						Err(ctx.Err()).
						Msg("BuildDockerImage aborted during retry delay due to context cancellation")
					return fmt.Errorf("build image aborted during retry delay: %w", ctx.Err())
				}

				// Exponential backoff
				retryDelay *= time.Duration(dc.backoffMultiplier)
				if retryDelay > dc.maxRetryDelay {
					retryDelay = dc.maxRetryDelay
				}

				continue
			}
		} else {
			// Success
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to initiate Docker image build for %s after %d attempts: %w", imageTag, dc.retries, err)
	}

	defer func() {
		if cerr := imageBuildResponse.Body.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Msg("Failed to close image build response body")
		}
	}()

	// Process build logs
	if err := dc.PrintBuildLogs(ctx, imageBuildResponse.Body); err != nil {
		log.Error().
			Err(err).
			Str("imageTag", imageTag).
			Msg("Error occurred during Docker build logs processing")
		return fmt.Errorf("error occurred during Docker build logs processing: %w", err)
	}

	log.Info().
		Str("imageTag", imageTag).
		Msg("Docker image built successfully")
	return nil
}

// isPermanentError determines whether an error is permanent (should not be retried).
// This function can be extended to handle more specific error types.
// Parameters:
// - err: The error to categorize.
// Returns:
// - true if the error is permanent, false otherwise.
func isPermanentError(err error) bool {
	// Example: Check for permission denied errors
	if strings.Contains(err.Error(), "permission denied") {
		return true
	}
	// Add more conditions for permanent errors as needed
	return false
}

// PrintBuildLogs reads the Docker build output and formats it for better readability.
// It processes each log message, applying color formatting for success and error streams.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - reader: An io.Reader from which to read Docker build logs.
// Returns:
// - An error if log processing fails or is aborted.
func (dc *DockerClient) PrintBuildLogs(ctx context.Context, reader io.Reader) error {
	decoder := json.NewDecoder(reader)

	var logMsg BuildLog

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Error().
				Err(ctx.Err()).
				Msg("PrintBuildLogs aborted due to context cancellation")
			return fmt.Errorf("print build logs aborted: %w", ctx.Err())
		default:
			// Continue processing
		}

		// Decode the next JSON object from the build logs
		if err := decoder.Decode(&logMsg); err != nil {
			if errors.Is(err, io.EOF) {
				break // No more logs to process
			}
			log.Error().
				Err(err).
				Msg("Error decoding Docker build logs")
			return fmt.Errorf("error decoding build logs: %w", err)
		}

		// Handle error messages in the build logs
		if logMsg.Error != "" {
			log.Error().
				Str("error", logMsg.Error).
				Msg("Docker build encountered an error")
			fmt.Println(dc.errorColor.Sprintf("Error: %s", logMsg.Error))
			continue
		}

		// Handle standard build stream messages
		if logMsg.Stream != "" {
			log.Debug().
				Msgf("Docker build log: %s", logMsg.Stream)
			fmt.Print(dc.successColor.Sprintf("%s", logMsg.Stream))
		}
	}

	log.Info().Msg("Docker build process completed successfully")
	return nil
}

// CreateTarFromDirectory creates a tar archive from a filesystem directory.
// Parameters:
// - srcDir: The source directory to archive.
// Returns:
// - An io.Reader for the tar archive.
// - An error if the tar creation fails.
func CreateTarFromDirectory(srcDir string) (io.Reader, error) {
	log.Debug().Str("srcDir", srcDir).Msg("Creating tar archive from directory")
	buf := bytes.NewBuffer(make([]byte, 0, tarBufferSize))
	tw := tar.NewWriter(buf)
	defer func() {
		if err := tw.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close tar writer")
		}
	}()

	// Walk through the filesystem and add each file to the tar archive
	err := filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing file")
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not open file")
			return fmt.Errorf("could not open file: %v", err)
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Error().Err(cerr).Msg("Failed to close file")
			}
		}()

		// Create a tar header from the file info
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not create tar header")
			return fmt.Errorf("could not create tar header: %v", err)
		}

		// Set the header name to the relative path in the archive
		header.Name = filepath.ToSlash(strings.TrimPrefix(path, srcDir+"/"))

		// Write the header to the tar writer
		if err := tw.WriteHeader(header); err != nil {
			log.Error().Err(err).Str("header", header.Name).Msg("Could not write tar header")
			return fmt.Errorf("could not write tar header: %v", err)
		}

		// Copy the file contents to the tar archive
		if _, err := io.Copy(tw, file); err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not copy file contents to tar")
			return fmt.Errorf("could not copy file contents to tar: %v", err)
		}

		log.Debug().Str("file", header.Name).Msg("Added file to tar archive")
		return nil
	})
	if err != nil {
		log.Error().Err(err).Str("srcDir", srcDir).Msg("Failed to create tar archive from directory")
		return nil, fmt.Errorf("failed to create tar archive: %w", err)
	}

	return io.NopCloser(buf), nil
}

// CreateTarFromEmbedded creates a tar archive from an embedded filesystem directory.
// Parameters:
// - embeddedFS: The embedded filesystem containing the source files.
// - srcDir: The source directory within the embedded filesystem to archive.
// Returns:
// - An io.ReadCloser for the tar archive.
// - An error if the tar creation fails.
func CreateTarFromEmbedded(embeddedFS fs.FS, srcDir string) (io.ReadCloser, error) {
	log.Debug().
		Str("srcDir", srcDir).
		Msg("Creating tar archive from embedded filesystem directory")

	// Initialize a pipe for streaming the tar archive.
	reader, writer := io.Pipe()
	tw := tar.NewWriter(writer)

	// Start a goroutine to write tar data.
	go func() {
		defer func() {
			if err := tw.Close(); err != nil {
				log.Error().
					Err(err).
					Str("srcDir", srcDir).
					Msg("Failed to close tar writer")
				// Attempt to close the writer with the error.
				writer.CloseWithError(err)
			}
			// Ensure the writer is closed.
			writer.Close()
		}()

		// Walk through the embedded filesystem directory.
		err := fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Error accessing embedded file")
				return err
			}

			// Skip directories; tar will handle directory structures implicitly.
			if d.IsDir() {
				return nil
			}

			// Open the embedded file.
			file, err := embeddedFS.Open(path)
			if err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Failed to open embedded file")
				return fmt.Errorf("could not open embedded file %s: %w", path, err)
			}
			defer func() {
				if cerr := file.Close(); cerr != nil {
					log.Error().
						Err(cerr).
						Str("path", path).
						Msg("Failed to close embedded file")
				}
			}()

			// Retrieve file info.
			info, err := file.Stat()
			if err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Failed to stat embedded file")
				return fmt.Errorf("could not stat embedded file %s: %w", path, err)
			}

			// Create a tar header from the file info.
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Could not create tar header for embedded file")
				return fmt.Errorf("could not create tar header for file %s: %w", path, err)
			}

			// Set the header name to the relative path in the archive.
			relativePath := strings.TrimPrefix(path, srcDir+"/")
			header.Name = filepath.ToSlash(relativePath)

			// Write the header to the tar writer.
			if err := tw.WriteHeader(header); err != nil {
				log.Error().
					Err(err).
					Str("header", header.Name).
					Msg("Could not write tar header")
				return fmt.Errorf("could not write tar header for file %s: %w", path, err)
			}

			// Copy the file contents to the tar archive.
			if _, err := io.Copy(tw, file); err != nil {
				log.Error().
					Err(err).
					Str("path", path).
					Msg("Could not copy embedded file contents to tar")
				return fmt.Errorf("could not copy file contents for %s: %w", path, err)
			}

			log.Debug().
				Str("file", header.Name).
				Msg("Added embedded file to tar archive")
			return nil
		})

		if err != nil {
			log.Error().
				Err(err).
				Str("srcDir", srcDir).
				Msg("Failed to create tar archive from embedded filesystem")
			// Propagate the error to the reader.
			writer.CloseWithError(err)
			return
		}

		log.Info().
			Str("srcDir", srcDir).
			Msg("Successfully created tar archive from embedded filesystem")
	}()

	return reader, nil
}

// ExportImageToTar exports a Docker image to a tar file with a retry mechanism and returns the file path.
// Parameters:
// - ctx: Context for managing cancellation and timeouts.
// - imageTag: The tag of the Docker image to export.
// Returns:
// - The file path to the exported tar file.
// - An error if the export fails.
func (dc *DockerClient) ExportImageToTar(ctx context.Context, imageTag string) (string, error) {
	log.Info().
		Str("imageTag", imageTag).
		Msg("Exporting Docker image to tar file")

	var imageReader io.ReadCloser
	var err error

	retryDelay := dc.initialRetryDelay

	for attempt := 1; attempt <= dc.retries; attempt++ {
		// Check if context is done before attempting
		select {
		case <-ctx.Done():
			log.Error().
				Str("imageTag", imageTag).
				Err(ctx.Err()).
				Msg("ExportImageToTar aborted due to context cancellation before attempting")
			return "", fmt.Errorf("export image aborted due to context cancellation: %w", ctx.Err())
		default:
			// Continue
		}

		// Attempt to save the image
		imageReader, err = dc.cli.ImageSave(ctx, []string{imageTag})
		if err != nil {
			log.Error().
				Err(err).
				Str("imageTag", imageTag).
				Int("attempt", attempt).
				Msg("Failed to export Docker image")

			// Categorize the error
			if isPermanentError(err) {
				log.Error().
					Err(err).
					Str("imageTag", imageTag).
					Msg("Permanent error encountered during Docker image export. Not retrying.")
				return "", fmt.Errorf("permanent error during Docker image export: %w", err)
			}

			// If not the last attempt, wait before retrying
			if attempt < dc.retries {
				// Calculate delay with jitter
				jitter := time.Duration(float64(retryDelay) * dc.jitterFactor * (rand.Float64()*2 - 1)) // +/- jitterFactor * retryDelay
				sleepDuration := retryDelay + jitter
				if sleepDuration < 0 {
					sleepDuration = 0
				}

				log.Warn().
					Int("attempt", attempt).
					Dur("retry_delay", sleepDuration).
					Msg("Retrying ExportImageToTar after delay")

				// Wait for the calculated duration or until context is canceled
				select {
				case <-time.After(sleepDuration):
					// Continue to next attempt
				case <-ctx.Done():
					log.Error().
						Str("imageTag", imageTag).
						Err(ctx.Err()).
						Msg("ExportImageToTar aborted during retry delay due to context cancellation")
					return "", fmt.Errorf("export image aborted during retry delay: %w", ctx.Err())
				}

				// Exponential backoff
				retryDelay *= time.Duration(dc.backoffMultiplier)
				if retryDelay > dc.maxRetryDelay {
					retryDelay = dc.maxRetryDelay
				}
				continue
			}
		} else {
			// Success
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to export Docker image %s after %d attempts: %w", imageTag, dc.retries, err)
	}

	defer func() {
		if cerr := imageReader.Close(); cerr != nil {
			log.Error().
				Str("imageTag", imageTag).
				Err(cerr).
				Msg("Failed to close image reader")
		}
	}()

	// Create a temporary tar file with a unique name
	sanitizedImageTag := sanitizeImageTag(imageTag)
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-image-*.tar", sanitizedImageTag))
	if err != nil {
		log.Error().
			Err(err).
			Str("imageTag", imageTag).
			Msg("Could not create temporary tar file")
		return "", fmt.Errorf("could not create temporary tar file: %w", err)
	}

	defer func() {
		if cerr := tempFile.Close(); cerr != nil {
			log.Error().
				Err(cerr).
				Str("tarFilePath", tempFile.Name()).
				Msg("Failed to close tar file")
		}
	}()

	// Copy the image data to the tar file
	written, err := io.Copy(tempFile, imageReader)
	if err != nil {
		log.Error().
			Err(err).
			Str("tarFilePath", tempFile.Name()).
			Msg("Failed to write Docker image to tar file")
		return "", fmt.Errorf("failed to write Docker image to tar file: %w", err)
	}

	if written > maxTarSize {
		log.Error().
			Int64("bytes_written", written).
			Str("tarFilePath", tempFile.Name()).
			Msg("Exported tar file exceeds maximum allowed size")
		return "", fmt.Errorf("exported tar file size (%d bytes) exceeds the maximum allowed size (%d bytes)", written, maxTarSize)
	}

	// Set file permissions to read/write for the owner only
	if err := os.Chmod(tempFile.Name(), 0600); err != nil {
		log.Error().
			Err(err).
			Str("tarFilePath", tempFile.Name()).
			Msg("Failed to set permissions on tar file")
		return "", fmt.Errorf("failed to set permissions on tar file: %w", err)
	}

	log.Info().
		Str("tarFilePath", tempFile.Name()).
		Int64("bytes_written", written).
		Msg("Docker image exported to tar file successfully")

	return tempFile.Name(), nil
}

// sanitizeImageTag sanitizes the image tag to be used in file names by replacing or removing invalid characters.
// Parameters:
// - imageTag: The Docker image tag to sanitize.
// Returns:
// - A sanitized string safe for use in file names.
func sanitizeImageTag(imageTag string) string {
	// Replace slashes, colons, and spaces with underscores to ensure valid file paths
	invalidChars := []string{"/", ":", " "}
	sanitized := imageTag
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	return sanitized
}
