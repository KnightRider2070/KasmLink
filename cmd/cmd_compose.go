package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"kasmlink/pkg/dockercompose"
	"kasmlink/pkg/procedures"
	"os"
	"strconv"
)

// Init initializes the root command.
func init() {
	// Define "compose" command
	composeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Manage Docker Compose files",
		Long:  `Commands to manage and generate Docker Compose files, including creating new ones, loading from YAML, and using embedded templates.`,
	}

	// Add subcommands for generating Docker Compose files
	composeCmd.AddCommand(createPopulateComposeWithTemplateCommand())

	// Add "compose" to the root command
	RootCmd.AddCommand(composeCmd)

}

// createPopulateComposeWithTemplateCommand populates or creates a Docker Compose file with instances of a specified template.
func createPopulateComposeWithTemplateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [composeFilePath] [templateFolderPath] [templateName] [count] [serviceNames...]",
		Short: "Populate or create a Docker Compose file with instances of a specified template",
		Long: `This command populates or creates a Docker Compose file with instances of a specified template.
You need to provide the compose file path, the folder path containing the template YAML, the template name, 
the count of service instances to create, and optional service names for serialization.`,
		Args: cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			composeFilePath := args[0]
			templateFolderPath := args[1]
			templateName := args[2]
			count, err := strconv.Atoi(args[3])
			if err != nil {
				log.Error().Err(err).Msg("Invalid count value")
				fmt.Printf("Error: invalid count value - %v\n", err)
				os.Exit(1)
			}

			// Parse optional service names
			serviceNames := parseServiceNames(args[4:], count)

			// Load or initialize the ComposeFile struct
			var composeFile dockercompose.ComposeFile
			if _, err := os.Stat(composeFilePath); os.IsNotExist(err) {
				log.Info().Str("composeFilePath", composeFilePath).Msg("Compose file does not exist, creating a new one")
				// Initialize a new empty compose file structure
				composeFile = dockercompose.ComposeFile{
					Version:  "3.8",
					Services: make(map[string]dockercompose.Service),
					Networks: make(map[string]dockercompose.Network),
					Volumes:  make(map[string]dockercompose.Volume),
					Configs:  make(map[string]dockercompose.Config),
					Secrets:  make(map[string]dockercompose.Secret),
				}
			} else {
				// Load existing Compose file if it exists
				loadedComposeFile, err := dockercompose.LoadComposeFile(composeFilePath)
				if err != nil {
					log.Error().Err(err).Str("composeFilePath", composeFilePath).Msg("Failed to load existing Compose file")
					fmt.Printf("Error loading compose file: %v\n", err)
					os.Exit(1)
				}
				composeFile = *loadedComposeFile // Dereference the pointer
			}

			// Populate the Compose file with the specified template
			err = procedures.PopulateComposeWithTemplate(&composeFile, templateFolderPath, templateName, count, serviceNames)
			if err != nil {
				log.Error().Err(err).Msg("Failed to populate compose file with template")
				fmt.Printf("Error populating compose file with template: %v\n", err)
				os.Exit(1)
			}

			// Save the populated Compose file back to the specified path
			err = saveComposeFile(composeFilePath, composeFile)
			if err != nil {
				log.Error().Err(err).Str("composeFilePath", composeFilePath).Msg("Failed to save populated Compose file")
				fmt.Printf("Error saving populated Compose file: %v\n", err)
				os.Exit(1)
			}

			log.Info().Str("composeFilePath", composeFilePath).Msg("Docker Compose file populated successfully")
			fmt.Println("Docker Compose file populated successfully")
		},
	}
}

// Helper function to save a Docker Compose file.
func saveComposeFile(filePath string, composeFile dockercompose.ComposeFile) error {
	fileData, err := yaml.Marshal(&composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file to YAML: %w", err)
	}

	// Write data to file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}
	return nil
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
