package images

import (
	"fmt"
	"kasmlink/pkg/api/base"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"

	"github.com/rs/zerolog/log"
)

const (
	ListImagesEndpoint = "/api/public/get_images"
)

// ImageService provides methods to interact with the images API.
type ImageService struct {
	*base.BaseService
}

// NewImageService creates a new instance of ImageService.
func NewImageService(handler http.RequestHandler) *ImageService {
	log.Info().Msg("Creating new ImageService.")
	return &ImageService{
		BaseService: base.NewBaseService(handler),
	}
}

// ListImages fetches the available images from the API.
func (is *ImageService) ListImages() ([]models.Image, error) {
	url := fmt.Sprintf("%s%s", is.BaseURL, ListImagesEndpoint)
	log.Info().
		Str("url", url).
		Msg("Fetching list of images.")

	payload := is.BuildPayload(nil)

	var imagesResponse models.GetImagesResponse
	if err := is.ExecuteRequest(url, payload, &imagesResponse); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to fetch images.")
		return nil, err
	}

	log.Info().
		Int("image_count", len(imagesResponse.Images)).
		Str("url", url).
		Msg("Successfully fetched images.")
	return imagesResponse.Images, nil
}
