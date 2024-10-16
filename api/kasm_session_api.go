package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// RequestKasmSession requests a new Kasm session.
func (api *KasmAPI) RequestKasmSession(userID, imageID string) (*KasmSessionResponse, error) {
	url := fmt.Sprintf("%s/api/public/request_kasm", api.BaseURL)
	log.Printf("Requesting new Kasm session for user ID: %s with image ID: %s", userID, imageID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"user_id":        userID,
		"image_id":       imageID,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error requesting Kasm session: %v", err)
		return nil, err
	}

	var kasmSession KasmSessionResponse
	if err := json.Unmarshal(response, &kasmSession); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully created Kasm session with ID: %s", kasmSession.KasmID)
	return &kasmSession, nil
}

// GetKasmStatus retrieves the status of a Kasm session.
func (api *KasmAPI) GetKasmStatus(userID, kasmID string) (*KasmStatusResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_kasm_status", api.BaseURL)
	log.Printf("Retrieving status for Kasm session with ID: %s for user ID: %s", kasmID, userID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"user_id":        userID,
		"kasm_id":        kasmID,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error retrieving Kasm status: %v", err)
		return nil, err
	}

	var kasmStatus KasmStatusResponse
	if err := json.Unmarshal(response, &kasmStatus); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully retrieved status for Kasm session with ID: %s", kasmID)
	return &kasmStatus, nil
}

// DestroyKasmSession destroys an existing Kasm session.
func (api *KasmAPI) DestroyKasmSession(userID, kasmID string) error {
	url := fmt.Sprintf("%s/api/public/destroy_kasm", api.BaseURL)
	log.Printf("Destroying Kasm session with ID: %s for user ID: %s", kasmID, userID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"user_id":        userID,
		"kasm_id":        kasmID,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error destroying Kasm session: %v", err)
		return err
	}

	if len(response) > 0 {
		log.Printf("Kasm session %s successfully destroyed", kasmID)
	}
	return nil
}
