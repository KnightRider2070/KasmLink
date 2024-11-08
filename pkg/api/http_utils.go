package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

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
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerification, // Configures TLS verification
		},
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		DisableKeepAlives:   false,
		MaxConnsPerHost:     20,
	}

	log.Info().Bool("tls_verification_skipped", skipTLSVerification).Msg("Configuring HTTP client")

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

// HandleResponse reads the response body and checks for errors or unexpected status codes
func HandleResponse(resp *http.Response, expectedStatusCode int) ([]byte, error) {
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		log.Warn().Str("url", resp.Request.URL.String()).Int("status", resp.StatusCode).Msg("Unexpected response status")
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Str("url", resp.Request.URL.String()).Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Info().Str("url", resp.Request.URL.String()).Msg("Request succeeded")
	return body, nil
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

	retries := 3
	for retries > 0 {
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).Str("url", url).Int("retries_left", retries-1).Msg("GET request failed, retrying")
			retries--
			if retries == 0 {
				return nil, fmt.Errorf("GET request failed after retries: %v", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}
		return HandleResponse(resp, http.StatusOK)
	}

	return nil, fmt.Errorf("GET request failed unexpectedly")
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	// Execute the request
	log.Info().Str("url", url).Msg("Sending POST request")
	retries := 3
	for retries > 0 {
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).Str("url", url).Int("retries_left", retries-1).Msg("POST request failed, retrying")
			retries--
			if retries == 0 {
				return nil, fmt.Errorf("POST request failed after retries: %v", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}
		return HandleResponse(resp, http.StatusOK)
	}

	return nil, fmt.Errorf("POST request failed unexpectedly")
}
