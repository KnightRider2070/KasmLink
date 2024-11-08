package api

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

// ListImages fetches the available images from the KASM API.
func (api *KasmAPI) ListImages() ([]Image, error) {
	url := fmt.Sprintf("%s/api/public/get_images", api.BaseURL)
	log.Info().Str("url", url).Msg("Fetching list of images")

	// Construct request payload
	request := GetImagesRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}

	// Make POST request with retries
	retries := 3
	var response []byte
	var err error
	for retries > 0 {
		response, err = api.MakePostRequest(url, request)
		if err != nil {
			log.Error().Err(err).Str("url", url).Int("retries_left", retries-1).Msg("Failed to fetch images, retrying")
			retries--
			if retries == 0 {
				return nil, fmt.Errorf("failed to fetch images after retries: %v", err)
			}
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	// Parse response
	var imagesResponse GetImagesResponse
	if err := json.Unmarshal(response, &imagesResponse); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to decode image list response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().
		Int("image_count", len(imagesResponse.Images)).
		Str("url", url).
		Msg("Successfully fetched images")

	return imagesResponse.Images, nil
}
