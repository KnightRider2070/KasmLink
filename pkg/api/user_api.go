package api

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// RequestKasmSession requests a new Kasm session.
func (api *KasmAPI) CreateUser(user TargetUser) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/create_user", api.BaseURL)
	log.Info().Str("url", url).Str("user", user.Username).Msg("Creating new user")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user":    user,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user", user.Username).Msg("Failed to create user")
		return nil, err
	}

	var createdUser UserResponse
	if err := json.Unmarshal(response, &createdUser); err != nil {
		log.Error().Err(err).Msg("Failed to decode create user response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("user_id", createdUser.UserID).Msg("User created successfully")
	return &createdUser, nil
}

// GetUser retrieves user details by userID or username.
func (api *KasmAPI) GetUser(userID, username string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_user", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Str("username", username).Msg("Fetching user details")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id":  userID,
			"username": username,
		},
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch user details")
		return nil, err
	}

	var user UserResponse
	if err := json.Unmarshal(response, &user); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to decode get user response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("user_id", user.UserID).Msg("User details retrieved successfully")
	return &user, nil
}

// GetUsers retrieves a list of all users.
func (api *KasmAPI) GetUsers() ([]UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_users", api.BaseURL)
	log.Info().Str("url", url).Msg("Fetching all users")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch users")
		return nil, err
	}

	var parsedResponse struct {
		Users []UserResponse `json:"users"`
	}

	if err := json.Unmarshal(response, &parsedResponse); err != nil {
		log.Error().Err(err).Msg("Failed to decode get users response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Int("user_count", len(parsedResponse.Users)).Msg("Users retrieved successfully")
	return parsedResponse.Users, nil
}

// UpdateUser updates an existing user's details.
func (api *KasmAPI) UpdateUser(user TargetUser) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/update_user", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", user.UserID).Msg("Updating user details")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user":    user,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("Failed to update user")
		return nil, err
	}

	var updatedUser UserResponse
	if err := json.Unmarshal(response, &updatedUser); err != nil {
		log.Error().Err(err).Msg("Failed to decode update user response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("user_id", updatedUser.UserID).Msg("User updated successfully")
	return &updatedUser, nil
}

// DeleteUser removes a user by userID with optional force.
func (api *KasmAPI) DeleteUser(userID string, force bool) error {
	url := fmt.Sprintf("%s/api/public/delete_user", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Bool("force", force).Msg("Deleting user")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
		"force": force,
	}

	_, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		return err
	}

	log.Info().Str("user_id", userID).Msg("User deleted successfully")
	return nil
}

// LogoutUser logs a user out by userID.
func (api *KasmAPI) LogoutUser(userID string) error {
	url := fmt.Sprintf("%s/api/public/logout_user", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Msg("Logging out user")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	_, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to logout user")
		return err
	}

	log.Info().Str("user_id", userID).Msg("User logged out successfully")
	return nil
}

// GetUserAttributes retrieves the attributes of a user.
func (api *KasmAPI) GetUserAttributes(userID string) (*UserAttributes, error) {
	url := fmt.Sprintf("%s/api/public/get_attributes", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Msg("Fetching user attributes")

	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch user attributes")
		return nil, err
	}

	var attributes UserAttributes
	if err := json.Unmarshal(response, &attributes); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to decode user attributes response")
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Info().Str("user_id", userID).Msg("User attributes retrieved successfully")
	return &attributes, nil
}

// UpdateUserAttributes updates a user's attributes.
func (api *KasmAPI) UpdateUserAttributes(attributes UserAttributes) error {
	url := fmt.Sprintf("%s/api/public/update_user_attributes", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", attributes.UserID).Msg("Updating user attributes")

	payload := map[string]interface{}{
		"api_key":                api.APIKey,
		"api_key_secret":         api.APIKeySecret,
		"target_user_attributes": attributes,
	}

	_, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Error().Err(err).Str("user_id", attributes.UserID).Msg("Failed to update user attributes")
		return err
	}

	log.Info().Str("user_id", attributes.UserID).Msg("User attributes updated successfully")
	return nil
}

// AddUserToGroup adds a user to a specific group.
func (api *KasmAPI) AddUserToGroup(userID, groupID string) error {
	url := fmt.Sprintf("%s/api/public/add_user_group", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Str("group_id", groupID).Msg("Adding user to group")

	// Inline struct fields in the request
	request := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
		"target_group": map[string]string{
			"group_id": groupID,
		},
	}

	_, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("group_id", groupID).Msg("Failed to add user to group")
		return err
	}

	log.Info().Str("user_id", userID).Str("group_id", groupID).Msg("User added to group successfully")
	return nil
}

// RemoveUserFromGroup removes a user from a specific group.
func (api *KasmAPI) RemoveUserFromGroup(userID, groupID string) error {
	url := fmt.Sprintf("%s/api/public/remove_user_group", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Str("group_id", groupID).Msg("Removing user from group")

	// Inline struct fields in the request
	request := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
		"target_group": map[string]string{
			"group_id": groupID,
		},
	}

	_, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("group_id", groupID).Msg("Failed to remove user from group")
		return err
	}

	log.Info().Str("user_id", userID).Str("group_id", groupID).Msg("User removed from group successfully")
	return nil
}

// GenerateLoginLink generates a login link for a user.
func (api *KasmAPI) GenerateLoginLink(userID string) (string, error) {
	url := fmt.Sprintf("%s/api/public/get_login", api.BaseURL)
	log.Info().Str("url", url).Str("user_id", userID).Msg("Generating login link")

	request := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	response, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to generate login link")
		return "", err
	}

	var loginResponse LoginResponse
	if err := json.Unmarshal(response, &loginResponse); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to decode login link response")
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().Str("user_id", userID).Str("login_url", loginResponse.URL).Msg("Login link generated successfully")
	return loginResponse.URL, nil
}
