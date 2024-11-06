package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"kasmlink/pkg/dockercompose"
	"kasmlink/pkg/procedures"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// initCmd initializes the templates and/or Dockerfiles folders with default templates.
var initCmd = &cobra.Command{
	Use:   "init [type] [folder-path]",
	Short: "Initialize folders with default templates and Dockerfiles",
	Long: `Create folders with default templates or Dockerfiles, allowing customization or addition of new ones.
Specify 'type' as "templates", "dockerfiles", or "both".`,
	Args: cobra.ExactArgs(2),
	Run:  executeInitCmd,
}

func executeInitCmd(cmd *cobra.Command, args []string) {
	initType, folderPath := args[0], args[1]
	var err error

	switch initType {
	case "templates":
		err = initFolder("templates", folderPath, procedures.InitTemplatesFolder)
	case "dockerfiles":
		err = initFolder("dockerfiles", folderPath, procedures.InitDockerfilesFolder)
	case "both":
		errTemplates := initFolder("templates", folderPath, procedures.InitTemplatesFolder)
		errDockerfiles := initFolder("dockerfiles", folderPath, procedures.InitDockerfilesFolder)
		if errTemplates == nil && errDockerfiles == nil {
			log.Info().Str("folderPath", folderPath).Msg("Both templates and Dockerfiles initialized successfully")
		}
	default:
		log.Error().Str("type", initType).Msg("Invalid type specified. Use 'templates', 'dockerfiles', or 'both'")
	}
	if err != nil {
		log.Error().Err(err).Msg("Initialization failed")
	}
}

func initFolder(folderType, folderPath string, initFunc func(string) error) error {
	log.Info().Str("folderPath", folderPath).Msgf("Initializing %s", folderType)
	if err := initFunc(folderPath); err != nil {
		log.Error().Err(err).Msgf("Failed to initialize %s", folderType)
		return err
	}
	log.Info().Str("folderPath", folderPath).Msgf("%s initialized successfully", strings.Title(folderType))
	return nil
}

// generateCmd generates or updates a Docker Compose file with specified template instances.
var generateCmd = &cobra.Command{
	Use:   "generate [template-name] [count] [service-names] [output-path] [folder-path]",
	Short: "Generate or update a Docker Compose file with specified template instances",
	Args:  cobra.RangeArgs(3, 5),
	Run:   executeGenerateCmd,
}

func executeGenerateCmd(cmd *cobra.Command, args []string) {
	templateName, countStr, names := args[0], args[1], args[2]
	outputPath, folderPath := resolvePaths(args)

	// Ensure the output file has a .yaml extension
	if !strings.HasSuffix(outputPath, ".yaml") {
		outputPath += ".yaml"
	}

	serviceCount, err := strconv.Atoi(countStr)
	if err != nil || serviceCount <= 0 {
		log.Error().Msg("Invalid count provided. Please specify a positive integer.")
		return
	}

	serviceNames := parseServiceNames(names)
	composeFile, err := loadOrInitializeComposeFile(outputPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load or initialize Docker Compose file")
		return
	}

	// Populate compose file with template
	if err := procedures.PopulateComposeWithTemplate(composeFile, folderPath, templateName, serviceCount, serviceNames); err != nil {
		log.Error().Err(err).Msg("Failed to populate compose file")
		return
	}

	// Save the compose file
	if err := saveComposeFile(outputPath, composeFile); err != nil {
		log.Error().Err(err).Str("outputPath", outputPath).Msg("Failed to write updated Docker Compose file")
	} else {
		log.Info().Str("outputPath", outputPath).Msg("Docker Compose file updated successfully")
	}
}

func loadOrInitializeComposeFile(outputPath string) (*dockercompose.ComposeFile, error) {
	if _, err := os.Stat(outputPath); err == nil {
		return dockercompose.LoadComposeFile(outputPath)
	}

	return &dockercompose.ComposeFile{
		Version:  "3.8",
		Services: make(map[string]dockercompose.Service),
		Volumes:  make(map[string]dockercompose.Volume),
		Networks: make(map[string]dockercompose.Network),
	}, nil
}

func saveComposeFile(outputPath string, composeFile *dockercompose.ComposeFile) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputData, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal Docker Compose file: %w", err)
	}

	return os.WriteFile(outputPath, outputData, 0644)
}

func resolvePaths(args []string) (outputPath, folderPath string) {
	outputPath, folderPath = "./compose/docker-compose.yaml", "./templates"
	if len(args) > 3 {
		outputPath = args[3]
	}
	if len(args) > 4 {
		folderPath = args[4]
	}
	return
}

func parseServiceNames(names string) map[int]string {
	serviceNames := make(map[int]string)
	nameParts := strings.Split(names, ",")
	for i, name := range nameParts {
		serviceNames[i+1] = strings.TrimSpace(name)
	}
	return serviceNames
}

func init() {
	// Register init and generate commands
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(generateCmd)
}
