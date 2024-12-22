package procedures

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
)

// CreateKasmWorkspace creates a workspace based on user-provided YAML file
func CreateKasmWorkspace(ctx context.Context, imageDetail webApi.ImageDetail, details userParser.UserDetails, kasmApi *webApi.KasmAPI) error {
	// Parse volume mounts
	volumeMappings, err := parseVolumeMounts(details)
	if err != nil {
		return fmt.Errorf("failed to parse volume mounts: %w", err)
	}

	// Serialize run configuration to JSON string
	runConfig := webApi.DockerRunConfig{
		Environment: details.EnvironmentArgs,
		Network:     details.Network,
	}

	runConfigJSON, err := json.Marshal(runConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal run configuration: %w", err)
	}

	volumeMappingsJSON, err := json.Marshal(volumeMappings)
	if err != nil {
		return fmt.Errorf("failed to marshal volume mappings: %w", err)
	}

	targetImage := webApi.TargetImage{
		Name:                  imageDetail.Name,
		Cores:                 imageDetail.Cores,
		Memory:                imageDetail.Memory * 1000000,
		FriendlyName:          imageDetail.FriendlyName,
		Description:           imageDetail.Description,
		RestrictNetworkNames:  []string{details.Network},  // Restrict to specified network
		VolumeMappings:        string(volumeMappingsJSON), // Pass as serialized JSON
		RunConfig:             string(runConfigJSON),      // Serialized run configuration
		AllowNetworkSelection: false,                      // Allows network selection
	}

	// Create the request payload
	req := webApi.CreateImageRequest{
		APIKey:       kasmApi.APIKey,
		APIKeySecret: kasmApi.APIKeySecret,
		TargetImage:  targetImage,
	}

	// Call the API to create the image
	response, err := kasmApi.CreateImage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	log.Info().
		Str("image_id", response.Image.ImageID).
		Msg("Workspace created successfully")
	return nil
}
