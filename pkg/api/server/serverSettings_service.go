package server

import (
	"fmt"
	"kasmlink/pkg/api/http"

	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api/base"
)

const (
	UpdateSettingEndpoint = "/api/public/update_setting"
)

// Setting represents the structure of a setting update.
type Setting struct {
	SettingID string `json:"setting_id"`
	Value     bool   `json:"value"`
	Token     string `json:"token"`
	Username  string `json:"username"`
}

// ServerSettingsService provides methods to manage server settings.
type ServerSettingsService struct {
	*base.BaseService
}

// NewServerSettingsService creates a new ServerSettingsService instance.
func NewServerSettingsService(handler http.RequestHandler) *ServerSettingsService {
	log.Info().Msg("Creating new ServerSettingsService.")
	return &ServerSettingsService{
		BaseService: base.NewBaseService(handler),
	}
}

// UpdateSetting updates the server setting using the BaseService.
func (s *ServerSettingsService) UpdateSetting(setting Setting) error {
	// Construct the payload using BaseService
	payload := s.BuildPayload(map[string]interface{}{
		"setting_id": setting.SettingID,
		"value":      setting.Value,
		"username":   setting.Username,
	})

	// Log the operation
	log.Info().
		Str("setting_id", setting.SettingID).
		Str("username", setting.Username).
		Msg("Sending request to update setting")

	// Execute the request
	err := s.ExecuteRequest(UpdateSettingEndpoint, payload, nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("endpoint", UpdateSettingEndpoint).
			Msg("Failed to update setting")
		return fmt.Errorf("failed to update setting: %w", err)
	}

	log.Info().
		Str("setting_id", setting.SettingID).
		Msg("Setting updated successfully")
	return nil
}

func (s *ServerSettingsService) UpdateAddWorkspaceToAllGroupsVar(enabled bool) error {

	setting := Setting{
		SettingID: "7ca1822daf0545ad8ad5428a05b8405d",
		Value:     enabled,
	}

	return s.UpdateSetting(setting)

}
