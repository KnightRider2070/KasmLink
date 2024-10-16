package api

import "log"

// KasmAPI holds the base URL and credentials for making requests to the KASM API.
type KasmAPI struct {
	BaseURL      string
	APIKey       string
	APIKeySecret string
}

// NewKasmAPI creates a new instance of KasmAPI with provided credentials.
func NewKasmAPI(baseURL, apiKey, apiKeySecret string) *KasmAPI {
	log.Printf("Creating new KasmAPI instance with BaseURL: %s", baseURL)
	return &KasmAPI{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		APIKeySecret: apiKeySecret,
	}
}
