package cmd

import (
	"fmt"
	"kasmlink/pkg/procedures"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercompose"
)

// initCmd initializes the templates folder with default templates.
var initCmd = &cobra.Command{
	Use:   "init [folder-path]",
	Short: "Initialize the templates folder with default templates",
	Long:  `Create a folder with default templates, allowing you to customize or add new ones as needed.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderPath := args[0]
		err := procedures.InitTemplatesFolder(folderPath)
		if err != nil {
			log.Fatalf("Failed to initialize templates: %v", err)
		} else {
			fmt.Printf("Templates initialized successfully in folder: %s\n", folderPath)
		}
	},
}

// generateCmd generates a Docker Compose file with multiple instances of a specified template.
var generateCmd = &cobra.Command{
	Use:   "generate [template-name] [count] [service-names] [output-path] [folder-path]",
	Short: "Generate a Docker Compose file with specified template instances",
	Long:  `Generate a Docker Compose file with multiple instances of a specified template, using either a base name or unique names for each service instance.`,
	Args:  cobra.RangeArgs(3, 5),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		count := args[1]
		names := args[2]
		outputPath := "./compose" // Default output directory
		folderPath := "templates" // Default folder path for templates

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
			log.Fatalf("Invalid count provided. Please specify a positive integer: %v", err)
		}

		// Parse service names
		serviceNames := make(map[int]string)
		nameParts := strings.Split(names, ",")
		if len(nameParts) == 1 {
			// If only one name is provided, treat it as a base name
			serviceNames[1] = nameParts[0]
		} else {
			// Use the names directly for each service instance
			for i, name := range nameParts {
				serviceNames[i+1] = name
			}
		}

		// Load the ComposeFile struct
		var composeFile dockercompose.ComposeFile

		// Populate the ComposeFile with the specified template and instances
		err = procedures.PopulateComposeWithTemplate(&composeFile, folderPath, templateName, serviceCount, serviceNames)
		if err != nil {
			log.Fatalf("Failed to populate compose file: %v", err)
		}

		// Determine if outputPath is a directory and set default file name
		if stat, err := os.Stat(outputPath); err == nil && stat.IsDir() {
			outputPath = filepath.Join(outputPath, "docker-compose.yaml")
		} else if !strings.HasSuffix(outputPath, ".yaml") {
			// If outputPath does not end with .yaml, append default filename
			outputPath = filepath.Join(outputPath, "docker-compose.yaml")
		}

		// Create directory if it does not exist
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
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
			log.Fatalf("Failed to load Docker Compose template: %v", err)
		}

		// Generate the Docker Compose file at finalPath
		err = dockercompose.GenerateDockerComposeFile(tmpl, composeFile, finalPath)
		if err != nil {
			log.Fatalf("Failed to generate Docker Compose file: %v", err)
		}

		fmt.Printf("Docker Compose file generated successfully at %s\n", finalPath)
	},
}

func init() {
	// Register init and generate commands
	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(generateCmd)
}
