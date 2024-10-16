package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// ListImages fetches the available images from KASM API.
func (api *KasmAPI) ListImages() ([]Image, error) {
	url := fmt.Sprintf("%s/api/public/images", api.BaseURL)
	log.Printf("Fetching list of images from URL: %s", url)
	response, err := api.MakeGetRequest(url)
	if err != nil {
		log.Printf("Error fetching images: %v", err)
		return nil, err
	}

	var images []Image
	if err := json.Unmarshal(response, &images); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully fetched %d images", len(images))
	return images, nil
}

// DeployImage deploys a new image to a specific host.
func (api *KasmAPI) DeployImage(imageID, host string) error {
	url := fmt.Sprintf("%s/api/public/deploy/%s", api.BaseURL, imageID)
	log.Printf("Deploying image %s to host %s", imageID, host)
	payload := map[string]string{
		"host": host,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error deploying image: %v", err)
		return err
	}

	if len(response) > 0 {
		log.Printf("Image %s successfully deployed on host %s", imageID, host)
	}
	return nil
}
