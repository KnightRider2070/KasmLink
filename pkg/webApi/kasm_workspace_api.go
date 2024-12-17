package webApi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

//NOTE: Using undocumented API endpoints. This might require changes for new versions of Kasm.
// Reference for further information https://kasmweb.atlassian.net/servicedesk/customer/portal/3/article/10682377

// TargetImage represents the structure for the "target_image" object used
// in create, update, and other image-related requests.
type TargetImage struct {
	AllowNetworkSelection  bool            `json:"allow_network_selection,omitempty"`
	Categories             string          `json:"categories,omitempty"`
	Cores                  float64         `json:"cores"`
	CPUAllocationMethod    string          `json:"cpu_allocation_method"`
	Description            string          `json:"description"`
	DockerRegistry         string          `json:"docker_registry,omitempty"`
	DockerToken            string          `json:"docker_token,omitempty"`
	DockerUser             string          `json:"docker_user,omitempty"`
	Enabled                bool            `json:"enabled"`
	ExecConfig             string          `json:"exec_config,omitempty"`
	FilterPolicyID         *string         `json:"filter_policy_id,omitempty"`
	FriendlyName           string          `json:"friendly_name"`
	GPUCount               float64         `json:"gpu_count"`
	Hash                   string          `json:"hash,omitempty"`
	Hidden                 bool            `json:"hidden,omitempty"`
	ImageID                string          `json:"image_id,omitempty"`
	ImageSrc               *string         `json:"image_src,omitempty"`
	ImageType              string          `json:"image_type"`
	IsRemoteApp            bool            `json:"is_remote_app,omitempty"`
	LaunchConfig           json.RawMessage `json:"launch_config,omitempty"`
	LinkURL                *string         `json:"link_url,omitempty"`
	Memory                 int64           `json:"memory"`
	Name                   string          `json:"name"`
	Notes                  string          `json:"notes,omitempty"`
	OverrideEgressGateways bool            `json:"override_egress_gateways,omitempty"`
	PersistentProfilePath  *string         `json:"persistent_profile_path,omitempty"`
	RDPClientType          *string         `json:"rdp_client_type,omitempty"`
	RemoteAppArgs          *string         `json:"remote_app_args,omitempty"`
	RemoteAppName          *string         `json:"remote_app_name,omitempty"`
	RemoteAppProgram       *string         `json:"remote_app_program,omitempty"`
	RequireGPU             bool            `json:"require_gpu,omitempty"`
	RestrictNetworkNames   []string        `json:"restrict_network_names,omitempty"`
	RestrictToNetwork      bool            `json:"restrict_to_network,omitempty"`
	RestrictToServer       bool            `json:"restrict_to_server,omitempty"`
	RestrictToZone         bool            `json:"restrict_to_zone,omitempty"`
	RunConfig              string          `json:"run_config,omitempty"`
	ServerID               string          `json:"server_id,omitempty"`
	ServerPoolID           *string         `json:"server_pool_id,omitempty"`
	SessionTimeLimit       string          `json:"session_time_limit,omitempty"`
	UncompressedSizeMB     int             `json:"uncompressed_size_mb,omitempty"`
	VolumeMappings         string          `json:"volume_mappings,omitempty"`
	ZoneID                 string          `json:"zone_id,omitempty"`
}

// CreateImageRequest represents the request structure for creating/updating an image.
// Now includes APIKey and APIKeySecret.
type CreateImageRequest struct {
	APIKey       string      `json:"api_key"`
	APIKeySecret string      `json:"api_key_secret"`
	TargetImage  TargetImage `json:"target_image"`
}

