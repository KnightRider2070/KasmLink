package api

import (
	"encoding/json"
	"fmt"
	"log"
)

func (api *KasmAPI) CreateUser(user TargetUser) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/create_user", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user":    user,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		return nil, err
	}

	var createdUser UserResponse
	if err := json.Unmarshal(response, &createdUser); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &createdUser, nil
}

func (api *KasmAPI) GetUser(userID, username string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_user", api.BaseURL)
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
		return nil, err
	}

	var user UserResponse
	if err := json.Unmarshal(response, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &user, nil
}

func (api *KasmAPI) GetUsers() ([]UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/get_users", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
	}

	// Make the POST request
	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		return nil, err
	}

	// Struct to hold parsed users
	var parsedResponse struct {
		Users []UserResponse `json:"users"`
	}

	// Parse the response
	if err := json.Unmarshal(response, &parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return parsedResponse.Users, nil
}

func (api *KasmAPI) UpdateUser(user TargetUser) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/public/update_user", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user":    user,
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		return nil, err
	}

	var updatedUser UserResponse
	if err := json.Unmarshal(response, &updatedUser); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &updatedUser, nil
}

func (api *KasmAPI) DeleteUser(userID string, force bool) error {
	url := fmt.Sprintf("%s/api/public/delete_user", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
		"force": force,
	}

	_, err := api.MakePostRequest(url, payload)
	return err
}

func (api *KasmAPI) LogoutUser(userID string) error {
	url := fmt.Sprintf("%s/api/public/logout_user", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	_, err := api.MakePostRequest(url, payload)
	return err
}

func (api *KasmAPI) GetUserAttributes(userID string) (*UserAttributes, error) {
	url := fmt.Sprintf("%s/api/public/get_attributes", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":        api.APIKey,
		"api_key_secret": api.APIKeySecret,
		"target_user": map[string]string{
			"user_id": userID,
		},
	}

	response, err := api.MakePostRequest(url, payload)
	if err != nil {
		return nil, err
	}

	var attributes UserAttributes
	if err := json.Unmarshal(response, &attributes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &attributes, nil
}

func (api *KasmAPI) UpdateUserAttributes(attributes UserAttributes) error {
	url := fmt.Sprintf("%s/api/public/update_user_attributes", api.BaseURL)
	payload := map[string]interface{}{
		"api_key":                api.APIKey,
		"api_key_secret":         api.APIKeySecret,
		"target_user_attributes": attributes,
	}

	_, err := api.MakePostRequest(url, payload)
	return err
}

func (api *KasmAPI) AddUserToGroup(userID string, groupID string) error {
	url := fmt.Sprintf("%s/api/public/add_user_group", api.BaseURL)
	log.Printf("Adding user %s to group %s", userID, groupID)

	// Create the request payload
	request := GroupUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}
	request.TargetUser.UserID = userID
	request.TargetGroup.GroupID = groupID

	// Make the API request
	_, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Printf("Error adding user to group: %v", err)
		return err
	}

	log.Printf("User %s successfully added to group %s", userID, groupID)
	return nil
}

func (api *KasmAPI) RemoveUserFromGroup(userID string, groupID string) error {
	url := fmt.Sprintf("%s/api/public/remove_user_group", api.BaseURL)
	log.Printf("Removing user %s from group %s", userID, groupID)

	// Create the request payload
	request := GroupUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}
	request.TargetUser.UserID = userID
	request.TargetGroup.GroupID = groupID

	// Make the API request
	_, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Printf("Error removing user from group: %v", err)
		return err
	}

	log.Printf("User %s successfully removed from group %s", userID, groupID)
	return nil
}

func (api *KasmAPI) GenerateLoginLink(userID string) (string, error) {
	url := fmt.Sprintf("%s/api/public/get_login", api.BaseURL)
	log.Printf("Generating login link for user %s", userID)

	// Create the request payload
	request := LoginRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}
	request.TargetUser.UserID = userID

	// Make the API request
	response, err := api.MakePostRequest(url, request)
	if err != nil {
		log.Printf("Error generating login link: %v", err)
		return "", err
	}

	// Parse the response
	var loginResponse LoginResponse
	if err := json.Unmarshal(response, &loginResponse); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Login link generated for user %s: %s", userID, loginResponse.URL)
	return loginResponse.URL, nil
}
