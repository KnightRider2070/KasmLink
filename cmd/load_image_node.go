package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [ssh-user] [ssh-password] [host] [local-tar-file] [remote-dir]",
	Short: "Import a Docker image to a remote node",
	Long:  `Copy a Docker image tar to the remote node and import it using SSH.`,
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		password := args[1]
		host := args[2]
		localTarFilePath := args[3]
		remoteDir := args[4]

		err := procedures.ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Import process failed")
		} else {
			log.Info().Msg("Import process completed successfully")
		}
	},
}

func init() {
	// Adding import command to root
	RootCmd.AddCommand(importCmd)
}
