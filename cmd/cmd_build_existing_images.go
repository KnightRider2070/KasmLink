package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"os"
)

func init() {
	// Define the "dockerbuild" command group
	dockerBuildCmd := &cobra.Command{
		Use:   "dockerbuild",
		Short: "Manage Docker container builds",
		Long:  `Use this command to manage Docker container builds, including NFS, Postgres, and others.`,
	}

	// Add subcommands for various Docker builds
	dockerBuildCmd.AddCommand(createNFSBuildCommand())
	dockerBuildCmd.AddCommand(createPostgresBuildCommand())

	// Add "dockerbuild" to the root command
	RootCmd.AddCommand(dockerBuildCmd)
}

// createNFSBuildCommand creates a command to build an NFS Docker container.
func createNFSBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build-nfs [imageTag] [domain] [exportDir] [exportNetwork] [nfsVersion]",
		Short: "Build the NFS Docker container",
		Long: `Build the NFS Docker container using the specified image tag, domain, export directory, 
export network, and NFS version.`,
		Args: cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			imageTag := args[0]
			domain := args[1]
			exportDir := args[2]
			exportNetwork := args[3]
			nfsVersion := args[4]

			err := procedures.BuildNFSContainer(imageTag, domain, exportDir, exportNetwork, nfsVersion)
			if err != nil {
				fmt.Printf("Error building NFS container: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("NFS Docker container built successfully")
		},
	}
}

// createPostgresBuildCommand creates a command to build a PostgreSQL Docker container.
func createPostgresBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build-postgres [imageTag] [postgresVersion] [postgresUser] [postgresPassword] [postgresDB]",
		Short: "Build the PostgreSQL Docker container",
		Long: `Build the PostgreSQL Docker container using the specified image tag, PostgreSQL version,
Postgres user, password, and database.`,
		Args: cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			imageTag := args[0]
			postgresVersion := args[1]
			postgresUser := args[2]
			postgresPassword := args[3]
			postgresDB := args[4]

			err := procedures.BuildPostgresContainer(imageTag, postgresVersion, postgresUser, postgresPassword, postgresDB)
			if err != nil {
				fmt.Printf("Error building PostgreSQL container: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("PostgreSQL Docker container built successfully")
		},
	}
}
