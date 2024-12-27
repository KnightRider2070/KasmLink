package dockercli_test

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/dockercli"
)

func TestCreateTarWithContext(t *testing.T) {
	t.Run("Valid Directory", func(t *testing.T) {
		// Setup: Create a temporary directory with sample files
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
		assert.NoError(t, err)
		err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "subdir", "file2.txt"), []byte("content2"), 0644)
		assert.NoError(t, err)

		// Create a mock DockerClient with a LocalFileSystem
		client := dockercli.NewDockerClient(nil, dockercli.NewLocalFileSystem())

		// Execute: Create tar from the directory
		tarReader, err := client.CreateTarWithContext(tempDir)
		assert.NoError(t, err)

		// Verify: Check contents of the tar archive
		tarContent := map[string]string{
			".":                "", // Root directory
			"file1.txt":        "content1",
			"subdir":           "", // Subdirectory
			"subdir/file2.txt": "content2",
		}

		// Extract tar and verify contents
		verifyTarContents(t, tarReader, tarContent)
	})

	t.Run("Empty Directory Path", func(t *testing.T) {
		// Create a mock DockerClient with a LocalFileSystem
		client := dockercli.NewDockerClient(nil, dockercli.NewLocalFileSystem())

		// Execute: Call CreateTarWithContext with an empty path
		tarReader, err := client.CreateTarWithContext("")
		assert.Error(t, err)
		assert.Nil(t, tarReader)
		assert.Contains(t, err.Error(), "build context directory cannot be empty")
	})

	t.Run("Non-Existent Directory", func(t *testing.T) {
		// Create a mock DockerClient with a LocalFileSystem
		client := dockercli.NewDockerClient(nil, dockercli.NewLocalFileSystem())

		// Execute: Call CreateTarWithContext with a non-existent path
		tarReader, err := client.CreateTarWithContext("/non/existent/path")
		assert.Error(t, err)
		assert.Nil(t, tarReader)
		assert.Contains(t, err.Error(), "failed to create tar archive")
	})
}

// Helper function to verify tar contents
func verifyTarContents(t *testing.T, tarReader io.Reader, expectedContent map[string]string) {
	tr := tar.NewReader(tarReader)
	foundFiles := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)

		// Normalize the header name to use forward slashes
		normalizedHeaderName := normalizePath(header.Name)

		// Check if the file/directory is expected
		content, exists := expectedContent[normalizedHeaderName]
		assert.True(t, exists, "Unexpected file or directory: "+normalizedHeaderName)

		if header.Typeflag == tar.TypeReg {
			// For regular files, verify content
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, tr)
			assert.NoError(t, err)
			assert.Equal(t, content, buf.String(), "Content mismatch for "+normalizedHeaderName)
		}

		// Mark the file/directory as found
		foundFiles[normalizedHeaderName] = content
	}

	// Ensure all expected files/directories were found
	assert.Equal(t, expectedContent, foundFiles)
}

// normalizePath ensures consistent use of forward slashes in paths
func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
