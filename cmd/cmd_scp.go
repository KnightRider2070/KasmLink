package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	shadowscp "kasmlink/pkg/scp"
)

func init() {
	// Define the "shadowscp" command
	shadowSCPCmd := &cobra.Command{
		Use:   "shadowscp",
		Short: "Manage secure file copy to remote nodes",
		Long:  `Use this command to copy files to a remote node via SSH with enhanced functionality using shadowscp package.`,
	}

	// Add subcommands for shadowscp functionalities
	shadowSCPCmd.AddCommand(
		createShadowCopyFileCommand(),
	)

	// Add "shadowscp" to the root command
	RootCmd.AddCommand(shadowSCPCmd)
}

// createShadowCopyFileCommand creates a command for copying a file to a remote node via SSH.
func createShadowCopyFileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copy [localFilePath] [remoteDir]",
		Short: "Copy a local file to a remote directory via SSH",
		Long: `Copy a local file to a remote directory using SSH. 
You need to specify the path to the local file and the remote directory.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			localFilePath := args[0]
			remoteDir := args[1]

			log.Info().Str("localFilePath", localFilePath).Str("remoteDir", remoteDir).Msg("Attempting to copy file via SSH")

			err := shadowscp.ShadowCopyFile(localFilePath, remoteDir)
			if err != nil {
				HandleError(err)
				return
			}

			log.Info().Msg("File copied successfully via SSH")
		},
	}
}
