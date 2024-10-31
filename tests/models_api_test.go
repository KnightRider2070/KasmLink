package tests

import (
	"encoding/json"
	"kasmlink/pkg/api"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test TargetUser Struct
func TestTargetUserStruct(t *testing.T) {
	user := api.TargetUser{
		UserID:       "1234",
		Username:     "test_user",
		FirstName:    "John",
		LastName:     "Doe",
		Locked:       false,
		Disabled:     false,
		Organization: "TestOrg",
		Phone:        "123-456-7890",
		Password:     "password123",
	}

	assert.Equal(t, "1234", user.UserID)
	assert.Equal(t, "test_user", user.Username)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.False(t, user.Locked)
	assert.False(t, user.Disabled)
	assert.Equal(t, "TestOrg", user.Organization)
	assert.Equal(t, "123-456-7890", user.Phone)
	assert.Equal(t, "password123", user.Password)
}

// Test UserResponse Struct
func TestUserResponseStruct(t *testing.T) {
	userResponseJSON := `{
		"user_id": "1234",
		"username": "test_user",
		"first_name": "John",
		"last_name": "Doe",
		"phone": "123-456-7890",
		"organization": "TestOrg",
		"realm": "local",
		"groups": [{"name": "group1", "group_id": "group123"}],
		"kasms": [{
			"kasm_id": "kasm123",
			"start_date": "2024-01-01T00:00:00Z",
			"keepalive_date": "2024-01-01T00:10:00Z",
			"expiration_date": "2024-01-01T01:00:00Z",
			"server": {
				"server_id": "server123",
				"hostname": "localhost",
				"port": 22
			}
		}],
		"disabled": false,
		"locked": false,
		"created": "2024-01-01T12:00:00Z"
	}`

	var userResponse api.UserResponse
	err := json.Unmarshal([]byte(userResponseJSON), &userResponse)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, "1234", userResponse.UserID)
	assert.Equal(t, "test_user", userResponse.Username)
	assert.Equal(t, "John", *userResponse.FirstName)
	assert.Equal(t, "Doe", *userResponse.LastName)
	assert.Equal(t, "123-456-7890", *userResponse.Phone)
	assert.Equal(t, "TestOrg", *userResponse.Organization)
	assert.Equal(t, "local", userResponse.Realm)
	assert.False(t, userResponse.Disabled)
	assert.False(t, userResponse.Locked)

	// Check groups
	assert.Len(t, userResponse.Groups, 1)
	assert.Equal(t, "group1", userResponse.Groups[0].Name)

	// Check Kasms
	assert.Len(t, userResponse.Kasms, 1)
	assert.Equal(t, "kasm123", userResponse.Kasms[0].KasmID)
	assert.Equal(t, "localhost", userResponse.Kasms[0].Server.Hostname)
	assert.Equal(t, 22, userResponse.Kasms[0].Server.Port)
}

// Test RequestKasmRequest Struct
func TestRequestKasmRequestStruct(t *testing.T) {
	request := api.RequestKasmRequest{
		APIKey:        "key123",
		APIKeySecret:  "secret123",
		UserID:        "user123",
		ImageID:       "image123",
		EnableSharing: true,
	}

	assert.Equal(t, "key123", request.APIKey)
	assert.Equal(t, "secret123", request.APIKeySecret)
	assert.Equal(t, "user123", request.UserID)
	assert.Equal(t, "image123", request.ImageID)
	assert.True(t, request.EnableSharing)
}

// Test DestroyKasmRequest Struct
func TestDestroyKasmRequestStruct(t *testing.T) {
	request := api.DestroyKasmRequest{
		APIKey:       "key123",
		APIKeySecret: "secret123",
		KasmID:       "kasm123",
		UserID:       "user123",
	}

	assert.Equal(t, "key123", request.APIKey)
	assert.Equal(t, "secret123", request.APIKeySecret)
	assert.Equal(t, "kasm123", request.KasmID)
	assert.Equal(t, "user123", request.UserID)
}

// Test UserAttributes Struct
func TestUserAttributesStruct(t *testing.T) {
	attrs := api.UserAttributes{
		SSHPublicKey:       "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA7s",
		ShowTips:           true,
		UserID:             "1234",
		ToggleControlPanel: true,
		ChatSFX:            false,
	}

	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA7s", attrs.SSHPublicKey)
	assert.True(t, attrs.ShowTips)
	assert.Equal(t, "1234", attrs.UserID)
	assert.True(t, attrs.ToggleControlPanel)
	assert.False(t, attrs.ChatSFX)
}
