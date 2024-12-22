package base

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api/http"
)

// BaseService provides common methods for API services.
type BaseService struct {
	http.RequestHandler
}

// NewBaseService creates a new instance of BaseService.
func NewBaseService(handler http.RequestHandler) *BaseService {
	return &BaseService{
		RequestHandler: handler,
	}
}

// BuildPayload constructs the standard payload with API credentials.
func (bs *BaseService) BuildPayload(extra map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"api_key":        bs.ApiSecret,
		"api_key_secret": bs.ApiSecretKey,
	}
	for key, value := range extra {
		payload[key] = value
	}
	return payload
}

// ExecuteRequest performs a POST request and unmarshals the response.
func (bs *BaseService) ExecuteRequest(url string, payload map[string]interface{}, result interface{}) error {
	response, err := bs.PostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("HTTP request failed.")
		return err
	}

	if result != nil {
		if err := json.Unmarshal(response, result); err != nil {
			log.Error().Err(err).Str("url", url).Msg("Failed to unmarshal response.")
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
