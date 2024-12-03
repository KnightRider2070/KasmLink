package procedures

/*
import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercli"
)

// BuildNFSContainer builds a Docker image for an NFS server using the embedded Dockerfile.
func BuildNFSContainer(imageTag, domain, exportDir, exportNetwork, nfsVersion string) error {
	log.Info().Str("imageTag", imageTag).Msg("Starting NFS Docker image build with custom arguments")

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
		Str("DOMAIN", domain).
		Str("EXPORT_DIR", exportDir).
		Str("EXPORT_NETWORK", exportNetwork).
		Str("NFS_VERSION", nfsVersion).
		Msg("Docker build arguments")

	// Define Docker build arguments
	buildArgs := map[string]*string{
		"DOMAIN":         &domain,
		"EXPORT_DIR":     &exportDir,
		"EXPORT_NETWORK": &exportNetwork,
		"NFS_VERSION":    &nfsVersion,
	}

	log.Info().Msg("Starting Docker image build process")
	if err := dockerutils.BuildDockerImage(cli, imageTag, "dockerfile-nfs-server", buildContextTar, buildArgs); err != nil {
		log.Error().Err(err).Msg("Failed to build Docker image for NFS server")
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	log.Info().Str("imageTag", imageTag).Msg("Successfully built NFS Docker image")
	return nil
}
*/
