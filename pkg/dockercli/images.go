package dockercli

import (
	"context"
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
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to pull Docker image")
		return fmt.Errorf("failed to pull image: %w", err)
	}
	log.Info().Str("image_name", imageName).Msg("Docker image pulled successfully")
	return nil
}

// PushImage pushes a Docker image to a registry with retry mechanism.
func PushImage(ctx context.Context, retries int, imageName string) error {
	log.Info().Str("image_name", imageName).Msg("Pushing Docker image")
	output, err := executeDockerCommand(ctx, retries, "docker", "push", imageName)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to push Docker image")
		return fmt.Errorf("failed to push image: %w", err)
	}
	log.Info().Str("image_name", imageName).Msg("Docker image pushed successfully")
	return nil
}

// RemoveImage removes a Docker image by name or ID with retry mechanism.
func RemoveImage(ctx context.Context, retries int, imageName string) error {
	log.Info().Str("image_name", imageName).Msg("Removing Docker image")
	output, err := executeDockerCommand(ctx, retries, "docker", "rmi", imageName)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to remove Docker image")
		return fmt.Errorf("failed to remove image: %w", err)
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
	// Step 1: List all Docker images
	images, err := ListImages(ctx, retries)
	if err != nil {
		return "", fmt.Errorf("failed to list Docker images: %v", err)
	}

	// Step 2: Find the image that matches the provided tag
	var imageID string
	for _, image := range images {
		// Compare the image tag with the input tag
		if image == imageTag {
			// Step 3: Use docker inspect to retrieve the Image ID
			output, err := executeDockerCommand(ctx, retries, "docker", "inspect", "--format", "{{.Id}}", image)
			if err != nil {
				log.Error().Err(err).Str("image", image).Msg("Failed to inspect Docker image")
				return "", fmt.Errorf("failed to inspect Docker image %s: %v", image, err)
			}
			imageID = string(output)
			break
		}
	}

	if imageID == "" {
		log.Warn().Str("imageTag", imageTag).Msg("No matching image found for tag")
		return "", fmt.Errorf("no matching image found for tag %s", imageTag)
	}

	log.Info().Str("imageTag", imageTag).Str("imageID", imageID).Msg("Found matching image ID")
	return imageID, nil
}

// UpdateAllImages pulls the latest version of all present Docker images with retry mechanism.
func UpdateAllImages(ctx context.Context, retries int) error {
	images, err := ListImages(ctx, retries)
	if err != nil {
		return fmt.Errorf("failed to list Docker images: %w", err)
	}

	for _, image := range images {
		log.Info().Str("image_name", image).Msg("Updating Docker image")
		err := PullImage(ctx, retries, image)
		if err != nil {
			log.Error().Err(err).Str("image_name", image).Msg("Failed to update Docker image")
			return fmt.Errorf("failed to update image %s: %w", image, err)
		}
	}

	log.Info().Msg("All Docker images have been updated successfully")
	return nil
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
		if err := imageReader.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close image reader")
		}
	}()

	// Determine the output file path
	if outputFile == "" {
		outputFile = filepath.Join(os.TempDir(), fmt.Sprintf("%s-image.tar", imageName))
	}

	// Create the output tar file
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Error().Err(err).Str("output_file", outputFile).Msg("Failed to create tar file")
		return "", fmt.Errorf("could not create tar file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil {
			err = fmt.Errorf("failed to close tar file: %v", cerr)
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

// ImportImageFromTar imports a Docker image from a tar file with retry mechanism.
func ImportImageFromTar(ctx context.Context, retries int, tarFilePath string) error {
	log.Info().Str("tar_file_path", tarFilePath).Msg("Importing Docker image from tar file")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("Could not create Docker client")
		return fmt.Errorf("could not create Docker client: %w", err)
	}

	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		log.Error().Err(err).Str("tar_file_path", tarFilePath).Msg("Failed to open tar file")
		return fmt.Errorf("could not open tar file: %w", err)
	}
	defer func() {
		if err := tarFile.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close tar file")
		}
	}()

	imageLoadResponse, err := cli.ImageLoad(context.Background(), tarFile, true)
	if err != nil {
		log.Error().Err(err).Str("tar_file_path", tarFilePath).Msg("Failed to load Docker image from tar")
		return fmt.Errorf("could not load Docker image from tar: %w", err)
	}

	defer func() {
		if err := imageLoadResponse.Body.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close image load response body")
		}
	}()

	_, err = io.Copy(os.Stdout, imageLoadResponse.Body)
	if err != nil {
		log.Error().Err(err).Str("tar_file_path", tarFilePath).Msg("Failed to read image load response")
		return fmt.Errorf("could not read image load response: %w", err)
	}

	log.Info().Str("tar_file_path", tarFilePath).Msg("Docker image imported from tar file successfully")
	return nil
}
