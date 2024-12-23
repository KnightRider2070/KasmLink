package SystemTests

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/webApi"
	"testing"
	"time"
)

func TestCreateImage(t *testing.T) {
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()

	// Prepare the RunConfig struct
	runConfig := webApi.DockerRunConfig{
		Environment: map[string]string{
			"LC_ALL": "fr_FR.UTF-8",
			"TZ":     "Europe/Paris",
		},
		CapAdd:      []string{"SYS_ADMIN", "MKNOD"},
		CapDrop:     []string{"SYS_RESOURCE"},
		ShmSize:     "4g",
		Privileged:  true,
		Hostname:    "HOST-123",
		Devices:     []string{"/dev/input/event0:/dev/input/event0:rwm"},
		SecurityOpt: []string{"seccomp=unconfined"},
	}

	// Convert RunConfig to JSON string
	runConfigBytes, err := json.Marshal(runConfig)

	// Prepare the request to create a new image, including API credentials
	createReq := webApi.CreateImageRequest{
		APIKey:       kApi.APIKey,
		APIKeySecret: kApi.APIKeySecret,
		TargetImage: webApi.TargetImage{
			Cores:               2,
			CPUAllocationMethod: "Inherit",
			Description:         "Test image creation",
			Enabled:             true,
			FriendlyName:        "test_integration",
			GPUCount:            0,
			ImageType:           "Container",
			Memory:              2786000000,
			Name:                "kasmweb/chrome",
			RunConfig:           string(runConfigBytes), // Pass as JSON string
		},
	}

	// Call CreateImage
	resp, err := kApi.CreateImage(ctx, createReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Image.ImageID)
	assert.Equal(t, "test_integration", resp.Image.FriendlyName)
	t.Logf("Created image with ID: %s", resp.Image.ImageID)
}

func TestUpdateImage(t *testing.T) {
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()

	// Prepare the RunConfig struct
	runConfig := webApi.DockerRunConfig{
		Environment: map[string]string{
			"LC_ALL": "fr_FR.UTF-8",
			"TZ":     "Europe/Paris",
		},
		CapAdd:      []string{"SYS_ADMIN", "MKNOD"},
		CapDrop:     []string{"SYS_RESOURCE"},
		ShmSize:     "4g",
		Privileged:  true,
		Hostname:    "HOST-123",
		Devices:     []string{"/dev/input/event0:/dev/input/event0:rwm"},
		SecurityOpt: []string{"seccomp=unconfined"},
	}

	// Convert RunConfig to JSON string
	runConfigBytes, err := json.Marshal(runConfig)

	createReq := webApi.CreateImageRequest{
		APIKey:       kApi.APIKey,
		APIKeySecret: kApi.APIKeySecret,
		TargetImage: webApi.TargetImage{
			Cores:               2,
			CPUAllocationMethod: "Inherit",
			Description:         "Test image for update",
			Enabled:             true,
			FriendlyName:        "test_update_before",
			GPUCount:            0,
			ImageType:           "Container",
			Memory:              2786000000,
			Name:                "kasmweb/chrome",
			RunConfig:           string(runConfigBytes),
		},
	}

	createdResp, err := kApi.CreateImage(ctx, createReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdResp.Image.ImageID)

	// Update the friendly name of the created image
	updateReq := createReq
	updateReq.TargetImage.ImageID = createdResp.Image.ImageID
	updateReq.TargetImage.FriendlyName = "test_update_after"

	updatedResp, err := kApi.UpdateImage(ctx, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "test_update_after", updatedResp.Image.FriendlyName)
	t.Logf("Updated image with ID: %s", updatedResp.Image.ImageID)
}

func TestDeleteImage(t *testing.T) {
	kApi := webApi.NewKasmAPI(baseUrl, apiSecret, apiKeySecret, true, 100*time.Second)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()

	// Prepare the RunConfig struct
	runConfig := webApi.DockerRunConfig{
		Environment: map[string]string{
			"LC_ALL": "fr_FR.UTF-8",
			"TZ":     "Europe/Paris",
		},
		CapAdd:      []string{"SYS_ADMIN", "MKNOD"},
		CapDrop:     []string{"SYS_RESOURCE"},
		ShmSize:     "4g",
		Privileged:  true,
		Hostname:    "HOST-123",
		Devices:     []string{"/dev/input/event0:/dev/input/event0:rwm"},
		SecurityOpt: []string{"seccomp=unconfined"},
	}

	// Convert RunConfig to JSON string
	runConfigBytes, err := json.Marshal(runConfig)

	createReq := webApi.CreateImageRequest{
		APIKey:       kApi.APIKey,
		APIKeySecret: kApi.APIKeySecret,
		TargetImage: webApi.TargetImage{
			Cores:               2,
			CPUAllocationMethod: "Inherit",
			Description:         "Test image for deletion",
			Enabled:             true,
			FriendlyName:        "test_delete",
			GPUCount:            0,
			ImageType:           "Container",
			Memory:              2786000000,
			Name:                "kasmweb/chrome",
			RunConfig:           string(runConfigBytes),
		},
	}

	createdResp, err := kApi.CreateImage(ctx, createReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdResp.Image.ImageID)

	// Now delete the image
	err = kApi.DeleteImage(ctx, createdResp.Image.ImageID)
	assert.NoError(t, err)
	t.Logf("Deleted image with ID: %s", createdResp.Image.ImageID)
}
