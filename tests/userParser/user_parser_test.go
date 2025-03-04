package userParser_tests

import (
	"kasmlink/pkg/userParser"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kasmlink/pkg/api/models"
)

func createTempConfigFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "user_config_*.yaml")
	require.NoError(t, err)
	defer tempFile.Close()

	_, err = tempFile.WriteString(content)
	require.NoError(t, err)

	return tempFile.Name()
}

func removeTempFile(path string) {
	_ = os.Remove(path)
}

func TestLoadConfig(t *testing.T) {
	configContent := `
workspaces:
  - workspace_id: "workspace1"
    image_config:
      name: "image1"
users:
  - target_user:
      username: "testuser"
      user_id: "12345"
    workspace_id: "workspace1"
    environment:
      key: "value"
    volume_mounts:
      "/data": "/mnt/data"
`
	tempFilePath := createTempConfigFile(t, configContent)
	defer removeTempFile(tempFilePath)

	parser := userParser.NewUserParser()
	config, err := parser.LoadDeploymentConfig(tempFilePath)
	require.NoError(t, err)

	t.Logf("Loaded Config: %+v", config)

	assert.Len(t, config.Users, 1)
	assert.Equal(t, "testuser", config.Users[0].TargetUser.Username)
	assert.Equal(t, "12345", config.Users[0].TargetUser.UserID)
}

func TestSaveConfig(t *testing.T) {
	tempFilePath := createTempConfigFile(t, "")
	defer removeTempFile(tempFilePath)

	config := &userParser.DeploymentConfig{
		Workspaces: []userParser.WorkspaceConfig{
			{
				WorkspaceID: "workspace1",
				ImageConfig: models.TargetImage{
					FriendlyName:    "image1",
					DockerImageName: "image1",
					// Ensure other fields match the actual default state.
					RestrictNetworkNames: []string{}, // Use empty slice instead of nil.
					Categories:           []string{},
					ExecConfig:           map[string]interface{}{},
					LaunchConfig:         map[string]interface{}{},
				},
			},
		},
		Users: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "testuser",
					UserID:   "12345",
				},
				GroupName:    "testgroup",
				Environment:  map[string]string{"key": "value"},
				VolumeMounts: map[string]string{"/data": "/mnt/data"},
			},
		},
	}

	parser := userParser.NewUserParser()
	err := parser.SaveDeploymentConfig(tempFilePath, config)
	require.NoError(t, err)

	loadedConfig, err := parser.LoadDeploymentConfig(tempFilePath)
	require.NoError(t, err)

	assert.EqualValues(t, config, loadedConfig)

}

func TestUpdateUserConfig(t *testing.T) {
	configContent := `
workspaces:
  - workspace_id: "workspace1"
    image_config:
      name: "image1"
users:
  - target_user:
      username: "testuser"
      user_id: "12345"
    workspace_id: "workspace1"
    environment:
      key: "value"
    volume_mounts:
      "/data": "/mnt/data"
`
	tempFilePath := createTempConfigFile(t, configContent)
	defer removeTempFile(tempFilePath)

	parser := userParser.NewUserParser()
	newUserID := "67890"
	newKasmSessionID := "sess789"

	err := parser.UpdateUserDetails(tempFilePath, "testuser", newUserID, newKasmSessionID)
	require.NoError(t, err)

	updatedConfig, err := parser.LoadDeploymentConfig(tempFilePath)
	require.NoError(t, err)

	updatedUser := updatedConfig.Users[0]
	assert.Equal(t, newUserID, updatedUser.TargetUser.UserID)
	assert.Equal(t, newKasmSessionID, updatedUser.KasmSessionID)
}

func TestValidateConfig(t *testing.T) {
	validConfig := &userParser.DeploymentConfig{
		Workspaces: []userParser.WorkspaceConfig{
			{
				WorkspaceID: "workspace1",
				ImageConfig: models.TargetImage{DockerImageName: "image1"},
			},
		},
		Users: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "testuser",
					UserID:   "12345",
				},
				GroupName: "testgroup",
			},
		},
	}
	invalidConfig := &userParser.DeploymentConfig{
		Workspaces: []userParser.WorkspaceConfig{
			{
				WorkspaceID: "workspace1",
				ImageConfig: models.TargetImage{DockerImageName: "image1"},
			},
		},
		Users: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "",
					UserID:   "",
				},
				GroupName: "",
			},
		},
	}

	parser := userParser.NewUserParser()

	err := parser.ValidateDeploymentConfig(validConfig)
	require.NoError(t, err)

	err = parser.ValidateDeploymentConfig(invalidConfig)
	require.Error(t, err)
}
