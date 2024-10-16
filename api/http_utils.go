package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// MakeGetRequest handles making GET requests to the KASM API.
func (api *KasmAPI) MakeGetRequest(url string) ([]byte, error) {
	log.Printf("Making GET request to URL: %s", url)
	client := &http.Client{}
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

// MakePostRequest handles making POST requests to the KASM API.
func (api *KasmAPI) MakePostRequest(url string, payload interface{}) ([]byte, error) {
	log.Printf("Making POST request to URL: %s with payload: %v", url, payload)
	client := &http.Client{}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to create POST request: %v", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("POST request failed: %v", err)
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Unexpected response status: %s", resp.Status)
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("POST request to URL %s succeeded", url)
	return body, nil
}
