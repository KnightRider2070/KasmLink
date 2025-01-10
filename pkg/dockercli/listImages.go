package dockercli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// DockerImage represents the structure of a Docker image as returned by `docker images`.
type DockerImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"id"`
	Size       string `json:"size"`
}

// ListImagesOptions defines the options for listing Docker images.
type ListImagesOptions struct {
	Repository string   // Filter by repository.
	Tag        string   // Filter by tag.
	Additional []string // Additional custom filters or options.
}

// ListImages lists all Docker images using the configured executor (local or SSH).
func (dc *DockerClient) ListImages(ctx context.Context, options ListImagesOptions) ([]DockerImage, error) {
	log.Info().Msg("Starting Docker image listing")

	// Build the command for listing Docker images.
	cmd := []string{"docker", "images", "--format", "{{json .}}"}
	if options.Repository != "" {
		cmd = append(cmd, options.Repository)
	}
	if len(options.Additional) > 0 {
		cmd = append(cmd, options.Additional...)
	}

	// Execute the command using the configured executor.
	output, err := dc.executor.Execute(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker images: %w", err)
	}

	// Parse and filter results.
	images, err := parseDockerImages(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Docker images: %w", err)
	}
	return filterDockerImages(images, options), nil
}

// parseDockerImages parses the output of the `docker images` command into a slice of DockerImage structs.
func parseDockerImages(output []byte) ([]DockerImage, error) {
	var images []DockerImage
	lines := bytes.Split(output, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var image DockerImage
		if err := json.Unmarshal(line, &image); err != nil {
			log.Warn().Err(err).Msg("Failed to parse Docker image JSON")
			continue
		}
		images = append(images, image)
	}

	log.Info().Int("count", len(images)).Msg("Docker images parsed successfully")
	return images, nil
}

// filterDockerImages applies filters like repository and tag to a list of Docker images.
func filterDockerImages(images []DockerImage, options ListImagesOptions) []DockerImage {
	var filtered []DockerImage
	for _, img := range images {
		if options.Repository != "" && img.Repository != options.Repository {
			continue
		}
		if options.Tag != "" && img.Tag != options.Tag {
			continue
		}
		filtered = append(filtered, img)
	}

	log.Info().Int("filtered_count", len(filtered)).Msg("Docker images filtered successfully")
	return filtered
}
