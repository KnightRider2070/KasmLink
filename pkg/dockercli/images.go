// dockercli/dockercli.go
package dockercli

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PullImage pulls a Docker image from a registry with retry mechanism.
func PullImage(ctx context.Context, retries int, imageName string) error {
	log.Info().Str("image_name", imageName).Msg("Pulling Docker image")
	output, err := executeDockerCommand(ctx, retries, "docker", "pull", imageName)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Str("image_name", imageName).Msg("Failed to pull Docker image")
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	log.Info().Str("image_name", imageName).Msg("Docker image pulled successfully")
	return nil
}

// RemoveImage removes a Docker image by name or ID with retry mechanism.
func RemoveImage(ctx context.Context, retries int, imageName string) error {
	log.Info().Str("image_name", imageName).Msg("Removing Docker image")
	output, err := executeDockerCommand(ctx, retries, "docker", "rmi", imageName)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Str("image_name", imageName).Msg("Failed to remove Docker image")
		return fmt.Errorf("failed to remove image %s: %w", imageName, err)
	}
	log.Info().Str("image_name", imageName).Msg("Docker image removed successfully")
	return nil
}

// ListImages lists all Docker images on the host with retry mechanism.
func ListImages(ctx context.Context, retries int) ([]string, error) {
	log.Info().Msg("Listing all Docker images")
	output, err := executeDockerCommand(ctx, retries, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to list Docker images")
		return nil, fmt.Errorf("failed to list Docker images: %w", err)
	}

	images := strings.Split(strings.TrimSpace(string(output)), "\n")
	log.Info().Int("image_count", len(images)).Msg("Docker images listed successfully")
	return images, nil
}

// GetImageIDByTag retrieves the Image ID for a given image tag.
func GetImageIDByTag(ctx context.Context, retries int, imageTag string) (string, error) {
	// Step 1: Inspect the Docker image to retrieve its ID
	log.Info().Str("image_tag", imageTag).Msg("Retrieving Docker image ID by tag")
	output, err := executeDockerCommand(ctx, retries, "docker", "inspect", "--format", "{{.Id}}", imageTag)
	if err != nil {
		log.Error().Err(err).Str("image_tag", imageTag).Msg("Failed to inspect Docker image")
		return "", fmt.Errorf("failed to inspect Docker image %s: %w", imageTag, err)
	}

	imageID := strings.TrimSpace(string(output))
	if imageID == "" {
		log.Warn().Str("image_tag", imageTag).Msg("No Image ID found for the provided tag")
		return "", fmt.Errorf("no Image ID found for tag %s", imageTag)
	}

	log.Info().Str("image_tag", imageTag).Str("image_id", imageID).Msg("Docker image ID retrieved successfully")
	return imageID, nil
}

// ExportImageToTar exports a Docker image to a tar file with retry mechanism.
// If outputFile is an empty string, it creates the tar file in a temporary directory.
func ExportImageToTar(ctx context.Context, retries int, imageName, outputFile string) (string, error) {
	log.Info().Str("image_name", imageName).Str("output_file", outputFile).Msg("Exporting Docker image to tar file")

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("Could not create Docker client")
		return "", fmt.Errorf("could not create Docker client: %w", err)
	}

	// Save the Docker image to a tar stream
	imageReader, err := cli.ImageSave(context.Background(), []string{imageName})
	if err != nil {
		log.Error().Err(err).Str("image_name", imageName).Msg("Failed to save Docker image")
		return "", fmt.Errorf("could not save Docker image: %w", err)
	}

	defer func() {
		if cerr := imageReader.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("Failed to close image reader")
		}
	}()

	// Determine the output file path
	if outputFile == "" {
		outputFile = filepath.Join(os.TempDir(), fmt.Sprintf("%s-image.tar", strings.ReplaceAll(imageName, "/", "_")))
	}

	// Create the output tar file
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Error().Err(err).Str("output_file", outputFile).Msg("Failed to create tar file")
		return "", fmt.Errorf("could not create tar file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			log.Error().Err(cerr).Str("output_file", outputFile).Msg("Failed to close tar file")
		}
	}()

	// Write the image data to the tar file
	written, err := io.Copy(outFile, imageReader)
	if err != nil {
		log.Error().Err(err).Str("output_file", outputFile).Msg("Failed to write Docker image to tar file")
		return "", fmt.Errorf("could not write image to tar file: %w", err)
	}

	log.Info().
		Str("image_name", imageName).
		Str("output_file", outputFile).
		Int64("bytes_written", written).
		Msg("Docker image exported to tar file successfully")

	return outputFile, nil
}

// BuildDockerImage builds a Docker image from a Dockerfile with retry mechanism.
func BuildDockerImage(ctx context.Context, retries int, dockerfilePath, imageName string) error {
	log.Info().Str("dockerfile_path", dockerfilePath).Str("image_name", imageName).Msg("Building Docker image")

	// Ensure the Dockerfile exists
	if _, err := os.Stat(dockerfilePath); errors.Is(err, os.ErrNotExist) {
		log.Error().Str("dockerfile_path", dockerfilePath).Msg("Dockerfile does not exist")
		return fmt.Errorf("dockerfile does not exist at path %s", dockerfilePath)
	}

	// Determine the build context directory (parent directory of Dockerfile)
	buildContext := filepath.Dir(dockerfilePath)

	// Execute the Docker build command with retries
	output, err := executeDockerCommand(ctx, retries, "docker", "build", "-t", imageName, "-f", dockerfilePath, buildContext)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Str("image_name", imageName).Msg("Failed to build Docker image")
		return fmt.Errorf("failed to build Docker image %s: %w", imageName, err)
	}

	log.Info().Str("image_name", imageName).Msg("Docker image built successfully")
	return nil
}
