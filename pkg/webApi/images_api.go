package webApi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// ListImages fetches the available images from the KASM API.
// Note: requires api key with "Images View" permission
func (api *KasmAPI) ListImages(ctx context.Context) ([]Image, error) {
	endpoint := "/api/public/get_images"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Msg("Initiating request to fetch list of images")

	// Construct request payload
	requestPayload := GetImagesRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Msg("Failed to fetch images from KASM API")
		return nil, fmt.Errorf("failed to fetch images: %w", err)
	}

	// Parse the response into GetImagesResponse struct
	var imagesResponse GetImagesResponse
	if err := json.Unmarshal(responseBytes, &imagesResponse); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Msg("Failed to decode images response")
		return nil, fmt.Errorf("failed to decode images response: %v", err)
	}

	log.Info().
		Int("image_count", len(imagesResponse.Images)).
		Str("method", "POST").
		Str("endpoint", endpoint).
		Msg("Successfully fetched images from KASM API")

	return imagesResponse.Images, nil
}
