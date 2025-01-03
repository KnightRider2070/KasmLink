package SystemTests

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/images"
	"testing"
	"time"
)

func TestListImages(t *testing.T) {
	// Initialize RequestHandler
	handler := http.NewRequestHandler(baseUrl, true)

	// Initialize ImageService
	imageService := images.NewImageService(*handler)

	// Create context with timeout
	_, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	// Fetch list of available images
	imagesAvailable, err := imageService.ListImages()

	// Log available images or errors
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch available images from Kasm API")
	} else {
		log.Debug().Int("Available image count", len(imagesAvailable)).Msg("Available Images on Kasm")
	}

	// Validate the results
	assert.NoError(t, err)
	assert.NotEmpty(t, imagesAvailable, "Expected at least one image to be available")
}
