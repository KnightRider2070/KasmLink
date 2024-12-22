package sessions

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api/base"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"
)

type SessionService struct {
	*base.BaseService
}

func NewSessionService(handler http.RequestHandler) *SessionService {
	log.Info().Msg("Creating new SessionService.")
	return &SessionService{
		BaseService: base.NewBaseService(handler),
	}
}

const (
	RequestSessionEndpoint   = "/api/public/request_kasm"
	GetSessionStatusEndpoint = "/api/public/get_kasm_status"
	DestroySessionEndpoint   = "/api/public/destroy_kasm"
	ExecCommandEndpoint      = "/api/public/exec_command_kasm"
)

// RequestSession starts a new session.
func (ss *SessionService) RequestSession(req models.RequestKasm) (*models.RequestKasmResponse, error) {
	url := fmt.Sprintf("%s%s", ss.BaseURL, RequestSessionEndpoint)
	log.Info().
		Str("url", url).
		Str("user_id", req.UserID).
		Msg("Requesting new session.")

	payload := ss.BuildPayload(map[string]interface{}{
		"request": req,
	})

	var sessionResponse models.RequestKasmResponse
	if err := ss.ExecuteRequest(url, payload, &sessionResponse); err != nil {
		log.Error().Err(err).Msg("Failed to request session.")
		return nil, err
	}

	log.Info().Str("kasm_id", sessionResponse.KasmID).Msg("Session requested successfully.")
	return &sessionResponse, nil
}

func (ss *SessionService) GetKasmStatus(req models.GetKasmStatus) (*models.GetKasmStatusResponse, error) {
	url := fmt.Sprintf("%s%s", ss.BaseURL, GetSessionStatusEndpoint)
	log.Info().
		Str("url", url).
		Str("kasm_id", req.KasmID).
		Msg("Fetching Kasm session status.")

	payload := ss.BuildPayload(map[string]interface{}{
		"request": req,
	})

	var statusResponse models.GetKasmStatusResponse
	if err := ss.ExecuteRequest(url, payload, &statusResponse); err != nil {
		log.Error().Err(err).Msg("Failed to fetch session status.")
		return nil, err
	}

	log.Info().
		Str("kasm_id", statusResponse.Kasm.KasmID).
		Msg("Kasm session status retrieved successfully.")
	return &statusResponse, nil
}

func (ss *SessionService) DestroyKasmSession(req models.DestroyKasmRequest) error {
	url := fmt.Sprintf("%s%s", ss.BaseURL, DestroySessionEndpoint)
	log.Info().
		Str("url", url).
		Str("kasm_id", req.KasmID).
		Str("user_id", req.UserID).
		Msg("Destroying Kasm session.")

	payload := ss.BuildPayload(map[string]interface{}{
		"request": req,
	})

	if err := ss.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().Err(err).Msg("Failed to destroy Kasm session.")
		return err
	}

	log.Info().Str("kasm_id", req.KasmID).Msg("Kasm session destroyed successfully.")
	return nil
}

func (ss *SessionService) ExecCommand(req models.ExecCommandRequest) error {
	url := fmt.Sprintf("%s%s", ss.BaseURL, ExecCommandEndpoint)
	log.Info().
		Str("url", url).
		Str("kasm_id", req.KasmID).
		Msg("Executing command in Kasm session.")

	payload := ss.BuildPayload(map[string]interface{}{
		"request": req,
	})

	if err := ss.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().Err(err).Msg("Failed to execute command in Kasm session.")
		return err
	}

	log.Info().Str("kasm_id", req.KasmID).Msg("Command executed successfully in Kasm session.")
	return nil
}
