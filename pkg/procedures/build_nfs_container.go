package procedures

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockerutils"
)

// BuildNFSContainer builds a Docker image for an NFS server using the embedded Dockerfile.
func BuildNFSContainer(imageTag, domain, exportDir, exportNetwork, nfsVersion string) error {
	log.Info().Str("imageTag", imageTag).Msg("Starting NFS Docker image build with custom arguments")

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
		"DOMAIN":         &domain,
		"EXPORT_DIR":     &exportDir,
		"EXPORT_NETWORK": &exportNetwork,
		"NFS_VERSION":    &nfsVersion,
	}

	// Build the Docker image
	return dockerutils.BuildDockerImage(cli, imageTag, "dockerfile-nfs-server", buildContextTar, buildArgs)
}
