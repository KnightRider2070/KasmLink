package cmd

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	shadowhub "kasmlink/pkg/github"
)

// Initialize "shadowhub" command and its subcommands.
func init() {
	// Define the root "shadowhub" command.
	shadowHubCmd := &cobra.Command{
		Use:   "shadowhub",
		Short: "Manage ShadowHub dependencies",
		Long:  `Commands to manage ShadowHub dependencies such as updating scripts from upstream repositories.`,
	}

	// Add subcommands.
	shadowHubCmd.AddCommand(createUpdateShadowDependenciesCommand())

	// Add "shadowhub" to the root command.
	RootCmd.AddCommand(shadowHubCmd)
}

// createUpdateShadowDependenciesCommand creates a command to update dependencies in a workspace.
func createUpdateShadowDependenciesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update-shadow-dependencies [workspaceImageFilePath] [token] [owner] [repo]",
		Short: "Update dependencies for all scripts in a workspace",
		Long: `This command updates dependencies for all scripts in the specified workspace by comparing the local scripts to the upstream versions on GitHub.
Provide the path to the workspace, an optional GitHub token for authenticated access, and the GitHub owner/repo details.`,
		Args: cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			workspaceImageFilePath := args[0]
			token := args[1]
			owner := args[2]
			repo := args[3]

			log.Info().
				Str("workspace_path", workspaceImageFilePath).
				Str("owner", owner).
				Str("repo", repo).
				Msg("Updating shadow dependencies in workspace")

			startTime := time.Now()

			// Initialize the GitHub client.
			ghClient := shadowhub.NewGitHubClient(token)

			// Define variables to check in scripts.
			variables := []string{"BASE_URL", "API_ENDPOINT"}

			// Initialize the ScriptProcessor.
			processor := shadowhub.NewScriptProcessor(ghClient, variables, workspaceImageFilePath, owner, repo)

			// Process scripts in the workspace.
			ctx := context.Background()
			processor.ProcessScripts(ctx)

			duration := time.Since(startTime)
			log.Info().
				Dur("duration", duration).
				Msg("UpdateShadowDependencies command completed")
		},
	}
}
