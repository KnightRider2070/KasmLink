package dockerutils

import (
	"archive/tar"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	retryCount    = 3
	retryDelay    = 2 * time.Second
	tarBufferSize = 10 * 1024 * 1024 // 10MB buffer size for tar creation
)

// CreateTarWithContext creates a tar archive of the specified build context directory.
func CreateTarWithContext(buildContextDir string) (io.Reader, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, tarBufferSize))
	tarWriter := tar.NewWriter(buffer)
	defer tarWriter.Close()

	err := filepath.Walk(buildContextDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name, _ = filepath.Rel(buildContextDir, path)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Str("buildContextDir", buildContextDir).Msg("Failed to create tar archive")
		return nil, fmt.Errorf("failed to create tar archive: %v", err)
	}
	return io.NopCloser(buffer), nil
}

// CopyEmbeddedFiles copies files from the embedded filesystem to a target directory.
func CopyEmbeddedFiles(fs embed.FS, srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking through source directory")
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Determine the destination path and relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			log.Error().Err(err).Str("srcDir", srcDir).Str("path", path).Msg("Error getting relative path")
			return err
		}
		dstPath := filepath.Join(dstDir, relPath)

		// Read the file from the embedded filesystem
		content, err := fs.ReadFile(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error reading file from embedded filesystem")
			return err
		}

		// Ensure the destination directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			log.Error().Err(err).Str("dstPath", dstPath).Msg("Failed to create destination directory")
			return err
		}

		// Write the file content to the destination path
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			log.Error().Err(err).Str("dstPath", dstPath).Msg("Failed to write file to destination path")
			return err
		}

		log.Debug().Str("dstPath", dstPath).Msg("File copied to destination successfully")
		return nil
	})
}

// CreateTarFromEmbedded creates a tar archive from an embedded filesystem directory
func CreateTarFromEmbedded(embedFS embed.FS, srcDir string) (io.Reader, error) {
	log.Debug().Str("srcDir", srcDir).Msg("Creating tar archive from embedded files")
	buf := bytes.NewBuffer(make([]byte, 0, tarBufferSize))
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Walk through the embedded filesystem and add each file to the tar archive
	err := fs.WalkDir(embedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error accessing embedded file")
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Open the file from the embedded filesystem
		file, err := embedFS.Open(path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not open embedded file")
			return fmt.Errorf("could not open embedded file: %v", err)
		}
		defer file.Close()

		// Get file info for the tar header
		info, err := d.Info()
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not retrieve file info")
			return fmt.Errorf("could not retrieve file info: %v", err)
		}

		// Create a tar header from the file info
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not create tar header")
			return fmt.Errorf("could not create tar header: %v", err)
		}

		// Set the header name to the relative path in the archive
		header.Name = filepath.ToSlash(path[len(srcDir)+1:]) // Remove srcDir prefix and use slashes

		// Write the header to the tar writer
		if err := tw.WriteHeader(header); err != nil {
			log.Error().Err(err).Str("header", header.Name).Msg("Could not write tar header")
			return fmt.Errorf("could not write tar header: %v", err)
		}

		// Copy the file contents to the tar archive
		if _, err := io.Copy(tw, file); err != nil {
			log.Error().Err(err).Str("path", path).Msg("Could not copy file contents to tar")
			return fmt.Errorf("could not copy file contents to tar: %v", err)
		}

		log.Debug().Str("file", header.Name).Msg("Added file to tar archive")
		return nil
	})
	if err != nil {
		log.Error().Err(err).Str("srcDir", srcDir).Msg("Failed to create tar archive from embedded files")
		return nil, fmt.Errorf("failed to create tar archive: %v", err)
	}

	return io.NopCloser(buf), nil
}

// ExportImageToTar exports a Docker image to a tar file and returns the file path.
func ExportImageToTar(cli *client.Client, imageTag string) (string, error) {
	log.Info().Str("imageTag", imageTag).Msg("Exporting Docker image to tar file")

	// Save the Docker image as a tar stream
	imageReader, err := cli.ImageSave(context.Background(), []string{imageTag})
	if err != nil {
		log.Error().Err(err).Str("imageTag", imageTag).Msg("Failed to export Docker image")
		return "", fmt.Errorf("failed to export Docker image: %v", err)
	}
	defer imageReader.Close()

	// Define the tar file path in a temporary directory
	tarFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-image.tar", imageTag))

	// Create the tar file
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		log.Error().Err(err).Str("tarFilePath", tarFilePath).Msg("Could not create tar file")
		return "", fmt.Errorf("could not create tar file: %v", err)
	}
	defer tarFile.Close()

	// Copy the image stream to the tar file
	if _, err = io.Copy(tarFile, imageReader); err != nil {
		log.Error().Err(err).Str("tarFilePath", tarFilePath).Msg("Failed to write Docker image to tar file")
		return "", fmt.Errorf("failed to write Docker image to tar file: %v", err)
	}

	log.Info().Str("tarFilePath", tarFilePath).Msg("Docker image exported to tar file successfully")
	return tarFilePath, nil
}

// PrintBuildLogs reads the Docker build output and formats it for better readability
func PrintBuildLogs(reader io.Reader) error {
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
			log.Error().Err(err).Msg("Error decoding Docker build logs")
			return fmt.Errorf("error decoding build logs: %v", err)
		}

		// Check for errors in the Docker build logs
		if msg.Error != "" {
			log.Error().Str("error", msg.Error).Msg("Docker build log error")
			fmt.Println(errorColor(fmt.Sprintf("Error: %s", msg.Error)))
		}

		// Print the normal stream of logs
		if msg.Stream != "" {
			log.Debug().Msgf("Docker build log: %s", msg.Stream)
			fmt.Print(successColor(msg.Stream))
		}
	}

	log.Info().Msg("Docker build process completed successfully")
	return nil
}

// BuildDockerImage builds a Docker image from a specified tar build context.
func BuildDockerImage(cli *client.Client, imageTag, dockerfilePath string, buildContext io.Reader, buildArgs map[string]*string) error {
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: dockerfilePath,
		Remove:     true,
		BuildArgs:  buildArgs,
	}

	log.Info().Str("imageTag", imageTag).Str("dockerfilePath", dockerfilePath).Msg("Starting Docker image build")

	imageBuildResponse, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		log.Error().Err(err).Str("imageTag", imageTag).Msg("Failed to build Docker image")
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer imageBuildResponse.Body.Close()

	if err := PrintBuildLogs(imageBuildResponse.Body); err != nil {
		log.Error().Err(err).Str("imageTag", imageTag).Msg("Error occurred during Docker build logs processing")
		return fmt.Errorf("error occurred during Docker build logs processing: %v", err)
	}

	log.Info().Str("imageTag", imageTag).Msg("Docker image built successfully")
	return nil
}
