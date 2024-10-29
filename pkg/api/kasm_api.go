package api

import (
	"github.com/rs/zerolog/log"
)

// KasmAPI holds the base URL and credentials for making requests to the KASM API.
type KasmAPI struct {
	BaseURL             string
	APIKey              string
	APIKeySecret        string
	SkipTLSVerification bool
}

// NewKasmAPI creates a new instance of KasmAPI with provided credentials.
func NewKasmAPI(baseURL, apiKey, apiKeySecret string) *KasmAPI {
	log.Info().
		Str("base_url", baseURL).
		Msg("Creating new KasmAPI instance")

	return &KasmAPI{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		APIKeySecret: apiKeySecret,
	}
}
