package cmd

import (
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"log"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a Docker image to a remote node",
	Long:  `Copy a Docker image tar to the remote node and import it using SSH.`,
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("ssh-user")
		password, _ := cmd.Flags().GetString("ssh-password")
		host, _ := cmd.Flags().GetString("host")
		localTarFilePath, _ := cmd.Flags().GetString("local-tar-file")
		remoteDir, _ := cmd.Flags().GetString("remote-dir")

		err := procedures.ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir)
		if err != nil {
			log.Fatalf("Import process failed: %v", err)
		} else {
			log.Println("Import process completed successfully")
		}
	},
}

func init() {
	// Adding import command to root
	RootCmd.AddCommand(importCmd)

	// Adding flags to import command
	importCmd.Flags().String("ssh-user", "username", "SSH username for remote node")
	importCmd.Flags().String("ssh-password", "password", "SSH password for remote node")
	importCmd.Flags().String("host", "192.168.0.10", "Remote host address")
	importCmd.Flags().String("local-tar-file", "./core-image.tar", "Path to local Docker image tar file")
	importCmd.Flags().String("remote-dir", "/tmp", "Directory on remote node to copy tar file")
}
