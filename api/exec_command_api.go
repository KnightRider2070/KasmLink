package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// ExecCommand executes an arbitrary command inside a user's Kasm session.
func (api *KasmAPI) ExecCommand(userID, kasmID string, execConfig ExecConfig) (*ExecCommandResponse, error) {
	url := fmt.Sprintf("%s/api/public/exec_command_kasm", api.BaseURL)
	log.Printf("Executing command in Kasm session with ID: %s for user ID: %s", kasmID, userID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"user_id":        userID,
		"kasm_id":        kasmID,
		"exec_config":    execConfig,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error executing command in Kasm session: %v", err)
		return nil, err
	}

	var execResponse ExecCommandResponse
	if err := json.Unmarshal(response, &execResponse); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully executed command in Kasm session with ID: %s", kasmID)
	return &execResponse, nil
}
