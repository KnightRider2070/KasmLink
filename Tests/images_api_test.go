package Tests

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

func TestListImages(t *testing.T) {

	//Create KASM API
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 50*time.Second)

	//Create context
	ctx, _ := context.WithTimeout(context.Background(), 10000*time.Second)

	imagesAvailable, err := kApi.ListImages(ctx)

	if err != nil {
		// No error
		log.Debug().Int("Available image count", len(imagesAvailable)).Msg("Available Images on Kasm")
	}

	assert.NoError(t, err)
	assert.Equal(t, 1, len(imagesAvailable))
}
