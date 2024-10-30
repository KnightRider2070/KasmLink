package procedures

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockerutils"
)

// BuildPostgresContainer builds a Docker image for PostgreSQL using the embedded Dockerfile.
func BuildPostgresContainer(imageTag, postgresVersion, postgresUser, postgresPassword, postgresDB string) error {
	log.Info().Str("imageTag", imageTag).Msg("Starting PostgreSQL Docker image build with custom arguments")

	// Create the Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Create tar archive from the embedded Dockerfile and build context
	buildContextTar, err := dockerutils.CreateTarFromEmbedded(embedfiles.EmbeddedDockerImagesDirectory, "dockerImages")
	if err != nil {
		return fmt.Errorf("failed to create build context tar: %v", err)
	}

	// Define Docker build arguments
	buildArgs := map[string]*string{
		"POSTGRES_VERSION":  &postgresVersion,
		"POSTGRES_USER":     &postgresUser,
		"POSTGRES_PASSWORD": &postgresPassword,
		"POSTGRES_DB":       &postgresDB,
	}

	// Build the Docker image
	return dockerutils.BuildDockerImage(cli, imageTag, "dockerfile-postgres", buildContextTar, buildArgs)
}
