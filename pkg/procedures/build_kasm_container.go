package procedures

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"io"
	"io/fs"
	embedfiles "kasmlink/embedded"
	"os"
	"path/filepath"
)

// buildCoreImageKasm builds a Docker image using the embedded Dockerfile and build context.
// This function orchestrates all the other functions to provide a complete image build process.
func buildCoreImageKasm(imageTag string) error {
	// Define Dockerfile path within the embedded context
	dockerfilePath := "workspace-core-image/dockerfile-kasm-core-suse"

	log.Info().Str("dockerfilePath", dockerfilePath).Str("imageTag", imageTag).Msg("Starting to build the specific Docker image")

	// Create a tar archive from the embedded build context
	buildContextTar, err := createTarFromEmbeddedContext("embedded")
	if err != nil {
		log.Error().Err(err).Msg("Failed to create build context tar from embedded files")
		return fmt.Errorf("could not create build context tar: %v", err)
	}
	log.Debug().Msg("Build context tar created successfully")

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Docker client")
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Build the Docker image
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: dockerfilePath,
		Remove:     true,
	}

	imageBuildResponse, err := cli.ImageBuild(context.Background(), buildContextTar, buildOptions)
	if err != nil {
		log.Error().Err(err).Str("imageTag", imageTag).Msg("Failed to build Docker image")
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer imageBuildResponse.Body.Close()

	// Print the build logs
	err = printBuildLogs(imageBuildResponse.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read build response")
		return fmt.Errorf("failed to read build response: %v", err)
	}

	log.Info().Str("imageTag", imageTag).Msg("Docker image built successfully")
	return nil
}

// createTarFromEmbeddedContext creates a tarball from the embedded build context files.
func createTarFromEmbeddedContext(embedDir string) (io.Reader, error) {
	log.Debug().Str("embedDir", embedDir).Msg("Creating tar with embedded build context")
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Walk the embedded filesystem and add each file to the tarball
	err := fs.WalkDir(embedfiles.DockerFilesKasm, embedDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("filePath", filePath).Msg("Error accessing file during tar creation")
			return err
		}

		// Skip directories as they do not need to be added to the tar
		if d.IsDir() {
			return nil
		}

		// Open the file from the embedded filesystem
		file, err := embedfiles.DockerFilesKasm.Open(filePath)
		if err != nil {
			return fmt.Errorf("could not open embedded file: %v", err)
		}
		defer file.Close()

		// Get file information
		fileInfo, err := d.Info()
		if err != nil {
			return fmt.Errorf("could not get file info: %v", err)
		}

		// Create the tar header for the file
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return fmt.Errorf("could not create tar header: %v", err)
		}

		// Set the correct name in the tar header
		header.Name = filepath.ToSlash(filePath)

		log.Debug().Str("file", header.Name).Msg("Adding file to build context tar")

		// Write the header to the tarball
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("could not write header to tar: %v", err)
		}

		// Copy the file contents to the tarball
		_, err = io.Copy(tw, file)
		file.Close() // Explicitly close to avoid too many open files
		if err != nil {
			return fmt.Errorf("could not copy file contents to tar: %v", err)
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create tar of build context")
		return nil, fmt.Errorf("could not tar build context: %v", err)
	}

	// Close the tar writer
	if err := tw.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close tar writer")
		return nil, fmt.Errorf("could not close tar writer: %v", err)
	}

	log.Debug().Msg("Build context tar creation completed successfully")
	return buf, nil
}

