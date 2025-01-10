package talos

import (
	"fmt"
	"github.com/rs/zerolog/log"
	programmConfig "kasmlink/internal/config"
)

// InitializeClusterDirectory initializes the cluster directory by copying files from the embedded "cluster" directory.
func InitializeClusterDirectory() error {
	// Ensure the target directory exists
	if err := createDirIfNotExist(programmConfig.ClusterDirectory); err != nil {
		return fmt.Errorf("failed to create or validate cluster directory %s: %w", "./", err)
	}

	// Copy files from the embedded "cluster" directory
	if err := copyFilesFromClusterDirectory("cluster", programmConfig.ClusterDirectory); err != nil {
		return fmt.Errorf("failed to copy files from embedded cluster: %w", err)
	}

	// Generate Age encryption keys
	if err := AgeKeyGen(); err != nil {
		log.Error().Msgf("Failed to generate Age keys: %v", err)
		return err
	}

	// Get the public key from the Age key file and replace it in the .sops.yaml
	publicKey, err := GetPubKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	if err := ReplaceInFile(programmConfig.ConfigPaths.SopsFile, "AGE_PUBLIC_KEY", publicKey); err != nil {
		return fmt.Errorf("failed to replace placeholder with Age public key: %w", err)
	}

	log.Info().Msgf("Cluster directory initialized and files copied to %s", "./")
	return nil
}
