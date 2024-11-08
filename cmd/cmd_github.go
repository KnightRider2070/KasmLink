package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	shadowhub "kasmlink/pkg/github"
	"time"
)

func init() {
	// Define "shadowhub" command
	shadowHubCmd := &cobra.Command{
		Use:   "shadowhub",
		Short: "Manage ShadowHub dependencies",
		Long:  `Commands to manage ShadowHub dependencies such as updating scripts from upstream repositories.`,
	}

	// Add subcommands for various utilities
	shadowHubCmd.AddCommand(
		createUpdateShadowDependenciesCommand(),
	)

	// Add "shadowhub" to the root command
	RootCmd.AddCommand(shadowHubCmd)
}

// createUpdateShadowDependenciesCommand creates a command to update dependencies in a workspace.
func createUpdateShadowDependenciesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update-shadow-dependencies [workspaceImageFilePath] [token]",
		Short: "Update dependencies for all scripts in a workspace",
		Long: `This command updates dependencies for all scripts in the specified workspace by comparing the local scripts to the upstream versions on GitHub.
Provide the path to the workspace and an optional GitHub token for authenticated access.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			workspaceImageFilePath := args[0]
			token := args[1]

			log.Info().Str("workspace_path", workspaceImageFilePath).Msg("Updating shadow dependencies in workspace")
			startTime := time.Now()

			shadowhub.UpdateShadowDependencies(workspaceImageFilePath, token)

			duration := time.Since(startTime)
			log.Info().Dur("duration", duration).Msg("UpdateShadowDependencies command completed")
		},
	}
}
