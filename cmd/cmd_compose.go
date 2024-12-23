package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal"
	"kasmlink/pkg/dockercompose"
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
				log.Error().Msgf("Error: invalid count value - %v", err)
				os.Exit(1)
			}

			// Parse optional service names
			serviceNames := parseServiceNames(args[4:], count)

			// Create serviceFilePath
			serviceFilePath := fmt.Sprintf("%s/%s.yaml", templateFolderPath, templateName)
			// Load service file
			serviceComposeFile := dockercompose.ComposeFile{}
			serviceComposeFilePtr, err := dockercompose.LoadComposeFile(serviceFilePath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to load compose file")
				os.Exit(1)
			}
			serviceComposeFile = *serviceComposeFilePtr

			log.Debug().Interface("serviceComposeFile", serviceComposeFile).Msg("Loaded service compose file")

			// Load the output compose file if present otherwise create a new one
			composeFile := dockercompose.ComposeFile{}
			composeFilePtr, err := dockercompose.LoadComposeFile(composeFilePath)
			if err != nil {
				log.Warn().Str("path", composeFilePath).Msg("Compose file does not exist, creating a new one")
				composeFile = dockercompose.ComposeFile{
					Version:  "3.8", // Set default version or any other default values
					Services: make(map[string]dockercompose.Service),
					Networks: make(map[string]dockercompose.Network),
					Volumes:  make(map[string]dockercompose.Volume),
					Configs:  make(map[string]dockercompose.Config),
					Secrets:  make(map[string]dockercompose.Secret),
				}
			} else {
				composeFile = *composeFilePtr
			}

			// Use composeFile in subsequent code
			log.Debug().Interface("composeFile", composeFile).Msg("Loaded compose file")

			// Create replicas of service if required inside serviceComposeFile
			err = internal.CreateServiceReplicas(&serviceComposeFile, count, serviceNames)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create service replicas")
				os.Exit(1)
			}

			log.Debug().Interface("serviceComposeFile", serviceComposeFile).Msg("Service compose file after creating replicas")

			// Merge serviceComposeFile into composeFile
			composeFile, err = internal.MergeComposeFiles(composeFile, serviceComposeFile)

			log.Debug().Interface("composeFile", composeFile).Msg("Compose file after merging")

			// Write the compose file to the output path
			log.Info().Msgf("Attempting to write the compose file to %s...", composeFilePath)
			err = internal.WriteComposeFile(&composeFile, composeFilePath)
			if err != nil {
				fmt.Printf("Error: Failed to write the compose file to %s: %v\n", composeFilePath, err)
				return
			}

			// Log success message
			log.Info().Msgf("Compose file successfully written to %s", composeFilePath)
		},
	}
}

// parseServiceNames parses the service names from the input arguments.
// If the number of service names is less than the count, it generates default service names.
func parseServiceNames(inputNames []string, count int) []string {
	serviceNames := make([]string, count)

	// Use provided names up to the count or the length of the inputNames
	for i := 0; i < count; i++ {
		if i < len(inputNames) {
			serviceNames[i] = inputNames[i]
		} else {
			// Generate default names if insufficient names are provided
			serviceNames[i] = fmt.Sprintf("service_%d", i+1)
		}
	}

	return serviceNames
}
