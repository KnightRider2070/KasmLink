package api

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

func (api *KasmAPI) RequestKasmSession(req RequestKasmRequest) (*RequestKasmResponse, error) {
	url := fmt.Sprintf("%s/api/public/request_kasm", api.BaseURL)
	log.Printf("Requesting Kasm session for user: %s with image: %s", req.UserID, req.ImageID)
	response, err := api.MakePostRequest(url, req)
	if err != nil {
		log.Printf("Error requesting Kasm session: %v", err)
		return nil, err
	}

	var kasmResponse RequestKasmResponse
	if err := json.Unmarshal(response, &kasmResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &kasmResponse, nil
}

func (api *KasmAPI) GetKasmStatus(req GetKasmStatusRequest) (*GetKasmStatusResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_kasm_status", api.BaseURL)
	log.Printf("Getting status for Kasm session: %s", req.KasmID)
	response, err := api.MakePostRequest(url, req)
	if err != nil {
		return nil, err
	}

	var statusResponse GetKasmStatusResponse
	if err := json.Unmarshal(response, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &statusResponse, nil
}

func (api *KasmAPI) DestroyKasmSession(req DestroyKasmRequest) error {
	url := fmt.Sprintf("%s/api/public/destroy_kasm", api.BaseURL)
	log.Printf("Destroying Kasm session: %s for user: %s", req.KasmID, req.UserID)
	_, err := api.MakePostRequest(url, req)
	return err
}

func (api *KasmAPI) ExecCommand(req ExecCommandRequest) error {
	url := fmt.Sprintf("%s/api/public/exec_command_kasm", api.BaseURL)
	log.Printf("Executing command in Kasm session: %s", req.KasmID)
	_, err := api.MakePostRequest(url, req)
	return err
}
