package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

// buildNFSCommand creates a Docker image for an NFS server with custom parameters.
var buildNFSCommand = &cobra.Command{
	Use:   "build-nfs [image-tag] [domain] [export-dir] [export-network] [nfs-version]",
	Short: "Build a Docker image for an NFS server",
	Long:  `Builds a Docker image for an NFS server using custom configuration options specified as arguments.`,
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		domain := args[1]
		exportDir := args[2]
		exportNetwork := args[3]
		nfsVersion := args[4]

		// Call BuildNFSContainer with the provided arguments
		err := procedures.BuildNFSContainer(imageTag, domain, exportDir, exportNetwork, nfsVersion)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to build NFS Docker image")
		} else {
			log.Info().Msg("NFS Docker image built successfully")
		}
	},
}

func init() {
	// Add buildNFSCommand to the root command
	RootCmd.AddCommand(buildNFSCommand)
}
