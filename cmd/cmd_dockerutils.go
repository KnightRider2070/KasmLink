package cmd

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"io"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockerutils"
	"os"
)

func init() {
	// Define "dockerutils" command
	dockerUtilsCmd := &cobra.Command{
		Use:   "dockerutils",
		Short: "Manage Docker utilities",
		Long:  `Commands to manage Docker utilities such as creating tar archives, copying embedded files, exporting Docker images, and building Docker images.`,
	}

	// Add subcommands for various utilities
	dockerUtilsCmd.AddCommand(
		createCreateTarWithContextCommand(),
		createCopyEmbeddedFilesCommand(),
		createExportImageToTarCommand(),
		createBuildDockerImageCommand(),
	)

	// Add "dockerutils" to the root command
	RootCmd.AddCommand(dockerUtilsCmd)
}

// createCreateTarWithContextCommand creates a command to create a tar archive of a build context.
func createCreateTarWithContextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create-tar [buildContextDir]",
		Short: "Create a tar archive from a build context directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			buildContextDir := args[0]

			tarReader, err := dockerutils.CreateTarWithContext(buildContextDir)
			if err != nil {
				HandleError(err)
				return
			}

			outputFile := fmt.Sprintf("%s.tar", buildContextDir)
			file, err := os.Create(outputFile)
			if err != nil {
				HandleError(err)
				return
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				HandleError(err)
				return
			}

			fmt.Printf("Tar archive created at %s\n", outputFile)
		},
	}
}

// createCopyEmbeddedFilesCommand creates a command to copy files from the embedded filesystem.
func createCopyEmbeddedFilesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "copy-embedded [srcDir] [dstDir]",
		Short: "Copy files from the embedded filesystem to a target directory",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			srcDir := args[0]
			dstDir := args[1]

			//TODO: Enhance embdedded files
			err := dockerutils.CopyEmbeddedFiles(embedfiles.EmbeddedDockerImagesDirectory, srcDir, dstDir)
			if err != nil {
				HandleError(err)
				return
			}

			fmt.Printf("Files copied from %s to %s successfully\n", srcDir, dstDir)
		},
	}
}

// createExportImageToTarCommand creates a command to export a Docker image to a tar file.
func createExportImageToTarCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export-image [imageTag]",
		Short: "Export a Docker image to a tar file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			imageTag := args[0]

			// Create Docker client
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				HandleError(err)
				return
			}

			tarFilePath, err := dockerutils.ExportImageToTar(cli, imageTag)
			if err != nil {
				HandleError(err)
				return
			}

			fmt.Printf("Docker image exported to %s successfully\n", tarFilePath)
		},
	}
}

// createBuildDockerImageCommand creates a command to build a Docker image from a tar context.
func createBuildDockerImageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build-image [buildContextDir] [dockerfilePath] [imageTag]",
		Short: "Build a Docker image from a build context directory",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			buildContextDir := args[0]
			dockerfilePath := args[1]
			imageTag := args[2]

			// Create Docker client
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				HandleError(err)
				return
			}

			// Create tar archive from the build context directory
			buildContextTar, err := dockerutils.CreateTarWithContext(buildContextDir)
			if err != nil {
				HandleError(err)
				return
			}

			// Call BuildDockerImage with the tar build context
			err = dockerutils.BuildDockerImage(cli, imageTag, dockerfilePath, buildContextTar, nil)
			if err != nil {
				HandleError(err)
				return
			}

			fmt.Printf("Docker image %s built successfully\n", imageTag)
		},
	}
}
