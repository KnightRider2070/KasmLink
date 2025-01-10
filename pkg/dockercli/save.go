package dockercli

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
)

// SaveImage saves a Docker image to a tarball file.
func (dc *DockerClient) SaveImage(ctx context.Context, image, tarPath string) error {
	log.Info().Str("image", image).Str("tar_path", tarPath).Msg("Saving Docker image to tarball")
	dockerCommand := []string{"docker", "save", "-o", tarPath, image}
	output, err := dc.executor.Execute(ctx, dockerCommand[0], dockerCommand[1:]...)
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to save Docker image")
		return fmt.Errorf("failed to save Docker image %s: %w", image, err)
	}
	log.Info().Str("tar_path", tarPath).Msg("Docker image saved successfully")
	return nil
}
