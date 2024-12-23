package test

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
user_details:
  - target_user:
      username: "testuser"
      user_id: "12345"
    role: "admin"
    assigned_container_tag: "tag1"
    assigned_container_id: "cont123"
    kasm_session_of_container: "sess567"
    network: "default"
    volume-mounts:
      "/data": "/mnt/data"
    environment_args:
      "key": "value"
`
	tempFilePath := createTempConfigFile(t, configContent)
	defer removeTempFile(tempFilePath)

	parser := userParser.NewUserParser()
	config, err := parser.LoadConfig(tempFilePath)
	require.NoError(t, err)

	t.Logf("Loaded Config: %+v", config)

	assert.Len(t, config.UserDetails, 1)
	assert.Equal(t, "testuser", config.UserDetails[0].TargetUser.Username)
	assert.Equal(t, "12345", config.UserDetails[0].TargetUser.UserID)
}

func TestSaveConfig(t *testing.T) {
	tempFilePath := createTempConfigFile(t, "")
	defer removeTempFile(tempFilePath)

	config := &userParser.UsersConfig{
		UserDetails: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "testuser",
					UserID:   "12345",
				},
				Role:                   "admin",
				AssignedContainerTag:   "tag1",
				AssignedContainerId:    "cont123",
				KasmSessionOfContainer: "sess567",
				Network:                "default",
				VolumeMounts:           map[string]string{"/data": "/mnt/data"},
				EnvironmentArgs:        map[string]string{"key": "value"},
			},
		},
	}

	parser := userParser.NewUserParser()
	err := parser.SaveConfig(tempFilePath, config)
	require.NoError(t, err)

	loadedConfig, err := parser.LoadConfig(tempFilePath)
	require.NoError(t, err)
	assert.Equal(t, config, loadedConfig)
}

func TestUpdateUserConfig(t *testing.T) {
	configContent := `
user_details:
  - target_user:
      username: "testuser"
      user_id: "12345"
    role: "admin"
    assigned_container_tag: "tag1"
    assigned_container_id: "cont123"
    kasm_session_of_container: "sess567"
    network: "default"
    volume-mounts:
      "/data": "/mnt/data"
    environment_args:
      "key": "value"
`
	tempFilePath := createTempConfigFile(t, configContent)
	defer removeTempFile(tempFilePath)

	parser := userParser.NewUserParser()
	newUserID := "67890"
	newKasmSessionID := "sess789"
	newContainerID := "cont456"

	err := parser.UpdateUserConfig(tempFilePath, "testuser", newUserID, newKasmSessionID, newContainerID)
	require.NoError(t, err)

	updatedConfig, err := parser.LoadConfig(tempFilePath)
	require.NoError(t, err)

	updatedUser := updatedConfig.UserDetails[0]
	assert.Equal(t, newUserID, updatedUser.TargetUser.UserID)
	assert.Equal(t, newKasmSessionID, updatedUser.KasmSessionOfContainer)
	assert.Equal(t, newContainerID, updatedUser.AssignedContainerId)
}

func TestValidateConfig(t *testing.T) {
	validConfig := &userParser.UsersConfig{
		UserDetails: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "testuser",
					UserID:   "12345",
				},
			},
		},
	}
	invalidConfig := &userParser.UsersConfig{
		UserDetails: []userParser.UserDetails{
			{
				TargetUser: models.TargetUser{
					Username: "",
					UserID:   "",
				},
			},
		},
	}

	parser := userParser.NewUserParser()

	err := parser.ValidateConfig(validConfig)
	require.NoError(t, err)

	err = parser.ValidateConfig(invalidConfig)
	require.Error(t, err)
}
