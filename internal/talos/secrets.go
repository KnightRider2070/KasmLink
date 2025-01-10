package talos

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	programmConfig "kasmlink/internal/config"
	"os"
	"path/filepath"
)

// GenSopsSecret generates a Kubernetes secret YAML for the SOPS key and saves it to the appropriate file path.
func GenSopsSecret() error {
	// Define the secret file path
	secretPath := filepath.Join(programmConfig.ConfigPaths.KubernetesDir, "flux-system", "flux", "sopssecret.secret.yaml")

	// Read age key file

	// Retrieve the SOPS secret key
	ageSecKey, err := GetAgeKey()
	if err != nil {
		return fmt.Errorf("failed to get SOPS secret key: %w", err)
	}

	// Create the secret content in a structured map
	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      "sops-age",
			"namespace": "flux-system",
		},
		"stringData": map[string]interface{}{
			"age.agekey": ageSecKey,
		},
		"type": string(corev1.SecretTypeOpaque),
	}

	// Marshal the secret data into YAML format
	secretYAML, err := yaml.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret to YAML: %w", err)
	}

	// Ensure the target directory exists before writing the file
	if err := ensureDirExists(filepath.Dir(secretPath)); err != nil {
		return fmt.Errorf("failed to create directories for secret path: %w", err)
	}

	// Write the generated YAML to the secret file
	if err := os.WriteFile(secretPath, secretYAML, 0644); err != nil {
		return fmt.Errorf("failed to write secret YAML to file: %w", err)
	}

	// Log the successful creation of the secret file
	log.Info().Msgf("SOPS secret YAML saved to: %s", secretPath)

	return nil
}

// generateAndSaveKeys generates the SSH key pair and Kubernetes secret YAML file.
func generateAndSaveKeys(gitURL, secretPath, publicKeyPath string) error {
	// Generate ECDSA private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ECDSA private key: %w", err)
	}

	// Encode private key to PEM format
	privateKeyPEMBlock, err := pemBlockForKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create PEM block for private key: %w", err)
	}

	// Generate OpenSSH formatted public key
	publicKey, err := publicKeyToOpenSSH(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate OpenSSH public key: %w", err)
	}

	// Write public key to file
	if err := os.WriteFile(publicKeyPath, []byte(publicKey), 0644); err != nil {
		return fmt.Errorf("failed to write public key to file: %w", err)
	}
	log.Info().Msgf("Public key saved to: %s\n", publicKeyPath)

	// Generate known_hosts entry for the Git URL
	knownHosts := getKnownHostsEntry(gitURL)

	// Generate Kubernetes secret YAML content
	secret := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name":      "deploy-key",
			"namespace": "flux-system",
		},
		"stringData": map[string]interface{}{
			"identity":     string(privateKeyPEMBlock),
			"identity.pub": publicKey,
			"known_hosts":  knownHosts,
		},
		"type": string(corev1.SecretTypeOpaque),
	}

	// Marshal secret into YAML format
	secretYAML, err := yaml.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret to YAML: %w", err)
	}

	// Create necessary directories and write the secret YAML to the file
	if err := os.MkdirAll(filepath.Dir(secretPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	if err := os.WriteFile(secretPath, secretYAML, 0644); err != nil {
		return fmt.Errorf("failed to write secret YAML to file: %w", err)
	}

	log.Info().Msgf("Kubernetes secret YAML saved to: %s\n", secretPath)
	return nil
}
