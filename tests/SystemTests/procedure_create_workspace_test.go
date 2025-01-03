package SystemTests

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/images"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/workspace"
	"testing"
	"time"
)

func TestCreateKasmWorkspace(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)

	// Initialize ImageService and WorkspaceService
	imageService := images.NewImageService(*handler)
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

	// Define the image and username details for the workspace to create
	imageDetail := models.TargetImage{
		FriendlyName:         "Ubuntu Test Workspace",
		Description:          "Test workspace for Firefox",
		Cores:                2,
		Memory:               2048,
		Name:                 "kasmweb/firefox:1.15.0-rolling",
		RestrictNetworkNames: []string{"test"},
		RunConfig:            runConfig,
		Notes:                "Test workspace for Firefox",
	}

	// Attempt to create the workspace
	wkspc, err := workspaceService.CreateWorkspace(imageDetail)

	// Log and assert results
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kasm Workspace")
		assert.Fail(t, "Expected no error from CreateKasmWorkspace", err)
		return
	}
	log.Info().Msg("Workspace created successfully")
	assert.NotNil(t, wkspc, "Expected a valid workspace to be created")

	// Verify that the image was actually created by listing images
	imagesAvailable, listErr := imageService.ListImages()
	if listErr != nil {
		log.Error().Err(listErr).Msg("Could not list images after workspace creation")
		assert.Fail(t, "Expected no error listing images", listErr)
		return
	}
	log.Info().Int("image_count", len(imagesAvailable)).Msg("Images fetched successfully")
	assert.NoError(t, listErr, "Expected no error listing images after workspace creation")

	// Check that the created image is among the returned images
	found := false
	for _, img := range imagesAvailable {
		if img.FriendlyName == imageDetail.FriendlyName {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected to find the created image in the list of images")
}
