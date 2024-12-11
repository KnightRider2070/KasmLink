package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/pkg/procedures"
	"os"
	"path/filepath"
)

func init() {
	// Define the "procedures" command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Run folder initialization and Docker Compose procedures",
		Long:  `Use this command to run various procedures such as initializing folders and populating Docker Compose files.`,
	}

	// Add subcommands for procedures functionalities
	initCmd.AddCommand(
		createInitFolderStructureCommand(),
		createInitTemplatesFolderCommand(),
		createInitDockerfilesFolderCommand(),
		createInitAllTemplatesCommand(),
	)

	// Add "procedures" to the root command
	RootCmd.AddCommand(initCmd)
}

// createInitTemplatesFolderCommand initializes the templates folder with embedded templates.
func createInitTemplatesFolderCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "service-templates [folderPath]",
		Short: "Initialize the templates folder with service templates",
		Long:  `This command initializes the templates folder by copying embedded service templates into the specified folder path.`,
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
		Use:   "dockerfiles-templates [folderPath]",
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

// createInitFolderStructureCommand initializes a folder structure with 'services' and 'dockerfiles' subdirectories.
func createInitFolderStructureCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "empty-structure [rootFolderPath]",
		Short: "Initialize a folder structure with 'services' and 'dockerfiles' subdirectories",
		Long: `This command initializes a folder structure with 'services' and 'dockerfiles' subdirectories.
You need to provide the root folder path where the structure will be created.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			rootFolderPath := args[0]

			// Define the subdirectories to create
			directories := []string{
				filepath.Join(rootFolderPath, "services"),
				filepath.Join(rootFolderPath, "dockerfiles"),
			}

			// Create each directory
			for _, dir := range directories {
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					HandleError(fmt.Errorf("failed to create directory %s: %v", dir, err))
					return
				}
			}

			log.Info().Msgf("Folder structure initialized successfully at %s", rootFolderPath)
		},
	}
}

// createInitAllTemplatesCommand initializes both the Dockerfiles and service templates folders.
func createInitAllTemplatesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "all-templates [folderPath]",
		Short: "Initialize both Dockerfiles and service templates folders",
		Long: `This command initializes both the Dockerfiles and service templates folders by 
copying embedded templates for each into the specified folder path.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			folderPath := args[0]

			// Initialize the service templates folder
			err := procedures.InitTemplatesFolder(filepath.Join(folderPath))
			if err != nil {
				HandleError(fmt.Errorf("failed to initialize service templates: %v", err))
				return
			}
			log.Info().Msg("Service templates folder initialized successfully")

			// Initialize the Dockerfiles folder
			err = procedures.InitDockerfilesFolder(filepath.Join(folderPath))
			if err != nil {
				HandleError(fmt.Errorf("failed to initialize Dockerfiles: %v", err))
				return
			}
			log.Info().Msg("Dockerfiles folder initialized successfully")

		},
	}
}