// deployCoreImage exports the Docker image to a tar file and copies it to a user-specified node.
func deployCoreImage(imageTag, targetNodePath string) error {
	log.Info().Str("imageTag", imageTag).Msg("Starting export of Docker image to tar file")

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Docker client")
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Export the Docker image to a tar file
	imageReader, err := cli.ImageSave(context.Background(), []string{imageTag})
	if err != nil {
		log.Error().Err(err).Str("imageTag", imageTag).Msg("Failed to export Docker image to tar")
		return fmt.Errorf("failed to export Docker image: %v", err)
	}
	defer imageReader.Close()

	tarFilePath := filepath.Join(os.TempDir(), "core-image.tar")
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		log.Error().Err(err).Str("tarFilePath", tarFilePath).Msg("Failed to create tar file")
		return fmt.Errorf("could not create tar file: %v", err)
	}
	defer tarFile.Close()

	_, err = io.Copy(tarFile, imageReader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write Docker image to tar file")
		return fmt.Errorf("failed to write Docker image to tar file: %v", err)
	}

	log.Info().Str("tarFilePath", tarFilePath).Msg("Docker image exported to tar file successfully")

	// Copy the tar file to the specified node path
	err = copyTarToNode(tarFilePath, targetNodePath)
	if err != nil {
		log.Error().Err(err).Str("targetNodePath", targetNodePath).Msg("Failed to copy tar file to target node")
		return fmt.Errorf("failed to copy tar file to target node: %v", err)
	}

	log.Info().Str("targetNodePath", targetNodePath).Msg("Docker image tar file copied to target node successfully")
	return nil
}

// copyTarToNode copies the tar file to the specified node.
func copyTarToNode(tarFilePath, targetNodePath string) error {
	log.Info().Str("tarFilePath", tarFilePath).Str("targetNodePath", targetNodePath).Msg("Starting to copy tar file to target node")

	// Open the tar file
	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		log.Error().Err(err).Str("tarFilePath", tarFilePath).Msg("Failed to open tar file")
		return fmt.Errorf("could not open tar file: %v", err)
	}
	defer tarFile.Close()

	// Create the target file on the specified node path
	targetFile, err := os.Create(targetNodePath)
	if err != nil {
		log.Error().Err(err).Str("targetNodePath", targetNodePath).Msg("Failed to create target file on node")
		return fmt.Errorf("could not create target file: %v", err)
	}
	defer targetFile.Close()

	// Copy the tar file contents to the target file
	_, err = io.Copy(targetFile, tarFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy tar file contents to target node")
		return fmt.Errorf("failed to copy tar file contents to target node: %v", err)
	}

	log.Info().Str("targetNodePath", targetNodePath).Msg("Tar file copied to target node successfully")
	return nil
}

// importImageToNodeSSH imports the Docker image on the target node using SSH by running docker load -i command.
func importImageToNodeSSH(targetNodeAddress, targetNodePath, sshUser, sshPassword string) error {
	log.Info().Str("targetNodeAddress", targetNodeAddress).Str("targetNodePath", targetNodePath).Msg("Starting to import Docker image on target node via SSH")

	// Set up SSH client configuration
	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the target node
	sshClient, err := ssh.Dial("tcp", targetNodeAddress, config)
	if err != nil {
		return fmt.Errorf("failed to connect to target node: %v", err)
	}
	defer sshClient.Close()

	// Open an SSH session
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Run the Docker load command to import the image
	cmd := fmt.Sprintf("docker load -i %s", targetNodePath)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to run docker load on target node: %v", err)
	}

	log.Info().Str("targetNodePath", targetNodePath).Msg("Docker image imported successfully on target node")
	return nil
}

// DeployDockerImage is a wrapper function that builds a Docker image, exports it to a tar file, copies it to the target node, and loads it on that node.
// The user can specify the required arguments to customize the entire process.
func DeployDockerImage(imageTag, dockerfilePath, targetNodeAddress, targetNodePath, sshUser, sshPassword string) error {
	// Step 1: Build the Docker image
	err := buildCoreImageKasm(imageTag)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}

	// Step 2: Export the Docker image to a tar file and copy it to the target node
	err = deployCoreImage(imageTag, targetNodePath)
	if err != nil {
		return fmt.Errorf("failed to deploy Docker image tar to the target node: %v", err)
	}

	// Step 3: Load the Docker image on the target node via SSH
	err = importImageToNodeSSH(targetNodeAddress, targetNodePath, sshUser, sshPassword)
	if err != nil {
		return fmt.Errorf("failed to load Docker image on target node: %v", err)
	}

	// Cleanup the tar file from the local machine
	tmpTarPath := fmt.Sprintf("%s/core-image.tar", os.TempDir())
	defer os.Remove(tmpTarPath)

	log.Info().Msg("Docker image deployed and imported successfully on target node")
	return nil
}
