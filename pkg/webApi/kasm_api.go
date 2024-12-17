package webApi

import (
	"crypto/tls"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

// KasmAPI holds the base URL, credentials, and HTTP client for making requests to the KASM API.
type KasmAPI struct {
	BaseURL             string
	APIKey              string
	APIKeySecret        string
	SkipTLSVerification bool
	RequestTimeout      time.Duration
	Client              *http.Client
}

// NewKasmAPI creates a new instance of KasmAPI with provided credentials.
// It initializes the HTTP client with appropriate configurations.
func NewKasmAPI(baseURL, apiKey, apiKeySecret string, skipTLSVerification bool, requestTimeout time.Duration) *KasmAPI {
	if requestTimeout == 0 {
		requestTimeout = 240 * time.Second
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLSVerification, // Configures TLS verification
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		IdleConnTimeout:     240 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		DisableKeepAlives:   false,
		MaxConnsPerHost:     100,
	}

	client := &http.Client{
		Timeout:   requestTimeout,
		Transport: transport,
	}

	log.Info().
		Str("base_url", baseURL).
		Bool("skip_tls_verification", skipTLSVerification).
		Dur("request_timeout", requestTimeout).
		Interface("tls_config", tlsConfig).
		Msg("Creating new KasmAPI instance with configured HTTP client")

	return &KasmAPI{
		BaseURL:             baseURL,
		APIKey:              apiKey,
		APIKeySecret:        apiKeySecret,
		SkipTLSVerification: skipTLSVerification,
		RequestTimeout:      requestTimeout,
		Client:              client,
	}
}
