package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"kasmlink/pkg/procedures"
	"strconv"
)

func init() {
	// Define the "procedures" command
	proceduresCmd := &cobra.Command{
		Use:   "procedures",
		Short: "Run folder initialization and Docker Compose procedures",
		Long:  `Use this command to run various procedures such as initializing folders and populating Docker Compose files.`,
	}

	// Add subcommands for procedures functionalities
	proceduresCmd.AddCommand(
		createInitFolderCommand(),
		createInitTemplatesFolderCommand(),
		createInitDockerfilesFolderCommand(),
		createPopulateComposeCommand(),
	)

	// Add "procedures" to the root command
	RootCmd.AddCommand(proceduresCmd)
}

// createInitFolderCommand initializes a specified folder by copying embedded templates or Dockerfiles.
func createInitFolderCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init-folder [folderPath] [subfolder] [sourcePath]",
		Short: "Initialize a folder by copying embedded templates or Dockerfiles",
		Long: `This command initializes a specified folder by copying embedded templates or Dockerfiles from the embedded file system.
You must provide the folder path, subfolder name, and the source path.`,
		Args: cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			folderPath := args[0]
			subfolder := args[1]
			sourcePath := args[2]

			err := procedures.InitFolder(folderPath, subfolder, sourcePath, embedfiles.EmbeddedTemplateFS)
			if err != nil {
				HandleError(err)
				return
			}

			log.Info().Msg("Folder initialization completed successfully")
		},
	}
}

// createInitTemplatesFolderCommand initializes the templates folder with embedded templates.
func createInitTemplatesFolderCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init-templates-folder [folderPath]",
		Short: "Initialize the templates folder with embedded templates",
		Long:  `This command initializes the templates folder by copying embedded templates into the specified folder path.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			folderPath := args[0]

			err := procedures.InitTemplatesFolder(folderPath)
			if err != nil {
				HandleError(err)
				return
			}

			log.Info().Msg("Templates folder initialized successfully")
		},
	}
}

// createInitDockerfilesFolderCommand initializes the Dockerfiles folder with embedded Dockerfile templates.
func createInitDockerfilesFolderCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init-dockerfiles-folder [folderPath]",
		Short: "Initialize the Dockerfiles folder with embedded Dockerfile templates",
		Long:  `This command initializes the Dockerfiles folder by copying embedded Dockerfile templates into the specified folder path.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			folderPath := args[0]

			err := procedures.InitDockerfilesFolder(folderPath)
			if err != nil {
				HandleError(err)
				return
			}

			log.Info().Msg("Dockerfiles folder initialized successfully")
		},
	}
}

// createPopulateComposeCommand populates a Docker Compose file with instances of a specified template.
func createPopulateComposeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "populate-compose [folderPath] [templateName] [count] [serviceNames...]",
		Short: "Populate a Docker Compose file with instances of a specified template",
		Long: `This command populates a Docker Compose file with instances of a specified template.
You need to provide the folder path, template name, count, and optional service names.`,
		Args: cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			folderPath := args[0]
			templateName := args[1]
			count, err := strconv.Atoi(args[2])
			if err != nil {
				HandleError(fmt.Errorf("invalid count value: %v", err))
				return
			}

			serviceNames := parseServiceNames(args[3:], count)

			// Create an empty compose file structure
			composeFile := &dockercompose.ComposeFile{
				Services: make(map[string]dockercompose.Service),
			}

			// Populate the Compose file using the provided template
			err = procedures.PopulateComposeWithTemplate(composeFile, folderPath, templateName, count, serviceNames)
			if err != nil {
				HandleError(err)
				return
			}

			log.Info().Msg("Docker Compose file populated successfully")
		},
	}
}

// parseServiceNames parses service names from command-line arguments.
func parseServiceNames(args []string, count int) map[int]string {
	serviceNames := make(map[int]string)
	for i, arg := range args {
		if i < count {
			serviceNames[i+1] = arg
		}
	}
	return serviceNames
}
