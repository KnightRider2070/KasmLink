package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

// buildPostgresCommand creates a Docker image for PostgreSQL with custom parameters.
var buildPostgresCommand = &cobra.Command{
	Use:   "build-postgres [image-tag] [postgres-user] [postgres-password] [postgres-db]",
	Short: "Build a Docker image for PostgreSQL",
	Long: `Builds a Docker image for PostgreSQL using custom configuration options specified as arguments.
The postgres-version flag is optional, defaulting to version 13 if omitted.`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		imageTag := args[0]
		postgresUser := args[1]
		postgresPassword := args[2]
		postgresDB := args[3]

		// Retrieve the optional postgres-version flag, defaulting to "13" if not set
		postgresVersion, err := cmd.Flags().GetString("postgres-version")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to retrieve postgres version")
		}

		// Call BuildPostgresContainer with the provided arguments
		err = procedures.BuildPostgresContainer(imageTag, postgresVersion, postgresUser, postgresPassword, postgresDB)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to build PostgreSQL Docker image")
		} else {
			log.Info().Msg("PostgreSQL Docker image built successfully")
			log.Info().Msg("Note: To access the PostgreSQL container from outside Docker, make sure to publish the port when starting the container.")
			log.Info().Msg(`Example: docker run -d -p 5432:5432 --name <container_name> <image_tag>`)
		}
	},
}

func init() {
	// Define the optional postgres-version flag with a default value of "13"
	buildPostgresCommand.Flags().String("postgres-version", "13", "Specify the PostgreSQL version (optional, defaults to 13)")

	// Add buildPostgresCommand to the root command
	RootCmd.AddCommand(buildPostgresCommand)
}
