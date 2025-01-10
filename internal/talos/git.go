package talos

import (
	"fmt"
	"github.com/rs/zerolog/log"
	programmConfig "kasmlink/internal/config"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UpdateGitRepo checks if a GitHub repository URL is present in the environment file and updates the reference in the "this-repo.yaml" file.
func UpdateGitRepo() error {
	// Retrieve the GITHUB_REPOSITORY value from the environment file
	repoURL, err := GetVarFromEnvFile("GITHUB_REPOSITORY")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get GitHub repository URL from the environment file.")
		return fmt.Errorf("failed to get GitHub repository URL from the environment file: %w", err)
	}

	// If the GITHUB_REPOSITORY is not empty, update the repository reference in the "this-repo.yaml"
	if repoURL != "" {
		// Define the path to the "this-repo.yaml" file
		repoPath := filepath.Join("repositories", "git", "this-repo.yaml")

		// Format the Git repository URL to match the ssh format
		gitRepoURL := FormatGitURL(repoURL)

		// Replace the placeholder in the "this-repo.yaml" file with the formatted Git URL
		err := ReplaceInFile(repoPath, "ssh://REPOSITORY_URL", gitRepoURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update Git repository reference")
			return fmt.Errorf("failed to update Git repository reference: %w", err)
		}

		log.Info().Msgf("Successfully updated Git repository reference in %s", repoPath)
	}
	return nil
}

// FormatGitURL formats the input Git URL according to specific rules to ensure it uses SSH with git@ syntax.
func FormatGitURL(input string) string {
	// Remove the "https://" prefix if it exists
	input = strings.TrimPrefix(input, "https://")

	// Ensure the URL is prefixed with "ssh://"
	if !strings.HasPrefix(input, "ssh://") {
		input = "ssh://" + input
	}

	// Ensure the URL has "ssh://git@" as the prefix
	if !strings.HasPrefix(input, "ssh://git@") {
		// Prepend "ssh://git@" if the prefix is missing
		input = strings.Replace(input, "ssh://", "ssh://git@", 1)
	}

	// Replace any malformed git@git@ with git@
	if strings.Contains(input, "git@git@") {
		input = strings.Replace(input, "git@git@", "git@", 1)
	}

	// Compile a regular expression to match the Git URL structure
	re := regexp.MustCompile(`^ssh://git@([^:/]+)([:/])([\w-]+)/([\w-]+)\.git$`)
	matches := re.FindStringSubmatch(input)

	// If the URL matches the expected pattern, return the formatted URL
	if len(matches) == 5 {
		return fmt.Sprintf("ssh://git@%s/%s/%s.git", matches[1], matches[3], matches[4])
	}

	// Return the input as-is if it doesn't match the expected format
	return input
}

// CreateGitSecret generates a Kubernetes secret YAML file and a public key text file, if not already existing.
func CreateGitSecret(gitURL string) error {
	if gitURL == "" {
		gitURL = "github.com" // Default GitHub URL if not provided
	}

	secretPath := filepath.Join(programmConfig.ConfigPaths.KubernetesDir, "flux-system", "flux", "deploykey.secret.yaml")
	publicKeyPath := filepath.Join(".", "ssh-public-key.txt")

	// Check if the Kubernetes secret YAML already exists
	if _, err := os.Stat(secretPath); os.IsNotExist(err) {
		// Generate and save the key pair if secret doesn't exist
		if err := generateAndSaveKeys(gitURL, secretPath, publicKeyPath); err != nil {
			return fmt.Errorf("failed to generate and save keys: %w", err)
		}
	} else {
		// Secret YAML exists, check if the public key exists and generate it from the secret
		if err := ensurePublicKeyExists(secretPath, publicKeyPath); err != nil {
			return fmt.Errorf("failed to ensure public key exists: %w", err)
		}
	}

	return nil
}
