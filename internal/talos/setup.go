package talos

import (
	"github.com/rs/zerolog/log"
	"path"
)

func InitFiles() error {

	// Initialize cluster directories and configuration
	if err := InitializeClusterDirectory(); err != nil {
		log.Error().Msgf("Failed to initialize cluster directory: %v", err)
		return err
	}

	// Generate JSON schema
	if err := GenSchema(); err != nil {
		log.Error().Msgf("Failed to generate schema: %v", err)
		return err
	}

	// Generate TalEnv ConfigMap
	if err := GenTalEnvConfigMap(); err != nil {
		log.Error().Msgf("Failed to generate TalEnv ConfigMap: %v", err)
		return err
	}

	// Update Git repository configuration
	if err := UpdateGitRepo(); err != nil {
		log.Error().Msgf("Failed to update Git repository: %v", err)
		return err
	}

	// Create Git secret
	repoURL, err := GetVarFromEnvFile("GITHUB_REPOSITORY")
	if err != nil {
		log.Error().Msgf("Failed to retrieve GitHub repository from environment: %v", err)
		return err
	}
	if err := CreateGitSecret(repoURL); err != nil {
		log.Error().Msgf("Failed to create Git secret: %v", err)
		return err
	}

	// Generate SOPS secret
	if err := GenSopsSecret(); err != nil {
		log.Error().Msgf("Failed to generate SOPS secret: %v", err)
		return err
	}

	// Process Kubernetes directory with Flux handler
	kubePath := path.Join("./", "kubernetes")
	if err := ProcessDirectory(kubePath); err != nil {
		log.Error().Msgf("Failed to process Kubernetes directory: %v", err)
		return err
	}
	log.Info().Msg("Kustomizations processed successfully.")

	log.Info().Msg("Init: Completed Successfully!")
	return nil
}
