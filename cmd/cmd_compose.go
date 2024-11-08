package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kasmlink/pkg/dockercompose"
)

// Init initializes the root command.
func init() {
	// Define "compose" command
	composeCmd := &cobra.Command{
		Use:   "compose",
		Short: "Manage Docker Compose files",
		Long:  `Commands to manage and generate Docker Compose files, including creating new ones, loading from YAML, and using embedded templates.`,
	}

	// Add subcommands for generating, loading, and embedding Docker Compose files
	composeCmd.AddCommand(
		createGenerateDockerComposeCommand(),
		createLoadDockerComposeFileCommand(),
		createLoadEmbeddedTemplateCommand(),
	)

	// Add "compose" to the root command
	RootCmd.AddCommand(composeCmd)
}

// createGenerateDockerComposeCommand creates the command for generating a Docker Compose file.
func createGenerateDockerComposeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [templatePath] [composeFilePath] [outputPath]",
		Short: "Generate a Docker Compose file",
		Long:  "Generate a Docker Compose file based on a specified template and structure.",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			templatePath := args[0]
			composeFilePath := args[1]
			outputPath := args[2]

			// Load the ComposeFile from the provided path
			composeFile, err := dockercompose.LoadComposeFile(composeFilePath)
			if err != nil {
				HandleError(err)
				return
			}

			// Load the template
			tmpl, err := dockercompose.LoadEmbeddedTemplate(templatePath)
			if err != nil {
				HandleError(err)
				return
			}

			// Generate the Docker Compose file
			err = dockercompose.GenerateDockerComposeFile(tmpl, *composeFile, outputPath)
			HandleError(err)
		},
	}
}

// createLoadDockerComposeFileCommand creates the command for loading a Docker Compose file.
func createLoadDockerComposeFileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "load [composeFilePath]",
		Short: "Load a Docker Compose configuration file",
		Long:  "Load and parse a Docker Compose configuration file from a given path.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			composeFilePath := args[0]

			// Load the Docker Compose configuration
			composeFile, err := dockercompose.LoadComposeFile(composeFilePath)
			if err != nil {
				HandleError(err)
				return
			}

			// Display a summary of the loaded configuration
			fmt.Printf("Successfully loaded Docker Compose configuration from %s\n", composeFilePath)
			fmt.Printf("Services: %+v\n", composeFile.Services)
		},
	}
}

// createLoadEmbeddedTemplateCommand creates the command for loading an embedded template.
func createLoadEmbeddedTemplateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "load-template [templatePath]",
		Short: "Load an embedded Docker Compose template",
		Long:  "Load a Docker Compose template from the embedded files.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			templatePath := args[0]

			// Load the embedded template
			tmpl, err := dockercompose.LoadEmbeddedTemplate(templatePath)
			if err != nil {
				HandleError(err)
				return
			}

			// If successfully loaded, print a success message
			if tmpl != nil {
				fmt.Printf("Successfully loaded embedded template from %s\n", templatePath)
			}
		},
	}
}
