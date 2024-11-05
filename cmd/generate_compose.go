package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/procedures"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercompose"
)

// initCmd initializes the templates and/or Dockerfiles folders with default templates.
var initCmd = &cobra.Command{
	Use:   "init [type] [folder-path]",
	Short: "Initialize folders with default templates and Dockerfiles",
	Long: `Create folders with default templates or Dockerfiles, allowing you to customize or add new ones as needed.
			The 'type' argument specifies what to initialize: "templates", "dockerfiles", or "both".`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		initType := args[0]
		folderPath := args[1]

		var err error

		// Determine what to initialize based on the type argument
		switch initType {
		case "templates":
			log.Info().Str("folderPath", folderPath).Msg("Initializing templates")
			err = procedures.InitTemplatesFolder(folderPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to initialize templates")
				return
			}
			log.Info().Str("folderPath", folderPath).Msg("Templates initialized successfully")

		case "dockerfiles":
			log.Info().Str("folderPath", folderPath).Msg("Initializing Dockerfiles")
			err = procedures.InitDockerfilesFolder(folderPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to initialize Dockerfiles")
				return
			}
			log.Info().Str("folderPath", folderPath).Msg("Dockerfiles initialized successfully")

		case "both":
			log.Info().Str("folderPath", folderPath).Msg("Initializing both templates and Dockerfiles")
			errTemplates := procedures.InitTemplatesFolder(folderPath)
			errDockerfiles := procedures.InitDockerfilesFolder(folderPath)
			if errTemplates != nil {
				log.Error().Err(errTemplates).Msg("Failed to initialize templates")
			}
			if errDockerfiles != nil {
				log.Error().Err(errDockerfiles).Msg("Failed to initialize Dockerfiles")
			}
			if errTemplates == nil && errDockerfiles == nil {
				log.Info().Str("folderPath", folderPath).Msg("Templates and Dockerfiles initialized successfully")
			}

		default:
			log.Error().Str("type", initType).Msg("Invalid type specified. Use 'templates', 'dockerfiles', or 'both'.")
		}
	},
}

// generateCmd generates or updates a Docker Compose file with multiple instances of a specified template.
var generateCmd = &cobra.Command{
	Use:   "generate [template-name] [count] [service-names] [output-path] [folder-path]",
	Short: "Generate or update a Docker Compose file with specified template instances",
	Long:  `Generate or update a Docker Compose file with multiple instances of a specified template, using either a base name or unique names for each service instance.`,
	Args:  cobra.RangeArgs(3, 5),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		count := args[1]
		names := args[2]
		outputPath := "./templates/services" // Default output directory
		folderPath := "./templates"          // Default folder path for templates

		// Optional arguments
		if len(args) > 3 {
			outputPath = args[3]
		}
		if len(args) > 4 {
			folderPath = args[4]
		}

		// Parse the count argument
		serviceCount, err := strconv.Atoi(count)
		if err != nil || serviceCount <= 0 {
			log.Error().Err(err).Msg("Invalid count provided. Please specify a positive integer")
			return
		}

		// Parse service names
		serviceNames := make(map[int]string)
		nameParts := strings.Split(names, ",")
		if len(nameParts) == 1 {
			serviceNames[1] = nameParts[0]
		} else {
			for i, name := range nameParts {
				serviceNames[i+1] = name
			}
		}
		log.Info().
			Str("templateName", templateName).
			Int("serviceCount", serviceCount).
			Str("outputPath", outputPath).
			Str("folderPath", folderPath).
			Msg("Generating Docker Compose file")

		// Declare composeFile as a pointer type
		var composeFile *dockercompose.ComposeFile

		if _, err := os.Stat(outputPath); err == nil {
			composeFile, err = dockercompose.LoadComposeFile(outputPath)
			if err != nil {
				log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to load existing Docker Compose file")
				return
			}
			log.Info().Str("outputPath", outputPath).Msg("Updating existing Docker Compose file")
		} else {
			// Initialize a new ComposeFile if the file does not exist
			composeFile = &dockercompose.ComposeFile{}
			log.Info().Str("outputPath", outputPath).Msg("Creating new Docker Compose file")
		}

		// Populate the ComposeFile with the specified template and instances
		err = procedures.PopulateComposeWithTemplate(composeFile, folderPath, templateName, serviceCount, serviceNames)
		if err != nil {
			log.Error().Err(err).Msg("Failed to populate compose file")
			return
		}

		// Determine if outputPath is a directory and set default file name
		if stat, err := os.Stat(outputPath); err == nil && stat.IsDir() {
			outputPath = filepath.Join(outputPath, "docker-compose.yaml")
		} else if !strings.HasSuffix(outputPath, ".yaml") {
			outputPath = filepath.Join(outputPath, "docker-compose.yaml")
		}

		// Create directory if it does not exist
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			log.Error().Err(err).Str("directory", filepath.Dir(outputPath)).Msg("Failed to create output directory")
			return
		}

		// Handle existing file names with suffixes
		base := strings.TrimSuffix(outputPath, ".yaml")
		finalPath := outputPath
		for i := 1; ; i++ {
			if _, err := os.Stat(finalPath); os.IsNotExist(err) {
				break
			}
			finalPath = fmt.Sprintf("%s-%d.yaml", base, i)
		}

		// Load the Docker Compose template
		tmpl, err := dockercompose.LoadEmbeddedTemplate()
		if err != nil {
			log.Error().Err(err).Msg("Failed to load Docker Compose template")
			return
		}

		// Generate or update the Docker Compose file at finalPath
		err = dockercompose.GenerateDockerComposeFile(tmpl, *composeFile, finalPath) // Dereference the pointer here
		if err != nil {
			log.Error().Err(err).Str("finalPath", finalPath).Msg("Failed to generate Docker Compose file")
			return
		}

		log.Info().Str("finalPath", finalPath).Msg("Docker Compose file generated successfully")
	},
}

func init() {
	// Register init and generate commands
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(generateCmd)
}
