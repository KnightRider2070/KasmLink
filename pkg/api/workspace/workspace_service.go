package workspace

import (
	"fmt"
	"kasmlink/pkg/api/base"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"

	"github.com/rs/zerolog/log"
)

const (
	CreateWorkspaceEndpoint = "/api/public/create_image"
	UpdateWorkspaceEndpoint = "/api/public/update_image"
	DeleteWorkspaceEndpoint = "/api/public/delete_image"
	GetWorkspaceEndpoint    = "/api/public/get_image"
	GetWorkspacesEndpoint   = "/api/public/get_images"
)

// WorkspaceService provides methods to manage workspaces.
type WorkspaceService struct {
	*base.BaseService
}

// NewWorkspaceService creates a new instance of WorkspaceService.
func NewWorkspaceService(handler http.RequestHandler) *WorkspaceService {
	return &WorkspaceService{
		BaseService: base.NewBaseService(handler),
	}
}

// CreateWorkspace sends a request to create a workspace.
func (ws *WorkspaceService) CreateWorkspace(workspace models.TargetImage) (*models.TargetImage, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, CreateWorkspaceEndpoint)
	log.Info().Str("url", url).Str("workspace_name", workspace.FriendlyName).Msg("Creating new workspace.")

	payload := ws.BuildPayload(map[string]interface{}{
		"target_image": workspace,
	})

	var createdWorkspace models.TargetImage
	if err := ws.ExecuteRequest(CreateWorkspaceEndpoint, payload, &createdWorkspace); err != nil {
		log.Error().Err(err).Msg("Failed to create workspace.")
		return nil, err
	}

	log.Info().Str("workspace_id", createdWorkspace.ImageID).Msg("Workspace created successfully.")
	return &createdWorkspace, nil
}

// UpdateWorkspace updates an existing workspace's details.
func (ws *WorkspaceService) UpdateWorkspace(workspace models.TargetImage) (*models.TargetImage, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, UpdateWorkspaceEndpoint)
	log.Info().
		Str("url", url).
		Str("workspace_id", workspace.ImageID).
		Msg("Updating workspace details.")

	payload := ws.BuildPayload(map[string]interface{}{
		"workspace": workspace,
	})

	var updatedWorkspace models.TargetImage
	if err := ws.ExecuteRequest(UpdateWorkspaceEndpoint, payload, &updatedWorkspace); err != nil {
		log.Error().Err(err).Str("workspace_id", workspace.ImageID).Msg("Failed to update workspace.")
		return nil, err
	}

	log.Info().Str("workspace_id", updatedWorkspace.ImageID).Msg("Workspace updated successfully.")
	return &updatedWorkspace, nil
}

// DeleteWorkspace removes a workspace by workspace ID.
func (ws *WorkspaceService) DeleteWorkspace(workspaceID string) error {
	url := fmt.Sprintf("%s%s", ws.BaseURL, DeleteWorkspaceEndpoint)
	log.Info().
		Str("url", url).
		Str("workspace_id", workspaceID).
		Msg("Deleting workspace.")

	payload := ws.BuildPayload(map[string]interface{}{
		"workspace": map[string]string{
			"workspace_id": workspaceID,
		},
	})

	if err := ws.ExecuteRequest(DeleteWorkspaceEndpoint, payload, nil); err != nil {
		log.Error().Err(err).Str("workspace_id", workspaceID).Msg("Failed to delete workspace.")
		return err
	}

	log.Info().Str("workspace_id", workspaceID).Msg("Workspace deleted successfully.")
	return nil
}

// GetWorkspace retrieves workspace details by workspace ID.
func (ws *WorkspaceService) GetWorkspace(workspaceID string) (*models.TargetImage, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, GetWorkspaceEndpoint)
	log.Info().
		Str("url", url).
		Str("workspace_id", workspaceID).
		Msg("Fetching workspace details.")

	payload := ws.BuildPayload(map[string]interface{}{
		"workspace": map[string]string{
			"workspace_id": workspaceID,
		},
	})

	var workspace models.TargetImage
	if err := ws.ExecuteRequest(GetWorkspaceEndpoint, payload, &workspace); err != nil {
		log.Error().Err(err).Str("workspace_id", workspaceID).Msg("Failed to fetch workspace details.")
		return nil, err
	}

	log.Info().Str("workspace_id", workspace.ImageID).Msg("Workspace details retrieved successfully.")
	return &workspace, nil
}

// GetWorkspaces retrieves a list of all workspaces.
func (ws *WorkspaceService) GetWorkspaces() ([]models.TargetImage, error) {
	endpoint := GetWorkspacesEndpoint
	log.Info().
		Str("endpoint", endpoint).
		Msg("Fetching list of images.")

	// Build the payload
	payload := ws.BuildPayload(nil)

	// Initialize the response object
	var imagesResponse models.GetWorkspaceResponse
	if err := ws.ExecuteRequest(endpoint, payload, &imagesResponse); err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to fetch images.")
		return nil, err
	}

	log.Info().Int("image_count", len(imagesResponse.Images)).Str("endpoint", endpoint).Msg("Successfully fetched images.")

	return imagesResponse.Images, nil
}
