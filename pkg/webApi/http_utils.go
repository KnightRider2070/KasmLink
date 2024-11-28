package webApi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// HandleResponse reads the response body and checks for errors or unexpected status codes
func HandleResponse(resp *http.Response, expectedStatusCode int) ([]byte, error) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().
				Err(err).
				Str("url", resp.Request.URL.String()).
				Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", resp.Request.URL.String()).
			Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != expectedStatusCode {
		log.Warn().
			Str("url", resp.Request.URL.String()).
			Int("status_code", resp.StatusCode).
			Str("response_body", strings.TrimSpace(string(body))).
			Msg("Unexpected response status")
		return nil, fmt.Errorf("unexpected response status: %s, body: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	if len(body) == 0 {
		log.Warn().
			Str("url", resp.Request.URL.String()).
			Msg("Response body is empty")
		return nil, fmt.Errorf("response body is empty")
	}

	log.Info().
		Str("url", resp.Request.URL.String()).
		Msg("Request succeeded")
	return body, nil
}

// MakeGetRequest handles making GET requests to the KASM API.
// It now accepts a context for better request management.
func (api *KasmAPI) MakeGetRequest(ctx context.Context, endpoint string, queryParams map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", api.BaseURL, endpoint)
	if len(queryParams) > 0 {
		// Append query parameters to URL
		query := "?"
		for key, value := range queryParams {
			query += fmt.Sprintf("%s=%s&", key, value)
		}
		url += strings.TrimSuffix(query, "&")
	}

	log.Info().
		Str("method", "GET").
		Str("url", url).
		Msg("Initiating GET request")

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "GET").
			Str("url", url).
			Msg("Failed to create GET request")
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	// Set Authorization header with both APIKey and APIKeySecret if required
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", api.APIKey, api.APIKeySecret))

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := api.Client.Do(req)
		if err != nil {
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "GET").
				Str("url", url).
				Msg("GET request failed, retrying")
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}

		body, err := HandleResponse(resp, http.StatusOK)
		if err != nil {
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "GET").
				Str("url", url).
				Msg("GET request returned unexpected status, retrying")
			lastErr = err
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("GET request to %s failed after retries: %w", url, lastErr)
}

// MakePostRequest handles making POST requests to the KASM API.
// It accepts a context for request cancellation, an endpoint path, and a payload.
// Returns the response body as bytes if the request is successful.
func (api *KasmAPI) MakePostRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", api.BaseURL, endpoint)

	log.Debug().
		Str("method", "POST").
		Str("url", url).
		Msg("Preparing POST request payload")

	// Convert payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("url", url).
			Msg("Failed to marshal payload for POST request")
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("url", url).
		Msg("Initiating POST request")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("url", url).
			Msg("Failed to create POST request")
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Set Authorization header with both APIKey and APIKeySecret if required
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", api.APIKey, api.APIKeySecret))

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := api.Client.Do(req)
		if err != nil {
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "POST").
				Str("url", url).
				Msg("POST request failed, retrying")
			lastErr = err
			sleepDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
			time.Sleep(sleepDuration + jitter) // Exponential backoff with jitter
			continue
		}

		responseBody, err := HandleResponse(resp, http.StatusOK)
		if err != nil {
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "POST").
				Str("url", url).
				Msg("POST request returned unexpected status, retrying")
			lastErr = err
			sleepDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
			time.Sleep(sleepDuration + jitter)
			continue
		}

		return responseBody, nil
	}

	return nil, fmt.Errorf("POST request to %s failed after retries: %w", url, lastErr)
}
