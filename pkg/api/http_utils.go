package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Configure zerolog to log to console
func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
}

// AskUserToSkipTLS asks the user whether to skip TLS certificate verification
func AskUserToSkipTLS() bool {
	var input string
	fmt.Print("Do you want to skip TLS certificate verification? (y/N): ")
	fmt.Scanln(&input)
	skipTLS := input == "y" || input == "Y"
	if skipTLS {
		log.Warn().Msg("User chose to skip TLS certificate verification.")
	} else {
		log.Info().Msg("User chose to enable TLS certificate verification.")
	}
	return skipTLS
}

// CreateHTTPClient creates an HTTP client with optional TLS verification
func CreateHTTPClient(skipTLSVerification bool) *http.Client {
	if skipTLSVerification {
		log.Warn().Msg("TLS certificate verification is disabled.")
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // This skips TLS verification
				},
			},
		}
	}

	log.Info().Msg("TLS certificate verification is enabled.")
	return &http.Client{Timeout: 30 * time.Second}
}

// MakeGetRequest handles making GET requests to the KASM API.
func (api *KasmAPI) MakeGetRequest(url string) ([]byte, error) {
	log.Info().Str("url", url).Msg("Initiating GET request")
	client := CreateHTTPClient(api.SkipTLSVerification)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to create GET request")
		return nil, fmt.Errorf("failed to create GET request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("GET request failed")
		return nil, fmt.Errorf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Str("url", url).Int("status", resp.StatusCode).Msg("Unexpected response status")
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Info().Str("url", url).Msg("GET request succeeded")
	return body, nil
}

// MakePostRequest handles making POST requests to the KASM API
func (api *KasmAPI) MakePostRequest(url string, payload interface{}) ([]byte, error) {
	log.Debug().Str("url", url).Msg("Preparing POST request payload")

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload for POST request")
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create an HTTP client that can skip TLS verification if needed
	client := CreateHTTPClient(api.SkipTLSVerification)

	// Create and send POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to create POST request")
		return nil, fmt.Errorf("failed to create POST request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	log.Info().Str("url", url).Msg("Sending POST request")
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("POST request failed")
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Warn().Str("url", url).Int("status", resp.StatusCode).Msg("Received non-OK response")
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to read POST response body")
		return nil, err
	}

	log.Info().Str("url", url).Msg("POST request succeeded")
	return responseData, nil
}
