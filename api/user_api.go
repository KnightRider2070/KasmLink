package api

import (
	"encoding/json"
	"fmt"
	"log"
)

// CreateUser creates a new user in the Kasm system.
func (api *KasmAPI) CreateUser(user CreateUserRequest) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/create_user", api.BaseURL)
	log.Printf("Creating user with username: %s", user.TargetUser.Username)
	response, err := api.MakePostRequest(url, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	var createdUser UserResponse
	if err := json.Unmarshal(response, &createdUser); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully created user with ID: %s", createdUser.UserID)
	return &createdUser, nil
}

// GetUser retrieves a user's details from the Kasm system.
func (api *KasmAPI) GetUser(userID string, username string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_user", api.BaseURL)
	log.Printf("Retrieving user with ID: %s or username: %s", userID, username)
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
		log.Printf("Error retrieving user: %v", err)
		return nil, err
	}

	var user UserResponse
	if err := json.Unmarshal(response, &user); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully retrieved user with ID: %s", user.UserID)
	return &user, nil
}

// GetUsers retrieves the list of users registered in the system.
func (api *KasmAPI) GetUsers() ([]UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_users", api.BaseURL)
	log.Printf("Retrieving list of users from URL: %s", url)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error retrieving users: %v", err)
		return nil, err
	}

	var users []UserResponse
	if err := json.Unmarshal(response, &users); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully retrieved %d users", len(users))
	return users, nil
}

// UpdateUser updates a user's details in the Kasm system.
func (api *KasmAPI) UpdateUser(user UpdateUserRequest) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/update_user", api.BaseURL)
	log.Printf("Updating user with ID: %s", user.TargetUser.UserID)
	response, err := api.MakePostRequest(url, user)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, err
	}

	var updatedUser UserResponse
	if err := json.Unmarshal(response, &updatedUser); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully updated user with ID: %s", updatedUser.UserID)
	return &updatedUser, nil
}

// DeleteUser deletes a user from the Kasm system.
func (api *KasmAPI) DeleteUser(userID string, force bool) error {
	url := fmt.Sprintf("%s/api/public/delete_user", api.BaseURL)
	log.Printf("Deleting user with ID: %s (force: %v)", userID, force)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
		"force": force,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return err
	}

	if len(response) > 0 {
		log.Printf("User %s successfully deleted", userID)
	}
	return nil
}

// LogoutUser logs out all sessions for an existing user.
func (api *KasmAPI) LogoutUser(userID string) error {
	url := fmt.Sprintf("%s/api/public/logout_user", api.BaseURL)
	log.Printf("Logging out user with ID: %s", userID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error logging out user: %v", err)
		return err
	}

	if len(response) > 0 {
		log.Printf("User %s successfully logged out", userID)
	}
	return nil
}

// GetUserAttributes retrieves the attribute (preferences) settings for an existing user.
func (api *KasmAPI) GetUserAttributes(userID string) (*UserAttributesResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_attributes", api.BaseURL)
	log.Printf("Retrieving attributes for user with ID: %s", userID)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		log.Printf("Error retrieving user attributes: %v", err)
		return nil, err
	}

	var userAttributes UserAttributesResponse
	if err := json.Unmarshal(response, &userAttributes); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Successfully retrieved attributes for user with ID: %s", userID)
	return &userAttributes, nil
}

// UpdateUserAttributes updates a user's attributes in the Kasm system.
func (api *KasmAPI) UpdateUserAttributes(attributes UpdateUserAttributesRequest) error {
	url := fmt.Sprintf("%s/api/public/update_user_attributes", api.BaseURL)
	log.Printf("Updating attributes for user with ID: %s", attributes.UserID)
	response, err := api.MakePostRequest(url, attributes)
	if err != nil {
		log.Printf("Error updating user attributes: %v", err)
		return err
	}

	if len(response) > 0 {
		log.Printf("User attributes successfully updated for user %s", attributes.UserID)
	}
	return nil
}
