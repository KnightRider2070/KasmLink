package api

import (
	"github.com/rs/zerolog/log"
	"time"
)

// KasmAPI holds the base URL and credentials for making requests to the KASM API.
type KasmAPI struct {
	BaseURL             string
	APIKey              string
	APIKeySecret        string
	SkipTLSVerification bool
	RequestTimeout      time.Duration
}

// NewKasmAPI creates a new instance of KasmAPI with provided credentials.
func NewKasmAPI(baseURL, apiKey, apiKeySecret string, skipTLSVerification bool, requestTimeout time.Duration) *KasmAPI {
	log.Info().
		Str("base_url", baseURL).
		Msg("Creating new KasmAPI instance")

	return &KasmAPI{
		BaseURL:             baseURL,
		APIKey:              apiKey,
		APIKeySecret:        apiKeySecret,
		SkipTLSVerification: skipTLSVerification,
		RequestTimeout:      requestTimeout,
	}
}
