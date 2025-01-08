package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/rs/zerolog/log"
)

// CreateHTTPClient creates and returns an HTTP client with optional TLS skipping.
func CreateHTTPClient(skipTLSVerification bool) *http.Client {
	if skipTLSVerification {
		log.Debug().Msg("Creating HTTP client with TLS verification skipped.")
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
	log.Debug().Msg("Creating HTTP client with TLS verification enabled.")
	return &http.Client{Timeout: 30 * time.Second}
}

// RequestHandler manages API requests.
type RequestHandler struct {
	Client       *http.Client
	BaseURL      string
	ApiSecret    string
	ApiSecretKey string
}

// NewRequestHandler initializes a new RequestHandler.
func NewRequestHandler(baseURL, apiSecret, apiSecretKey string, skipTLS bool) *RequestHandler {
	log.Info().Str("base_url", baseURL).Msg("Initializing new RequestHandler.")
	return &RequestHandler{
		Client:       CreateHTTPClient(skipTLS),
		BaseURL:      baseURL,
		ApiSecret:    apiSecret,
		ApiSecretKey: apiSecretKey,
	}
}
func (rh *RequestHandler) PostRequest(endpoint string, payload interface{}) ([]byte, error) {
	// Validate endpoint
	if urlParse, err := url.Parse(endpoint); err == nil && urlParse.IsAbs() {
		log.Error().Str("endpoint", endpoint).Msg("Endpoint must be a relative path, not a full URL.")
		return nil, fmt.Errorf("invalid endpoint: must be a relative path")
	}

	// Construct full URL
	baseURL, err := url.Parse(rh.BaseURL)
	if err != nil {
		log.Error().Err(err).Msg("Invalid base URL.")
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	fullURL := baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, endpoint)}).String()
	log.Debug().Str("base_url", rh.BaseURL).Str("endpoint", endpoint).Str("full_url", fullURL).Msg("Constructed full URL")

	// Serialize payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload.")
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	log.Debug().Str("serialized_payload", string(jsonPayload)).Msg("Serialized payload")

	// Create HTTP request
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request.")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	log.Debug().Str("method", req.Method).Str("urlParse", fullURL).Str("content_type", req.Header.Get("Content-Type")).Msg("HTTP request details")

	// Execute HTTP request
	resp, err := rh.Client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("POST request failed.")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body.")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	log.Debug().Int("status_code", resp.StatusCode).Str("raw_response", string(responseBody)).Msg("Received response")

	// Handle non-OK status codes
	if resp.StatusCode != http.StatusOK {
		log.Warn().Str("urlParse", fullURL).Int("status_code", resp.StatusCode).Msg("Non-OK response received.")
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	log.Info().Str("urlParse", fullURL).Msg("POST request successful.")
	return responseBody, nil
}
