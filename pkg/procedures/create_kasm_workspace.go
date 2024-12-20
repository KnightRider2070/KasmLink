package procedures

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/userParser"
	"kasmlink/pkg/webApi"
)

// createKasmWorkspace creates a workspace based on user-provided YAML file
func createKasmWorkspace(ctx context.Context, imageDetail webApi.ImageDetail, details userParser.UserDetails, kasmApi *webApi.KasmAPI) error {
	// Serialize volume mappings to JSON string
	volumeMappings, err := json.Marshal(details.VolumeMounts)
	if err != nil {
		return fmt.Errorf("failed to marshal volume mappings: %w", err)
	}

	conf, err := parseVolumeMounts(details)

	if err != nil {
		return fmt.Errorf("failed to parse volume mounts: %w", err)
	}

	// Serialize run configuration to JSON string
	runConfig := webApi.DockerRunConfig{
		Environment: details.EnvironmentArgs,
		Volumes:     conf,
		Network:     details.Network,
	}

	runConfigJSON, err := json.Marshal(runConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal run configuration: %w", err)
	}

	// Build the target image configuration
	targetImage := webApi.TargetImage{
		Name:                  imageDetail.Name,
		Cores:                 imageDetail.Cores,
		Memory:                imageDetail.Memory,
		FriendlyName:          imageDetail.FriendlyName,
		RestrictNetworkNames:  []string{details.Network}, // Restrict to specified network
		VolumeMappings:        string(volumeMappings),    // Serialized volume mappings
		RunConfig:             string(runConfigJSON),     // Serialized run configuration
		AllowNetworkSelection: false,                     // Allows network selection
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
