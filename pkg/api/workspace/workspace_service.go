package workspace

import (
	"fmt"
	"kasmlink/pkg/api/base"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"

	"github.com/rs/zerolog/log"
)

const (
	CreateWorkspaceEndpoint = "/api/workspace/create"
	UpdateWorkspaceEndpoint = "/api/workspace/update"
	DeleteWorkspaceEndpoint = "/api/workspace/delete"
	GetWorkspaceEndpoint    = "/api/workspace/get"
	GetWorkspacesEndpoint   = "/api/workspace/list"
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
func (ws *WorkspaceService) CreateWorkspace(workspace models.TargetImage) (*models.ImageDetail, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, CreateWorkspaceEndpoint)
	log.Info().Str("url", url).Str("workspace_name", workspace.FriendlyName).Msg("Creating new workspace.")

	payload := ws.BuildPayload(map[string]interface{}{
		"workspace": workspace,
	})

	var createdWorkspace models.ImageDetail
	if err := ws.ExecuteRequest(url, payload, &createdWorkspace); err != nil {
		log.Error().Err(err).Msg("Failed to create workspace.")
		return nil, err
	}

	log.Info().Str("workspace_id", createdWorkspace.ImageID).Msg("Workspace created successfully.")
	return &createdWorkspace, nil
}

// UpdateWorkspace updates an existing workspace's details.
func (ws *WorkspaceService) UpdateWorkspace(workspace models.TargetImage) (*models.ImageDetail, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, UpdateWorkspaceEndpoint)
	log.Info().
		Str("url", url).
		Str("workspace_id", workspace.ImageID).
		Msg("Updating workspace details.")

	payload := ws.BuildPayload(map[string]interface{}{
		"workspace": workspace,
	})

	var updatedWorkspace models.ImageDetail
	if err := ws.ExecuteRequest(url, payload, &updatedWorkspace); err != nil {
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

	if err := ws.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().Err(err).Str("workspace_id", workspaceID).Msg("Failed to delete workspace.")
		return err
	}

	log.Info().Str("workspace_id", workspaceID).Msg("Workspace deleted successfully.")
	return nil
}

// GetWorkspace retrieves workspace details by workspace ID.
func (ws *WorkspaceService) GetWorkspace(workspaceID string) (*models.ImageDetail, error) {
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

	var workspace models.ImageDetail
	if err := ws.ExecuteRequest(url, payload, &workspace); err != nil {
		log.Error().Err(err).Str("workspace_id", workspaceID).Msg("Failed to fetch workspace details.")
		return nil, err
	}

	log.Info().Str("workspace_id", workspace.ImageID).Msg("Workspace details retrieved successfully.")
	return &workspace, nil
}

// GetWorkspaces retrieves a list of all workspaces.
func (ws *WorkspaceService) GetWorkspaces() ([]models.ImageDetail, error) {
	url := fmt.Sprintf("%s%s", ws.BaseURL, GetWorkspacesEndpoint)
	log.Info().Str("url", url).Msg("Fetching all workspaces.")

	payload := ws.BuildPayload(nil)

	var parsedResponse struct {
		Workspaces []models.ImageDetail `json:"workspaces"`
	}
	if err := ws.ExecuteRequest(url, payload, &parsedResponse); err != nil {
		log.Error().Err(err).Msg("Failed to fetch workspaces.")
		return nil, err
	}

	log.Info().Int("workspace_count", len(parsedResponse.Workspaces)).Msg("Workspaces retrieved successfully.")
	return parsedResponse.Workspaces, nil
}
