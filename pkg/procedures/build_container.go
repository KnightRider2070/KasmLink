package procedures

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"io"
	"os"
	"path/filepath"
)

// BuildContainer builds a Docker image from a Dockerfile located in the specified build context directory.
// It also tags the image with the provided tag name.
func BuildContainer(buildContextDir, dockerfilePath, imageTag string) error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Create a tar archive of the build context directory
	buildContextTar, err := createTarWithBuildContext(buildContextDir)
	if err != nil {
		return fmt.Errorf("could not create build context tar: %v", err)
	}

	// Build the Docker image
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag}, // Tag the image with the provided tag
		Dockerfile: dockerfilePath,     // Use the Dockerfile as the build file
		Remove:     true,               // Remove intermediate containers after build
	}

	imageBuildResponse, err := cli.ImageBuild(context.Background(), buildContextTar, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer imageBuildResponse.Body.Close()

	// Print the build logs
	err = printBuildLogs(imageBuildResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read build response: %v", err)
	}

	return nil
}

// createTarWithBuildContext creates a tarball containing the build context (the directory with Dockerfile and other files)
func createTarWithBuildContext(buildContextDir string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Walk the build context directory and add each file to the tarball
	err := filepath.Walk(buildContextDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the header based on the file info
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return fmt.Errorf("could not create tar header: %v", err)
		}

		// Use relative path for the file inside the tarball
		// Normalize paths by replacing backslashes with forward slashes
		header.Name, err = filepath.Rel(buildContextDir, file)
		if err != nil {
			return fmt.Errorf("could not get relative path: %v", err)
		}

		// Normalize path separators to forward slashes for Docker compatibility
		header.Name = filepath.ToSlash(header.Name)

		// Print the file being added to the tarball for debugging
		fmt.Println("Adding file to build context:", header.Name)

		// Write the header to the tarball
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("could not write header to tar: %v", err)
		}

		// If the file is a directory, no need to copy contents
		if fi.IsDir() {
			return nil
		}

		// Open the file to copy its contents to the tarball
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("could not open file: %v", err)
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("could not copy file contents to tar: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not tar build context: %v", err)
	}

	// Close the tar writer
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("could not close tar writer: %v", err)
	}

	return buf, nil
}

// printBuildLogs reads the Docker build output and formats it for better readability
func printBuildLogs(reader io.Reader) error {
	decoder := json.NewDecoder(reader)

	// Define a structure to parse Docker's build log JSON
	var msg struct {
		Stream string `json:"stream"`
		Error  string `json:"error"`
	}

	// Use color formatting for success and error logs
	successColor := color.New(color.FgGreen).SprintFunc()
	errorColor := color.New(color.FgRed).SprintFunc()

	// Read each line from the Docker build output and print it
	for decoder.More() {
		if err := decoder.Decode(&msg); err != nil {
			return fmt.Errorf("error decoding build logs: %v", err)
		}

		// Check for errors in the Docker build logs
		if msg.Error != "" {
			fmt.Println(errorColor(fmt.Sprintf("Error: %s", msg.Error)))
		}

		// Print the normal stream of logs
		if msg.Stream != "" {
			fmt.Print(successColor(msg.Stream))
		}
	}

	return nil
}
