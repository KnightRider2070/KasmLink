package internal

/*
import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockerutils"
)

// BuildPostgresContainer builds a Docker image for PostgreSQL using the embedded Dockerfile.
func BuildPostgresContainer(imageTag, postgresVersion, postgresUser, postgresPassword, postgresDB string) error {
	log.Info().Str("imageTag", imageTag).Msg("Starting PostgreSQL Docker image build with custom arguments")

	// Create the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Docker client")
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Create tar archive from the embedded Dockerfile and build context
	buildContextTar, err := dockercli.CreateTarFromEmbedded(embedfiles.EmbeddedDockerImagesDirectory, "dockerfiles")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create build context tar")
		return fmt.Errorf("failed to create build context tar: %v", err)
	}

	// Log build arguments with enhanced context and security
	log.Info().
		Str("imageTag", imageTag).
		Str("POSTGRES_VERSION", postgresVersion).
		Str("POSTGRES_USER", postgresUser).
		Str("POSTGRES_DB", postgresDB).
		Msg("Docker build arguments")
	log.Debug().Str("POSTGRES_PASSWORD", "********").Msg("Docker build password argument (hidden for security)")

	// Define Docker build arguments
	buildArgs := map[string]*string{
		"POSTGRES_VERSION":  &postgresVersion,
		"POSTGRES_USER":     &postgresUser,
		"POSTGRES_PASSWORD": &postgresPassword,
		"POSTGRES_DB":       &postgresDB,
	}

	log.Info().Msg("Starting Docker image build process")
	if err := dockerutils.BuildDockerImage(cli, imageTag, "dockerfile-postgres", buildContextTar, buildArgs); err != nil {
		log.Error().Err(err).Msg("Failed to build Docker image for PostgreSQL")
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	log.Info().Str("imageTag", imageTag).Msg("Successfully built PostgreSQL Docker image")
	return nil
}
*/
