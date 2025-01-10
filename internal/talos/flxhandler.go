package talos

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// fileExists checks if a file exists at the given path.
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// createKustomizationFile generates a ks.yaml file for Flux Kustomization.
func createKustomizationFile(directory, parentFolder string) error {
	linuxPath := filepath.ToSlash(directory) // Ensure the path uses forward slashes

	// Prepare ks.yaml content for Flux Kustomization
	content := fmt.Sprintf(`
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: %s
  namespace: flux-system
spec:
  interval: 10m
  path: %s/app
  prune: true
  sourceRef:
    kind: GitRepository
    name: cluster
`, parentFolder, linuxPath)

	// Write the ks.yaml file
	ksYamlPath := filepath.Join(directory, "ks.yaml")
	return ioutil.WriteFile(ksYamlPath, []byte(content), 0644)
}

// createOrUpdateKustomizationFile creates or updates a kustomization.yaml file in the specified directory.
func createOrUpdateKustomizationFile(directory string) error {
	kustomizationFile := filepath.Join(directory, "kustomization.yaml")
	var content string

	// If the kustomization.yaml exists, read its current content
	if fileExists(kustomizationFile) {
		data, err := ioutil.ReadFile(kustomizationFile)
		if err != nil {
			return err
		}
		content = string(data)
	} else {
		// Initialize the content if it doesn't exist
		content = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
`
	}

	// List files and directories in the current directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	// Collect resources to add to the kustomization.yaml
	var resources []string
	for _, file := range files {
		name := file.Name()

		// Skip existing kustomization.yaml and ks.yaml
		if name == "kustomization.yaml" || name == "ks.yaml" {
			continue
		}

		// Add YAML files and directories
		if strings.HasSuffix(name, ".yaml") || file.IsDir() {
			if file.IsDir() && fileExists(filepath.Join(directory, name, "ks.yaml")) {
				// If it's a directory with ks.yaml inside, include the ks.yaml file
				name = fmt.Sprintf("%s/ks.yaml", name)
			}

			// Avoid duplicates in the kustomization.yaml file
			if !strings.Contains(content, name) {
				if name == "namespace.yaml" {
					// Namespace entries should be added first
					resources = append([]string{name}, resources...)
				} else {
					resources = append(resources, name)
				}
			}
		}
	}

	// Append new resources to kustomization.yaml content
	for _, resource := range resources {
		content += fmt.Sprintf("  - %s\n", resource)
	}

	// Clean up duplicate resource entries
	contentLines := strings.Split(content, "\n")
	for _, line := range contentLines {
		if strings.HasSuffix(line, "/ks.yaml") {
			// Remove redundant directory entries for ks.yaml
			dirPrefix := strings.TrimSuffix(line, "/ks.yaml")
			var updatedContent []string
			for _, item := range contentLines {
				if item != dirPrefix {
					updatedContent = append(updatedContent, item)
				}
			}
			contentLines = updatedContent
		}
	}
	content = strings.Join(contentLines, "\n")

	// Write back the updated kustomization.yaml
	return ioutil.WriteFile(kustomizationFile, []byte(content), 0644)
}

// ProcessDirectory processes directories and creates or updates necessary Flux files recursively.
func ProcessDirectory(directory string) error {
	var hasAppFolder, hasKsYaml bool

	// Read directory contents
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	// Check for "app" folder and "ks.yaml" file
	for _, file := range files {
		if file.IsDir() && file.Name() == "app" {
			hasAppFolder = true
		}
		if file.Name() == "ks.yaml" {
			hasKsYaml = true
		}
	}

	// Create ks.yaml file if "app" folder exists and ks.yaml does not
	if hasAppFolder && !hasKsYaml {
		parentFolder := filepath.Base(directory)
		if err := createKustomizationFile(directory, parentFolder); err != nil {
			return err
		}
	} else if !hasKsYaml {
		// Update kustomization.yaml if ks.yaml does not exist
		if err := createOrUpdateKustomizationFile(directory); err != nil {
			return err
		}
	}

	// Recursively process subdirectories
	for _, file := range files {
		if file.IsDir() {
			subdirPath := filepath.Join(directory, file.Name())
			if err := ProcessDirectory(subdirPath); err != nil {
				return err
			}
		}
	}

	return nil
}
