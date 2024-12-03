package webApi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// RequestKasmSession requests a new Kasm session.
func (api *KasmAPI) RequestKasmSession(ctx context.Context, userID string, imageID string, envArgs map[string]string) (*RequestKasmResponse, error) {
	endpoint := "/api/public/request_kasm"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("image_id", imageID).
		Msg("Requesting Kasm session")

	// Create a new RequestKasmRequest struct
	req := RequestKasmRequest{
		APIKey:        api.APIKey,
		APIKeySecret:  api.APIKeySecret,
		UserID:        userID,
		ImageID:       imageID,
		EnableSharing: false, //TODO: Think about if this should be configurable, securtiy wise not a good idea
		Environment:   envArgs,
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", req.UserID).
			Str("image_id", req.ImageID).
			Msg("Error requesting Kasm session")
		return nil, fmt.Errorf("error requesting Kasm session: %w", err)
	}

	// Parse the response into RequestKasmResponse struct
	var kasmResponse RequestKasmResponse
	if err := json.Unmarshal(responseBytes, &kasmResponse); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", req.UserID).
			Msg("Failed to decode Kasm session response")
		return nil, fmt.Errorf("failed to decode Kasm session response: %v", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", req.UserID).
		Str("session_id", kasmResponse.KasmID).
		Str("status", kasmResponse.Status).
		Msg("Successfully created Kasm session")

	return &kasmResponse, nil
}

// GetKasmStatus retrieves the status of an existing Kasm session.
func (api *KasmAPI) GetKasmStatus(ctx context.Context, req GetKasmStatusRequest) (*GetKasmStatusResponse, error) {
	endpoint := "/api/public/get_kasm_status"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Msg("Getting status for Kasm session")

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("kasm_id", req.KasmID).
			Msg("Error getting Kasm session status")
		return nil, fmt.Errorf("error getting Kasm session status: %w", err)
	}

	// Parse the response into GetKasmStatusResponse struct
	var statusResponse GetKasmStatusResponse
	if err := json.Unmarshal(responseBytes, &statusResponse); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("kasm_id", req.KasmID).
			Msg("Failed to decode Kasm status response")
		return nil, fmt.Errorf("failed to decode Kasm status response: %v", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Str("operational_status", statusResponse.OperationalStatus).
		Msg("Successfully retrieved Kasm session status")

	return &statusResponse, nil
}

// DestroyKasmSession destroys an existing Kasm session.
func (api *KasmAPI) DestroyKasmSession(ctx context.Context, req DestroyKasmRequest) error {
	endpoint := "/api/public/destroy_kasm"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Str("user_id", req.UserID).
		Msg("Destroying Kasm session")

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("kasm_id", req.KasmID).
			Str("user_id", req.UserID).
			Msg("Error destroying Kasm session")
		return fmt.Errorf("error destroying Kasm session: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Str("user_id", req.UserID).
		Msg("Successfully destroyed Kasm session")
	return nil
}

// ExecCommand executes a command in an existing Kasm session.
func (api *KasmAPI) ExecCommand(ctx context.Context, req ExecCommandRequest) error {
	endpoint := "/api/public/exec_command_kasm"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Str("command", req.ExecConfig.Cmd).
		Msg("Executing command in Kasm session")

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("kasm_id", req.KasmID).
			Str("command", req.ExecConfig.Cmd).
			Msg("Error executing command in Kasm session")
		return fmt.Errorf("error executing command in Kasm session: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("kasm_id", req.KasmID).
		Str("command", req.ExecConfig.Cmd).
		Msg("Successfully executed command in Kasm session")
	return nil
}
