package SystemTests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/sessions"
	"testing"
	"time"
)

func TestRequestKasm(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, apiSecret, apiKeySecret, true)
	kApi := sessions.NewSessionService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userID := "a2b9d2932e484280bc0a64822a5c8d42"  // testuser
	imageID := "6a335ca1505a4e0eb966930823bcc691" // Brave
	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	req := models.RequestKasm{
		UserID:      userID,
		ImageID:     imageID,
		Environment: envArgs,
	}

	kasmResponse, err := kApi.RequestSession(req)
	assert.NoError(t, err)
	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Equal(t, userID, kasmResponse.UserID)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)
}

func TestGetKasmStatus(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, apiSecret, apiKeySecret, true)
	kApi := sessions.NewSessionService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userID := "a2b9d2932e484280bc0a64822a5c8d42"  // testuser
	imageID := "6a335ca1505a4e0eb966930823bcc691" // Brave
	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	req := models.RequestKasm{
		UserID:      userID,
		ImageID:     imageID,
		Environment: envArgs,
	}

	kasmResponse, err := kApi.RequestSession(req)
	assert.NoError(t, err)
	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Equal(t, userID, kasmResponse.UserID)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)

	statusReq := models.GetKasmStatus{
		UserID:         userID,
		KasmID:         kasmResponse.KasmID,
		SkipAgentCheck: true,
	}

	status, err := kApi.GetKasmStatus(statusReq)
	assert.NoError(t, err)
	assert.Equal(t, imageID, status.Kasm.ImageID)
}

func TestDestroyKasmSession(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, apiSecret, apiKeySecret, true)
	kApi := sessions.NewSessionService(*handler)

	// Create context
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userID := "a2b9d2932e484280bc0a64822a5c8d42"  // testuser
	imageID := "6a335ca1505a4e0eb966930823bcc691" // Brave
	envArgs := map[string]string{
		"ENV_VAR": "value",
	}

	req := models.RequestKasm{
		UserID:      userID,
		ImageID:     imageID,
		Environment: envArgs,
	}

	kasmResponse, err := kApi.RequestSession(req)
	assert.NoError(t, err)
	assert.Contains(t, kasmResponse.Status, "starting")
	assert.Equal(t, userID, kasmResponse.UserID)
	assert.NotEmpty(t, kasmResponse.KasmID)
	assert.NotEmpty(t, kasmResponse.SessionToken)

	statusReq := models.GetKasmStatus{
		UserID:         userID,
		KasmID:         kasmResponse.KasmID,
		SkipAgentCheck: true,
	}

	status, err := kApi.GetKasmStatus(statusReq)
	assert.NoError(t, err)
	assert.Equal(t, imageID, status.Kasm.ImageID)

	destroyReq := models.DestroyKasmRequest{
		UserID: userID,
		KasmID: kasmResponse.KasmID,
	}

	err = kApi.DestroyKasmSession(destroyReq)
	assert.NoError(t, err)
}

func TestExecCommand(t *testing.T) {
}
