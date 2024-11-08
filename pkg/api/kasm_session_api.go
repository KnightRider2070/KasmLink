package api

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// RequestKasmSession requests a new Kasm session.
func (api *KasmAPI) RequestKasmSession(req RequestKasmRequest) (*RequestKasmResponse, error) {
	url := fmt.Sprintf("%s/api/public/request_kasm", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", req.UserID).Str("image_id", req.ImageID).Msg("Requesting Kasm session")

	response, err := api.MakePostRequest(url, req)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Str("image_id", req.ImageID).Msg("Error requesting Kasm session")
		return nil, err
	}

	var kasmResponse RequestKasmResponse
	if err := json.Unmarshal(response, &kasmResponse); err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Failed to decode Kasm session response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("user_id", req.UserID).Str("session_id", kasmResponse.KasmID).Str("status", kasmResponse.Status).Msg("Successfully created Kasm session")
	return &kasmResponse, nil
}

// GetKasmStatus retrieves the status of an existing Kasm session.
func (api *KasmAPI) GetKasmStatus(req GetKasmStatusRequest) (*GetKasmStatusResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_kasm_status", api.BaseURL)
	log.Info().Str("url", url).Str("kasm_id", req.KasmID).Msg("Getting status for Kasm session")

	response, err := api.MakePostRequest(url, req)
	if err != nil {
		log.Error().Err(err).Str("kasm_id", req.KasmID).Msg("Error getting Kasm session status")
		return nil, err
	}

	var statusResponse GetKasmStatusResponse
	if err := json.Unmarshal(response, &statusResponse); err != nil {
		log.Error().Err(err).Str("kasm_id", req.KasmID).Msg("Failed to decode Kasm status response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("kasm_id", req.KasmID).Str("operational_status", statusResponse.OperationalStatus).Msg("Successfully retrieved Kasm session status")
	return &statusResponse, nil
}

// DestroyKasmSession destroys an existing Kasm session.
func (api *KasmAPI) DestroyKasmSession(req DestroyKasmRequest) error {
	url := fmt.Sprintf("%s/api/public/destroy_kasm", api.BaseURL)
	log.Info().Str("url", url).Str("kasm_id", req.KasmID).Str("user_id", req.UserID).Msg("Destroying Kasm session")

	_, err := api.MakePostRequest(url, req)
	if err != nil {
		log.Error().Err(err).Str("kasm_id", req.KasmID).Str("user_id", req.UserID).Msg("Error destroying Kasm session")
		return err
	}

	log.Info().Str("kasm_id", req.KasmID).Str("user_id", req.UserID).Msg("Successfully destroyed Kasm session")
	return nil
}

// ExecCommand executes a command in an existing Kasm session.
func (api *KasmAPI) ExecCommand(req ExecCommandRequest) error {
	url := fmt.Sprintf("%s/api/public/exec_command_kasm", api.BaseURL)
	log.Info().Str("url", url).Str("kasm_id", req.KasmID).Str("command", req.ExecConfig.Cmd).Msg("Executing command in Kasm session")

	_, err := api.MakePostRequest(url, req)
	if err != nil {
		log.Error().Err(err).Str("kasm_id", req.KasmID).Str("command", req.ExecConfig.Cmd).Msg("Error executing command in Kasm session")
		return err
	}

	log.Info().Str("kasm_id", req.KasmID).Str("command", req.ExecConfig.Cmd).Msg("Successfully executed command in Kasm session")
	return nil
}
