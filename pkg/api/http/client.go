package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func NewRequestHandler(baseURL string, skipTLS bool) *RequestHandler {
	log.Info().Str("base_url", baseURL).Msg("Initializing new RequestHandler.")
	return &RequestHandler{
		Client:  CreateHTTPClient(skipTLS),
		BaseURL: baseURL,
	}
}

// PostRequest sends a POST request and returns the response body.
func (rh *RequestHandler) PostRequest(endpoint string, payload interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", rh.BaseURL, endpoint)
	log.Debug().Str("url", url).Msg("Sending POST request.")

	// Marshal the payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload.")
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request.")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := rh.Client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("POST request failed.")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body.")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		log.Warn().Str("url", url).Int("status_code", resp.StatusCode).Msg("Non-OK response received.")
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	log.Info().Str("url", url).Msg("POST request successful.")
	return responseBody, nil
}
