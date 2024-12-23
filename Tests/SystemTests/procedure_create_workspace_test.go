package SystemTests

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/internal"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

func TestCreateKasmWorkspace(t *testing.T) {
	// Create a Kasm API client
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	// Create a context for the request
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// Define the image and user details for the workspace to create
	imageDetail := webApi.ImageDetail{
		Name:         "kasmweb/firefox:1.15.0-rolling", // Ensure this image exists in your Kasm environment
		Cores:        2,
		Memory:       2048, // Adjust this value if needed to match the Kasm API's expected unit
		FriendlyName: "Ubuntu Test Workspace",
		Description:  "Test workspace for Firefox",
	}

	// Provide volume mounts as map[string]string
	details := userParser.UserDetails{
		VolumeMounts: map[string]string{
			"/tmp": "/container:rw", // Ensure this path is valid for your environment
		},
		EnvironmentArgs: map[string]string{
			"EXAMPLE_ENV_VAR": "test-value",
		},
		Network: "test", // Use a valid network that exists in your Kasm configuration
	}

	// Call the function under test
	err := internal.CreateKasmWorkspace(ctx, imageDetail, details, kApi)

	// Log and assert results
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kasm Workspace")
		assert.Fail(t, "Expected no error from CreateKasmWorkspace", err)
		return
	}
	log.Info().Msg("Workspace created successfully")
	assert.NoError(t, err, "Expected no error from CreateKasmWorkspace")

	// Verify that the image was actually created by listing images
	imagesAvailable, listErr := kApi.ListImages(ctx)
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
