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
		trimmedBody := strings.TrimSpace(string(body))
		log.Warn().
			Str("url", resp.Request.URL.String()).
			Int("status_code", resp.StatusCode).
			Str("response_body", trimmedBody).
			Msg("Unexpected response status")

		return nil, fmt.Errorf("unexpected response status: %s, body: %s",
			resp.Status, trimmedBody)
	}

	if len(body) == 0 {
		log.Warn().
			Str("url", resp.Request.URL.String()).
			Msg("Response body is empty")
		return nil, fmt.Errorf("response body is empty")
	}

	log.Info().
		Str("url", resp.Request.URL.String()).
		Int("status_code", resp.StatusCode).
		Msg("Request succeeded")

	// If the response is JSON, you can log it as raw JSON:
	log.Debug().
		Str("url", resp.Request.URL.String()).
		Int("status_code", resp.StatusCode).
		RawJSON("response_body", body).
		Msg("Response details")

	return body, nil
}

// MakeGetRequest handles making GET requests to the KASM API.
// It now accepts a context for better request management.
func (api *KasmAPI) MakeGetRequest(ctx context.Context, endpoint string, queryParams map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", api.BaseURL, endpoint)
	if len(queryParams) > 0 {
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

	// Set Authorization header if required
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", api.APIKey, api.APIKeySecret))

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := api.Client.Do(req)
		if err != nil {
			backoff := time.Duration(attempt) * time.Second
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "GET").
				Str("url", url).
				Dur("backoff", backoff).
				Msg("GET request failed, will retry")

			lastErr = err
			time.Sleep(backoff)
			continue
		}

		body, err := HandleResponse(resp, http.StatusOK)
		if err != nil {
			backoff := time.Duration(attempt) * time.Second
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "GET").
				Str("url", url).
				Dur("backoff", backoff).
				Msg("GET request returned unexpected status, will retry")

			lastErr = err
			time.Sleep(backoff)
			continue
		}

		log.Debug().
			Str("method", "GET").
			Str("url", url).
			RawJSON("response_body", body).
			Msg("Received successful response")

		return body, nil
	}

	return nil, fmt.Errorf("GET request to %s failed after retries: %w", url, lastErr)
}

// MakePostRequest handles making POST requests to the KASM API.
// It accepts a context for request cancellation, an endpoint path, and a payload.
// Returns the response body as bytes if the request is successful.
func (api *KasmAPI) MakePostRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", api.BaseURL, endpoint)

	// Marshal payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to marshal payload for POST request")
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Log payload as structured data
	log.Debug().
		Str("method", "POST").
		Str("url", url).
		RawJSON("payload", body).
		Msg("Sending POST request")

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("Failed to create POST request")
			return nil, fmt.Errorf("failed to create POST request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s:%s", api.APIKey, api.APIKeySecret))

		resp, err := api.Client.Do(req)
		if err != nil {
			backoff := time.Second*time.Duration(math.Pow(2, float64(attempt))) + time.Millisecond*time.Duration(rand.Intn(1000))
			log.Error().
				Err(err).
				Int("attempt", attempt).
				Str("method", "POST").
				Str("url", url).
				Dur("backoff", backoff).
				Msg("POST request failed, retrying")
			lastErr = err
			time.Sleep(backoff)
			continue
		}

		responseBody, err := HandleResponse(resp, http.StatusOK)
		if err != nil {
			backoff := time.Second*time.Duration(math.Pow(2, float64(attempt))) + time.Millisecond*time.Duration(rand.Intn(1000))
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Str("method", "POST").
				Str("url", url).
				Dur("backoff", backoff).
				Msg("POST request returned unexpected status, retrying")
			lastErr = err
			time.Sleep(backoff)
			continue
		}

		log.Debug().
			Str("method", "POST").
			Str("url", url).
			RawJSON("response_body", responseBody).
			Msg("Received successful response")

		return responseBody, nil
	}

	return nil, fmt.Errorf("POST request to %s failed after retries: %w", url, lastErr)
}
