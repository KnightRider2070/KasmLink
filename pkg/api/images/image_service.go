package images

import (
	"context"
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

// CreateImage sends a POST request to create an image.
func (is *ImageService) CreateImage(ctx context.Context, targetImage models.Image) (*models.Image, error) {
	url := fmt.Sprintf("%s/api/public/create_image", is.BaseURL)
	log.Info().Str("url", url).Msg("Creating a new image.")

	payload := is.BuildPayload(map[string]interface{}{
		"target_image": targetImage,
	})

	var response struct {
		Image models.Image `json:"image"`
	}

	if err := is.ExecuteRequest(url, payload, &response); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to create image.")
		return nil, err
	}

	log.Info().Str("image_id", response.Image.ImageID).Msg("Image created successfully.")
	return &response.Image, nil
}

// UpdateImage sends a POST request to update an image.
func (is *ImageService) UpdateImage(ctx context.Context, targetImage models.Image) (*models.Image, error) {
	if targetImage.ImageID == "" {
		return nil, fmt.Errorf("image_id must be set in TargetImage before calling UpdateImage")
	}

	url := fmt.Sprintf("%s/api/public/update_image", is.BaseURL)
	log.Info().Str("url", url).Msg("Updating an existing image.")

	payload := is.BuildPayload(map[string]interface{}{
		"target_image": targetImage,
	})

	var response struct {
		Image models.Image `json:"image"`
	}

	if err := is.ExecuteRequest(url, payload, &response); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to update image.")
		return nil, err
	}

	log.Info().Str("image_id", response.Image.ImageID).Msg("Image updated successfully.")
	return &response.Image, nil
}

// DeleteImage sends a POST request to delete an image.
func (is *ImageService) DeleteImage(ctx context.Context, imageID string) error {
	if imageID == "" {
		return fmt.Errorf("image_id must be provided")
	}

	url := fmt.Sprintf("%s/api/public/delete_image", is.BaseURL)
	log.Info().Str("url", url).Str("image_id", imageID).Msg("Deleting an image.")

	payload := is.BuildPayload(map[string]interface{}{
		"target_image": map[string]string{
			"image_id": imageID,
		},
	})

	if err := is.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to delete image.")
		return err
	}

	log.Info().Str("image_id", imageID).Msg("Image deleted successfully.")
	return nil
}
