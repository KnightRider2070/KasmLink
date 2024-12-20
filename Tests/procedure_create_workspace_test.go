package Tests

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/procedures"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

func TestCreateKasmWorkspace(t *testing.T) {
	// Create a Kasm API client
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	// Create a context for the request
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()

	// Define the image and user details for the workspace to create
	imageDetail := webApi.ImageDetail{
		Name:         "kasmweb/firefox:1.15.0-rolling", // Use an image that exists or you want to create
		Cores:        2,
		Memory:       2048,
		FriendlyName: "Ubuntu Test Workspace",
	}

	// Provide volume mounts as map[string]string
	details := userParser.UserDetails{
		VolumeMounts: map[string]string{
			"/tmp": "/container:rw",
		},
		EnvironmentArgs: map[string]string{
			"EXAMPLE_ENV_VAR": "test-value",
		},
		Network: "testi", // A valid network that exists in your Kasm environment
	}

	// Call the function under test
	err := procedures.CreateKasmWorkspace(ctx, imageDetail, details, kApi)

	// Log and assert results
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kasm Workspace")
	} else {
		log.Info().Msg("Workspace created successfully")
	}

	assert.NoError(t, err, "Expected no error from createKasmWorkspace")

	// If you want, verify that the image was actually created by listing images:
	imagesAvailable, listErr := kApi.ListImages(ctx)
	if listErr != nil {
		log.Debug().Err(listErr).Msg("Could not list images after workspace creation")
	}
	assert.NoError(t, listErr, "Expected no error listing images after workspace creation")

	// Check that the created image is among the returned images
	found := false
	for _, img := range imagesAvailable {
		if img.FriendlyName == "ubuntu" {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected to find the created 'ubuntu' image in the list of images")
}
