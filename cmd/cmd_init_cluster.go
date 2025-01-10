package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal/talos"
	"time"
)

// Initialize the "init" command and its subcommands.
func init() {
	// Define the root "init" command.
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize cluster configuration and resources",
		Long:  `Initialize required directories, configurations, and resources for the cluster deployment.`,
		Run:   runInitCommand,
	}

	// Add "init" to the root command.
	RootCmd.AddCommand(initCmd)
}

// runInitCommand executes the initialization process.
func runInitCommand(cmd *cobra.Command, args []string) {
	log.Info().Msg("Starting cluster initialization process...")

	startTime := time.Now()

	// Call the InitFiles function from the talos package.
	if err := talos.InitFiles(); err != nil {
		log.Error().Err(err).Msg("Cluster initialization failed")
		return
	}

	duration := time.Since(startTime)
	log.Info().
		Dur("duration", duration).
		Msg("Cluster initialization completed successfully")
}
