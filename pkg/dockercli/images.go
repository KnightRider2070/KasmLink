package dockercli

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io"
	"os"
	"os/exec"
	"strings"
)

// PullImage pulls a Docker image from a registry.
func PullImage(imageName string) error {
	cmd := exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull image: %v - %s", err, string(output))
	}
	fmt.Printf("Image %s pulled successfully.\n", imageName)
	return nil
}

// PushImage pushes a Docker image to a registry.
func PushImage(imageName string) error {
	cmd := exec.Command("docker", "push", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push image: %v - %s", err, string(output))
	}
	fmt.Printf("Image %s pushed successfully.\n", imageName)
	return nil
}

// RemoveImage removes a Docker image by name or ID.
func RemoveImage(imageName string) error {
	cmd := exec.Command("docker", "rmi", imageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove image: %v - %s", err, string(output))
	}
	fmt.Printf("Image %s removed successfully.\n", imageName)
	return nil
}

// ListImages lists all Docker images on the host.
func ListImages() ([]string, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker images: %v - %s", err, string(output))
	}

	// Split the output into lines and return the image names (with tags)
	images := strings.Split(strings.TrimSpace(string(output)), "\n")
	return images, nil
}

// UpdateAllImages pulls the latest version of all present Docker images.
func UpdateAllImages() error {
	// Get the list of all images
	images, err := ListImages()
	if err != nil {
		return fmt.Errorf("failed to list Docker images: %v", err)
	}

	// Iterate over each image and pull the latest version
	for _, image := range images {
		fmt.Printf("Updating image: %s\n", image)
		err := PullImage(image)
		if err != nil {
			return fmt.Errorf("failed to update image %s: %v", image, err)
		}
	}

	fmt.Println("All images have been updated successfully.")
	return nil
}

// ExportImageToTar exports a Docker image to a tar file for later import
func ExportImageToTar(imageName, outputFile string) error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Fetch the image and save it as a tar archive
	fmt.Printf("Saving Docker image %s to tar file %s...\n", imageName, outputFile)
	imageReader, err := cli.ImageSave(context.Background(), []string{imageName})
	if err != nil {
		return fmt.Errorf("could not save Docker image: %v", err)
	}
	defer imageReader.Close()

	// Check if the image exists and was fetched
	if imageReader == nil {
		return fmt.Errorf("image reader is nil, possibly failed to find image: %s", imageName)
	}

	// Create or open the output tar file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("could not create tar file: %v", err)
	}
	defer outFile.Close()

	// Copy the image tarball from the Docker API to the file
	written, err := io.Copy(outFile, imageReader)
	if err != nil {
		return fmt.Errorf("could not write image to tar file: %v", err)
	}

	fmt.Printf("Successfully saved %d bytes of image %s to %s\n", written, imageName, outputFile)
	return nil
}

// ImportImageFromTar imports a Docker image from a tar file
func ImportImageFromTar(tarFilePath string) error {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create Docker client: %v", err)
	}

	// Open the tar file that contains the Docker image
	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		return fmt.Errorf("could not open tar file: %v", err)
	}
	defer tarFile.Close()

	// Load the image into Docker from the tar file
	imageLoadResponse, err := cli.ImageLoad(context.Background(), tarFile, true)
	if err != nil {
		return fmt.Errorf("could not load Docker image from tar: %v", err)
	}
	defer imageLoadResponse.Body.Close()

	// Print the response
	_, err = io.Copy(os.Stdout, imageLoadResponse.Body)
	if err != nil {
		return fmt.Errorf("could not read image load response: %v", err)
	}

	fmt.Printf("Docker image successfully loaded from %s\n", tarFilePath)
	return nil
}
