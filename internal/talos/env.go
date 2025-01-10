package talos

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	embedfiles "kasmlink/embedded"
	programmConfig "kasmlink/internal/config"
	"os"
	"path/filepath"
	"strings"
)

// GenTalEnvConfigMap generates the TalEnv configmap reference by reading the `clusterenv.yaml` file,
// adding indentation, appending the cluster name, and replacing placeholders in the config files.
func GenTalEnvConfigMap() error {
	log.Info().Msg("Creating TalEnv configmap reference 'clustersettings'.")

	// Ensure necessary cluster directories exist
	if err := EnsureClusterDirsExist(); err != nil {
		return logAndReturnError(err, "Failed to ensure cluster directories exist.")
	}

	// Read the content of the TalEnv environment file
	talenvContent, err := os.ReadFile(programmConfig.ConfigPaths.ClusterEnvFilePath)
	if err != nil {
		return logAndReturnError(err, "Failed to read TalEnv YAML file.")
	}

	// Add indentation to the content and append the cluster name
	talenvLines := addIndentationToLines(talenvContent)
	clusterName, err := GetVarFromEnvFile("CLUSTER_NAME")
	if err != nil {
		return logAndReturnError(err, "Failed to get Cluster Name from TalEnv YAML file.")
	}

	// Append the cluster name and join the content back
	talenvLines = append(talenvLines, fmt.Sprintf("  CLUSTERNAME: %s", clusterName))
	indentedTalenvContent := strings.Join(talenvLines, "\n")

	// Set up the cluster settings paths
	clusterSettingsPath := filepath.Join("flux-system", "flux", "clustersettings.secret.yaml")
	clusterSettingsDest := filepath.Join(programmConfig.ClusterDirectory, programmConfig.ConfigPaths.KubernetesDir, clusterSettingsPath)

	//TODO: Find issue regarding cant read kubernetes/flux-system/flux/clustersettings.secret.yaml

	// Retrieve the embedded file content for cluster settings
	fileContent, err := embedfiles.ClusterDirectory.ReadFile("kubernetes/flux-system/flux/clustersettings.secret.yaml")
	if err != nil {
		return logAndReturnError(err, "Failed to read embedded file 'kubernetes/flux-system/flux/clustersettings.secret.yaml'")
	}

	// Ensure necessary directories exist before writing the file
	if err := ensureDirExists(filepath.Join(programmConfig.ConfigPaths.KubernetesDir, "flux-system", "flux")); err != nil {
		return logAndReturnError(err, "Failed to create necessary directories for cluster settings.")
	}

	// Write the embedded content to the destination file
	if err := os.WriteFile(clusterSettingsDest, fileContent, 0644); err != nil {
		return logAndReturnError(err, fmt.Sprintf("Failed to write cluster settings to %s", clusterSettingsDest))
	}

	log.Info().Msgf("Cluster settings copied successfully from embedded file to %s", clusterSettingsDest)

	// Replace placeholders in the cluster settings file with actual values
	if err := ReplaceInFile(clusterSettingsDest, "REPLACEWITHENV", indentedTalenvContent); err != nil {
		return logAndReturnError(err, "Failed to replace content in cluster settings file.")
	}

	log.Info().Msg("Configmap reference 'clustersettings' created successfully.")
	return nil
}

// GetVarFromEnvFile reads the environment YAML file and retrieves the value associated with the specified key.
func GetVarFromEnvFile(key string) (string, error) {
	envFilePath := programmConfig.ConfigPaths.ClusterEnvFilePath
	fileContent, err := os.ReadFile(envFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read the cluster environment file %s: %w", envFilePath, err)
	}

	var envData map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &envData); err != nil {
		return "", fmt.Errorf("failed to unmarshal the cluster environment file %s: %w", envFilePath, err)
	}

	value, ok := envData[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("key %s not found or is empty in the environment file %s", key, envFilePath)
	}

	return value, nil
}
