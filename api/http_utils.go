package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AskUserToSkipTLS asks user whether to skip TLS certificate verification
func AskUserToSkipTLS() bool {
	var input string
	fmt.Println("Do you want to skip TLS certificate verification? (y/N): ")
	fmt.Scanln(&input)
	return input == "y" || input == "Y"
}

// CreateHTTPClient creates an HTTP client with optional TLS verification
func CreateHTTPClient(skipTLSVerification bool) *http.Client {
	if skipTLSVerification {
		// Warning: Skipping TLS certificate verification
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // This skips TLS verification
				},
			},
		}
	}

	// Default client with proper TLS verification
	return &http.Client{Timeout: 30 * time.Second}
}

// MakeGetRequest handles making GET requests to the KASM API.
func (api *KasmAPI) MakeGetRequest(url string) ([]byte, error) {
	log.Printf("Making GET request to URL: %s", url)
	client := CreateHTTPClient(api.SkipTLSVerification)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Failed to create GET request: %v", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("GET request failed: %v", err)
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response status: %s", resp.Status)
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("GET request to URL %s succeeded", url)
	return body, nil
}

// MakePostRequest handles making POST requests to the KASM API
func (api *KasmAPI) MakePostRequest(url string, payload interface{}) ([]byte, error) {
	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create an HTTP client that can skip TLS verification if needed
	client := CreateHTTPClient(api.SkipTLSVerification)

	// Create and send POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}
