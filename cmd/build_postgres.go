package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

// buildPostgresCommand creates a Docker image for PostgreSQL with custom parameters.
var buildPostgresCommand = &cobra.Command{
	Use:   "build-postgres [image-tag] [postgres-version] [postgres-user] [postgres-password] [postgres-db]",
	Short: "Build a Docker image for PostgreSQL",
	Long:  `Builds a Docker image for PostgreSQL using custom configuration options specified as arguments.`,
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		postgresVersion := args[1]
		postgresUser := args[2]
		postgresPassword := args[3]
		postgresDB := args[4]

		// Call BuildPostgresContainer with the provided arguments
		err := procedures.BuildPostgresContainer(imageTag, postgresVersion, postgresUser, postgresPassword, postgresDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to build PostgreSQL Docker image")
		} else {
			log.Info().Msg("PostgreSQL Docker image built successfully")
		}
	},
}

func init() {
	// Add buildPostgresCommand to the root command
	RootCmd.AddCommand(buildPostgresCommand)
}
