package SystemTests

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/workspace"
	"testing"
	"time"
)

func TestCreateKasmWorkspace(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, apiSecret, apiKeySecret, true)

	// Initialize WorkspaceService
	workspaceService := workspace.NewWorkspaceService(*handler)

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	runConfig := models.DockerRunConfig{
		Name:       "test",
		Hostname:   "test",
		Network:    "test",
		CapAdd:     []string{"NET_ADMIN"},
		Privileged: false,
		Init:       true,
		Environment: map[string]string{
			"DISPLAY": ":1",
		},
	}

	runJSON, err := json.Marshal(runConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize volume_mappings")
	}

	volumeMappings := map[string]models.VolumeMapping{
		"/mnt/kasm_user_share": {
			Bind: "/share",
			Mode: "rw",
			Gid:  1000,
			Uid:  1000,
		},
	}

	volumeMappingsJSON, err := json.Marshal(volumeMappings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize volume_mappings")
	}

	// Minimal volume_mappings
	workspaceDetail := models.TargetImage{
		FriendlyName:    "Ubuntu Test Workspace",
		DockerImageName: "kasmweb/firefox:1.15.0-rolling",
		ImageType:       "Container",
		RunConfig:       string(runJSON),
		VolumeMappings:  string(volumeMappingsJSON),
	}

	// Attempt to create the workspace
	createdWorkspace, err := workspaceService.CreateWorkspace(workspaceDetail)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kasm Workspace")
		assert.Fail(t, "Expected no error from CreateKasmWorkspace", err)
		return
	}
	log.Info().Str("workspace_id", createdWorkspace.ImageID).Msg("Workspace created successfully")
	assert.NotNil(t, createdWorkspace, "Expected a valid workspace to be created")
}
