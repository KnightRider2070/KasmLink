package cmd

//
//func init() {
//	// Define the "dockerimport" command group
//	dockerImportCmd := &cobra.Command{
//		Use:   "dockerimport",
//		Short: "Import Docker images to remote nodes",
//		Long:  `Use this command to import Docker images to remote nodes by copying and loading them via SSH.`,
//	}
//
//	// Add subcommands for importing Docker images
//	dockerImportCmd.AddCommand(createImportDockerImageCommand())
//
//	// Add "dockerimport" to the root command
//	RootCmd.AddCommand(dockerImportCmd)
//}
//
//// createImportDockerImageCommand creates a command to import a Docker image tar to a remote node.
//func createImportDockerImageCommand() *cobra.Command {
//	return &cobra.Command{
//		Use:   "import-image [username] [password] [host] [localTarFilePath] [remoteDir]",
//		Short: "Import Docker image to remote node",
//		Long: `Import a Docker image to a remote node by copying the image tar file to the remote directory and executing the docker load command.
//
//		Parameters:
//		- username: SSH username to authenticate the connection.
//		- password: SSH password to authenticate the connection.
//		- host: Remote node address.
//		- localTarFilePath: Path to the local Docker image tar file.
//		- remoteDir: Directory on the remote node where the tar file will be copied.`,
//		Args: cobra.ExactArgs(5),
//		Run: func(cmd *cobra.Command, args []string) {
//			username := args[0]
//			password := args[1]
//			host := args[2]
//			localTarFilePath := args[3]
//			remoteDir := args[4]
//
//			err := procedures.ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir)
//			if err != nil {
//				fmt.Printf("Error importing Docker image: %v\n", err)
//				os.Exit(1)
//			}
//			fmt.Println("Docker image imported successfully on remote node")
//		},
//	}
//}
