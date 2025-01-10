package talos

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"filippo.io/age"
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	programmConfig "kasmlink/internal/config"
	"os"
	"strings"
	"time"
)

// AgeKeyGen creates a new Age key pair and saves it to a file with secure permissions.
// It logs different stages of the process with configurable verbosity levels.
func AgeKeyGen() error {
	ageKeyFile := programmConfig.ConfigPaths.AgeKeyFile

	// Check if the Age key file already exists
	if _, err := os.Stat(ageKeyFile); err == nil {
		log.Info().Msgf("Age key file %s already exists. Skipping key generation.", ageKeyFile)
		return nil
	} else if os.IsNotExist(err) {
		// Proceed with key generation since the file doesn't exist

		file, err := os.Create(ageKeyFile)
		if err != nil {
			return fmt.Errorf("failed to create output file %q: %w", ageKeyFile, err)
		}
		defer file.Close()

		// Set file permissions to be secure (0600)
		if err := os.Chmod(ageKeyFile, 0600); err != nil {
			return fmt.Errorf("failed to set secure permissions for %q: %w", ageKeyFile, err)
		}

		log.Debug().Msg("Generating a new Age key pair.")
		keyPair, err := age.GenerateX25519Identity()
		if err != nil {
			return fmt.Errorf("failed to generate Age key pair: %w", err)
		}

		log.Trace().Msg("Writing key pair details to the file.")
		fmt.Fprintf(file, "# created: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(file, "# public key: %s\n", keyPair.Recipient())
		fmt.Fprintf(file, "%s\n", keyPair)

		log.Info().Msgf("Age key pair generated and saved to %s", ageKeyFile)
	} else {
		return fmt.Errorf("failed to check existence of file %q: %w", ageKeyFile, err)
	}

	return nil
}

// GetPubKey retrieves the public key from the specified Age key file.
func GetPubKey() (string, error) {
	filePath := programmConfig.ConfigPaths.AgeKeyFile

	// Ensure the file exists before proceeding
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Error().Err(err).Msgf("The Age key file does not exist: %s", filePath)
		return "", fmt.Errorf("key file %s does not exist", filePath)
	}

	// Read the file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to read the file: %s", filePath)
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Search for the public key in the file content
	lines := strings.Split(string(fileContent), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# public key:") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				publicKey := strings.TrimSpace(parts[1])
				if publicKey == "" {
					return "", fmt.Errorf("empty public key found in file %s", filePath)
				}
				log.Info().Msgf("Public key found: %s", publicKey)
				return publicKey, nil
			}
		}
	}

	log.Error().Msg("Public key not found in the file")
	return "", fmt.Errorf("public key not found in %s", filePath)
}

// GetAgeKey reads the specified file and returns the secret key found within it.
func GetAgeKey() (string, error) {
	// Specify the path to the secret key file
	filename := programmConfig.ConfigPaths.AgeKeyFile
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var secretKey string

	// Use a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Look for the secret key line (the line that starts with "AGE-SECRET-KEY-")
		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			secretKey = line
			break
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// If no secret key is found in the file, return an error
	if secretKey == "" {
		return "", fmt.Errorf("secret key not found in file %s", filename)
	}

	return secretKey, nil
}

// ensurePublicKeyExists ensures the public key file exists, generating it from the existing secret YAML if necessary.
func ensurePublicKeyExists(secretPath, publicKeyPath string) error {
	// Check if public key already exists
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		// Public key file doesn't exist, generate it from the secret YAML
		secretYAML, err := os.ReadFile(secretPath)
		if err != nil {
			return fmt.Errorf("failed to read existing secret YAML: %w", err)
		}

		// Unmarshal the secret YAML into a Secret object
		var secret corev1.Secret
		if err := yaml.Unmarshal(secretYAML, &secret); err != nil {
			return fmt.Errorf("failed to unmarshal secret YAML: %w", err)
		}

		// Extract the public key from the secret
		if ppk, ok := secret.StringData["identity.pub"]; ok {
			// Write the public key to the file
			if err := os.WriteFile(publicKeyPath, []byte(ppk), 0644); err != nil {
				return fmt.Errorf("failed to write public key to file: %w", err)
			}
			log.Info().Msgf("Public key saved to: %s\n", publicKeyPath)
		} else {
			return fmt.Errorf("identity.pub not found in existing secret YAML")
		}
	} else {
		log.Info().Msgf("Public key file already exists: %s\n", publicKeyPath)
	}

	return nil
}

// pemBlockForKey creates a PEM block for the private key.
func pemBlockForKey(key *ecdsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}), nil
}

// publicKeyToOpenSSH converts an ECDSA public key to OpenSSH format.
func publicKeyToOpenSSH(pubKey *ecdsa.PublicKey) (string, error) {
	// Convert the public key to PEM format
	pubASN1, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Convert the public key to OpenSSH format
	return fmt.Sprintf("ecdsa-sha2-nistp384 %s", base64.StdEncoding.EncodeToString(pubASN1)), nil
}
