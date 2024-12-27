package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"kasmlink/internal"
	"kasmlink/pkg/dockercompose"
)

// Init initializes the root command and its subcommands.
func init() {
	// Define "compose" command
	composeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Manage Docker Compose files",
		Long:  `Commands to manage and generate Docker Compose files, including creating new ones, loading from YAML, and using embedded templates.`,
	}

	// Add subcommands for Docker Compose management
	composeCmd.AddCommand(createPopulateComposeWithTemplateCommand())

	// Add "compose" to the root command
	RootCmd.AddCommand(composeCmd)
}

// createPopulateComposeWithTemplateCommand generates or populates a Docker Compose file.
func createPopulateComposeWithTemplateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [composeFilePath] [templateFolderPath] [templateName] [count] [serviceNames...]",
		Short: "Generate or populate a Docker Compose file",
		Long: `This command generates or populates a Docker Compose file with instances of a specified template.
You need to provide the compose file path, the folder path containing the template YAML, the template name, 
the count of service instances to create, and optional service names for serialization.`,
		Args: cobra.MinimumNArgs(4),
		Run:  runPopulateComposeWithTemplate,
	}
}

// runPopulateComposeWithTemplate executes the generation/population of a Docker Compose file.
func runPopulateComposeWithTemplate(cmd *cobra.Command, args []string) {
	composeFilePath := args[0]
	templateFolderPath := args[1]
	templateName := args[2]

	// Parse and validate count argument
	count, err := strconv.Atoi(args[3])
	if err != nil {
		log.Error().Err(err).Msg("Invalid count value provided")
		os.Exit(1)
	}

	// Parse optional service names
	serviceNames := parseServiceNames(args[4:], count)

	// Construct service file path
	serviceFilePath := fmt.Sprintf("%s/%s.yaml", templateFolderPath, templateName)

	// Load the service Compose file
	serviceComposeFile, err := loadComposeFile(serviceFilePath)
	if err != nil {
		log.Error().Err(err).Str("file_path", serviceFilePath).Msg("Failed to load service Compose file")
		os.Exit(1)
	}

	log.Debug().Interface("serviceComposeFile", serviceComposeFile).Msg("Loaded service Compose file")

	// Load or create the output Compose file
	composeFile, err := loadOrCreateComposeFile(composeFilePath)
	if err != nil {
		log.Error().Err(err).Str("file_path", composeFilePath).Msg("Failed to load or create Compose file")
		os.Exit(1)
	}

	log.Debug().Interface("composeFile", composeFile).Msg("Compose file loaded/created")

	// Create replicas of the service within the serviceComposeFile
	err = internal.CreateServiceReplicas(&serviceComposeFile, count, serviceNames)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create service replicas")
		os.Exit(1)
	}

	log.Debug().Interface("serviceComposeFile", serviceComposeFile).Msg("Service Compose file after creating replicas")

	// Merge the serviceComposeFile into the output Compose file
	mergedComposeFile, err := internal.MergeComposeFiles(composeFile, serviceComposeFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to merge Compose files")
		os.Exit(1)
	}

	log.Debug().Interface("mergedComposeFile", mergedComposeFile).Msg("Compose file after merging")

	// Write the merged Compose file to the output path
	log.Info().Str("path", composeFilePath).Msg("Writing the merged Compose file")
	err = internal.WriteComposeFile(&mergedComposeFile, composeFilePath)
	if err != nil {
		log.Error().Err(err).Str("path", composeFilePath).Msg("Failed to write the Compose file")
		os.Exit(1)
	}

	// Log success message
	log.Info().Str("path", composeFilePath).Msg("Compose file successfully written")
}

// parseServiceNames parses or generates service names for the given count.
func parseServiceNames(inputNames []string, count int) []string {
	serviceNames := make([]string, count)

	for i := 0; i < count; i++ {
		if i < len(inputNames) {
			serviceNames[i] = inputNames[i]
		} else {
			serviceNames[i] = fmt.Sprintf("service_%d", i+1)
		}
	}

	return serviceNames
}

// loadComposeFile loads a Docker Compose file from the given path.
func loadComposeFile(filePath string) (dockercompose.DockerCompose, error) {
	log.Info().Str("file_path", filePath).Msg("Loading Compose file")
	composeFilePtr, err := dockercompose.LoadAndParseComposeFile(filePath)
	if err != nil {
		return dockercompose.DockerCompose{}, fmt.Errorf("failed to load Compose file: %w", err)
	}
	return composeFilePtr, nil
}

// loadOrCreateComposeFile loads an existing Compose file or creates a new one if the file doesn't exist.
func loadOrCreateComposeFile(filePath string) (dockercompose.DockerCompose, error) {
	composeFile, err := loadComposeFile(filePath)
	if err == nil {
		return composeFile, nil
	}

	// If the file doesn't exist, create a new Compose file with defaults
	log.Warn().Str("file_path", filePath).Msg("Compose file does not exist; creating a new one")
	return dockercompose.DockerCompose{
		Version:  "3.8",
		Services: make(map[string]dockercompose.ServiceDefinition),
		Networks: make(map[string]dockercompose.NetworkDefinition),
		Volumes:  make(map[string]dockercompose.VolumeDefinition),
		Configs:  make(map[string]dockercompose.ConfigDefinition),
		Secrets:  make(map[string]dockercompose.SecretDefinition),
	}, nil
}