// ImageDetail represents the structure of the "image" object in the response.
type ImageDetail struct {
	ImageID                   string                 `json:"image_id"`
	Cores                     float64                `json:"cores"`
	Description               string                 `json:"description"`
	DockerRegistry            *string                `json:"docker_registry,omitempty"`
	DockerToken               *string                `json:"docker_token,omitempty"`
	DockerUser                *string                `json:"docker_user,omitempty"`
	Enabled                   bool                   `json:"enabled"`
	FriendlyName              string                 `json:"friendly_name"`
	Hash                      *string                `json:"hash,omitempty"`
	Memory                    int64                  `json:"memory"`
	Name                      string                 `json:"name"`
	XRes                      int                    `json:"x_res"`
	YRes                      int                    `json:"y_res"`
	ImageAttributes           []string               `json:"imageAttributes,omitempty"`
	RestrictToNetwork         bool                   `json:"restrict_to_network"`
	RestrictNetworkNames      []string               `json:"restrict_network_names,omitempty"`
	RestrictToServer          bool                   `json:"restrict_to_server"`
	ServerID                  *string                `json:"server_id,omitempty"`
	PersistentProfileConfig   map[string]interface{} `json:"persistent_profile_config,omitempty"`
	VolumeMappings            map[string]interface{} `json:"volume_mappings,omitempty"`
	RunConfig                 DockerRunConfig        `json:"run_config,omitempty"`
	ImageSrc                  *string                `json:"image_src,omitempty"`
	Available                 bool                   `json:"available"`
	ExecConfig                map[string]interface{} `json:"exec_config,omitempty"`
	RestrictToZone            bool                   `json:"restrict_to_zone"`
	ZoneID                    *string                `json:"zone_id,omitempty"`
	ZoneName                  *string                `json:"zone_name,omitempty"`
	PersistentProfilePath     *string                `json:"persistent_profile_path,omitempty"`
	FilterPolicyID            *string                `json:"filter_policy_id,omitempty"`
	FilterPolicyName          *string                `json:"filter_policy_name,omitempty"`
	FilterPolicyForceDisabled bool                   `json:"filter_policy_force_disabled"`
	SessionTimeLimit          *string                `json:"session_time_limit,omitempty"`
	Categories                []string               `json:"categories,omitempty"`
	DefaultCategory           string                 `json:"default_category,omitempty"`
	AllowNetworkSelection     bool                   `json:"allow_network_selection"`
	RequireGPU                bool                   `json:"require_gpu"`
	GPUCount                  float64                `json:"gpu_count"`
	Hidden                    bool                   `json:"hidden"`
	Notes                     *string                `json:"notes,omitempty"`
	ImageType                 string                 `json:"image_type"`
	ServerPoolID              *string                `json:"server_pool_id,omitempty"`
	LinkURL                   *string                `json:"link_url,omitempty"`
	CPUAllocationMethod       string                 `json:"cpu_allocation_method"`
	UncompressedSizeMB        int                    `json:"uncompressed_size_mb,omitempty"`
	LaunchConfig              map[string]interface{} `json:"launch_config,omitempty"`
	RDPClientType             *string                `json:"rdp_client_type,omitempty"`
	OverrideEgressGateways    bool                   `json:"override_egress_gateways"`
	RemoteAppName             *string                `json:"remote_app_name,omitempty"`
	RemoteAppArgs             *string                `json:"remote_app_args,omitempty"`
	RemoteAppIcon             *string                `json:"remote_app_icon,omitempty"`
	RemoteAppProgram          *string                `json:"remote_app_program,omitempty"`
	IsRemoteApp               bool                   `json:"is_remote_app"`
}

// DockerRunConfig represents the Docker Run Config Override structure
type DockerRunConfig struct {
	Environment map[string]string `json:"environment,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	User        string            `json:"user,omitempty"`
	Devices     []string          `json:"devices,omitempty"`
	SecurityOpt []string          `json:"security_opt,omitempty"`
	ShmSize     string            `json:"shm_size,omitempty"`
	Privileged  bool              `json:"privileged,omitempty"`
	CapAdd      []string          `json:"cap_add,omitempty"`
	CapDrop     []string          `json:"cap_drop,omitempty"`
	DNS         []string          `json:"dns,omitempty"`
	ExtraHosts  map[string]string `json:"extra_hosts,omitempty"`
}

// Response represents the full response structure containing the image details.
type Response struct {
	Image ImageDetail `json:"image"`
}

// DeleteImageRequest represents the payload required to delete an image.
// Now includes APIKey and APIKeySecret.
type DeleteImageRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
	TargetImage  struct {
		ImageID string `json:"image_id"`
	} `json:"target_image"`
}

// CreateImage sends a POST request to /api/public/create_image with the given CreateImageRequest payload.
// On success, it returns the parsed Response object.
func (api *KasmAPI) CreateImage(ctx context.Context, req CreateImageRequest) (*Response, error) {
	endpoint := "/api/public/create_image"

	// Populate API credentials
	req.APIKey = api.APIKey
	req.APIKeySecret = api.APIKeySecret

	respBody, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create image at %s: %w", endpoint, err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().
			Err(err).
			Str("endpoint", endpoint).
			RawJSON("response_body", respBody).
			Msg("Failed to unmarshal create_image response")
		return nil, fmt.Errorf("failed to unmarshal create_image response: %w", err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", response.Image.ImageID).
		Msg("Image created successfully")

	return &response, nil
}

// UpdateImage sends a POST request to /api/public/update_image to update an existing image.
// req.TargetImage.ImageID must be set. On success, it returns the parsed Response object.
func (api *KasmAPI) UpdateImage(ctx context.Context, req CreateImageRequest) (*Response, error) {
	endpoint := "/api/public/update_image"

	// Populate API credentials
	req.APIKey = api.APIKey
	req.APIKeySecret = api.APIKeySecret

	if req.TargetImage.ImageID == "" {
		return nil, fmt.Errorf("image_id must be set in TargetImage before calling UpdateImage")
	}

	respBody, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update image at %s: %w", endpoint, err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().
			Err(err).
			Str("endpoint", endpoint).
			Str("image_id", req.TargetImage.ImageID).
			RawJSON("response_body", respBody).
			Msg("Failed to unmarshal update_image response")
		return nil, fmt.Errorf("failed to unmarshal update_image response: %w", err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", response.Image.ImageID).
		Msg("Image updated successfully")

	return &response, nil
}

// DeleteImage sends a POST request to /api/public/delete_image to remove an existing image.
// imageID must be provided and must be a valid image ID.
func (api *KasmAPI) DeleteImage(ctx context.Context, imageID string) error {
	endpoint := "/api/public/delete_image"

	if imageID == "" {
		return fmt.Errorf("image_id must be provided")
	}

	reqPayload := DeleteImageRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}
	reqPayload.TargetImage.ImageID = imageID

	_, err := api.MakePostRequest(ctx, endpoint, reqPayload)
	if err != nil {
		return fmt.Errorf("failed to delete image (id=%s) at %s: %w", imageID, endpoint, err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", imageID).
		Msg("Image deleted successfully")

	return nil
}
