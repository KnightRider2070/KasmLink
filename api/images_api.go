package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// ListImages fetches the available images from KASM API.
func (api *KasmAPI) ListImages() ([]Image, error) {
	url := fmt.Sprintf("%s/api/public/get_images", api.BaseURL)
	log.Printf("Fetching list of images from URL: %s", url)

	// Construct request payload
	request := GetImagesRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}

	// Make POST request
	response, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Printf("Error fetching images: %v", err)
		return nil, err
	}

	// Parse response
	var imagesResponse GetImagesResponse
	if err := json.Unmarshal(response, &imagesResponse); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully fetched %d images", len(imagesResponse.Images))
	return imagesResponse.Images, nil
}
