package webApi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// CreateUserRequest represents the payload for creating a user.
type CreateUserRequest struct {
	APIKey       string     `json:"api_key"`
	APIKeySecret string     `json:"api_key_secret"`
	TargetUser   TargetUser `json:"target_user"`
}

// CreateUserResponse represents the response after creating a user.
type CreateUserResponse struct {
	UserResponse
	// Add additional fields if the API returns more data upon user creation.
}

// GetUsersRequest represents the payload for fetching all users.
type GetUsersRequest struct {
	APIKey       string     `json:"api_key"`
	APIKeySecret string     `json:"api_key_secret"`
	TargetUser   TargetUser `json:"target_user"`
}

// GetUserRequest represents the payload for fetching a user.
type GetUserRequest struct {
	APIKey       string     `json:"api_key"`
	APIKeySecret string     `json:"api_key_secret"`
	TargetUser   TargetUser `json:"target_user"`
}

// GetUsersResponse represents the response containing a list of users.
type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
}

// CreateUser creates a new KASM user.
func (api *KasmAPI) CreateUser(ctx context.Context, user TargetUser) (*UserResponse, error) {
	endpoint := "/api/public/create_user"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("username", user.Username).
		Msg("Creating new user")

	// Construct request payload
	requestPayload := CreateUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser:   user,
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("username", user.Username).
			Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Parse the response into UserResponse struct
	var createdUser UserResponse
	if err := json.Unmarshal(responseBytes, &createdUser); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("username", user.Username).
			Msg("Failed to decode create user response")
		return nil, fmt.Errorf("failed to decode create user response: %v", err)
	}

	log.Info().
		Str("user_id", createdUser.UserID).
		Str("username", createdUser.Username).
		Msg("User created successfully")
	return &createdUser, nil
}

// GetUser retrieves user details by userID or username.
func (api *KasmAPI) GetUser(ctx context.Context, userID, username string) (*UserResponse, error) {
	endpoint := "/api/public/get_user"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("username", username).
		Msg("Fetching user details")

	// Construct request payload
	requestPayload := GetUsersRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: TargetUser{
			UserID:   userID,
			Username: username,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to fetch user details")
		return nil, fmt.Errorf("failed to fetch user details: %w", err)
	}

	// Parse the response into UserResponse struct
	var user UserResponse
	if err := json.Unmarshal(responseBytes, &user); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to decode get user response")
		return nil, fmt.Errorf("failed to decode get user response: %v", err)
	}

	log.Info().
		Str("user_id", user.UserID).
		Str("username", user.Username).
		Msg("User details retrieved successfully")
	return &user, nil
}

// GetUsers retrieves a list of all users.
func (api *KasmAPI) GetUsers(ctx context.Context) ([]UserResponse, error) {
	endpoint := "/api/public/get_users"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Msg("Fetching all users")

	// Construct request payload
	requestPayload := GetUsersRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Msg("Failed to fetch users")
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Parse the response into GetUsersResponse struct
	var parsedResponse GetUsersResponse
	if err := json.Unmarshal(responseBytes, &parsedResponse); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Msg("Failed to decode get users response")
		return nil, fmt.Errorf("failed to decode get users response: %v", err)
	}

	log.Info().
		Int("user_count", len(parsedResponse.Users)).
		Str("method", "POST").
		Str("endpoint", endpoint).
		Msg("Users retrieved successfully")
	return parsedResponse.Users, nil
}

// UpdateUserRequest represents the payload for updating a user.
type UpdateUserRequest struct {
	APIKey       string     `json:"api_key"`
	APIKeySecret string     `json:"api_key_secret"`
	TargetUser   TargetUser `json:"target_user"`
}

// UpdateUser updates an existing user's details.
func (api *KasmAPI) UpdateUser(ctx context.Context, user TargetUser) (*UserResponse, error) {
	endpoint := "/api/public/update_user"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", user.UserID).
		Msg("Updating user details")

	// Construct request payload
	requestPayload := UpdateUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser:   user,
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", user.UserID).
			Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Parse the response into UserResponse struct
	var updatedUser UserResponse
	if err := json.Unmarshal(responseBytes, &updatedUser); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", user.UserID).
			Msg("Failed to decode update user response")
		return nil, fmt.Errorf("failed to decode update user response: %v", err)
	}

	log.Info().
		Str("user_id", updatedUser.UserID).
		Str("username", updatedUser.Username).
		Msg("User updated successfully")
	return &updatedUser, nil
}

// DeleteUserRequest represents the payload for deleting a user.
type DeleteUserRequest struct {
	APIKey       string           `json:"api_key"`
	APIKeySecret string           `json:"api_key_secret"`
	TargetUser   DeleteUserTarget `json:"target_user"`
	Force        bool             `json:"force"`
}

// DeleteUserTarget represents the target user details for deletion.
type DeleteUserTarget struct {
	UserID string `json:"user_id"`
}

// DeleteUser removes a user by userID with optional force.
func (api *KasmAPI) DeleteUser(ctx context.Context, userID string, force bool) error {
	endpoint := "/api/public/delete_user"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Bool("force", force).
		Msg("Deleting user")

	// Construct request payload
	requestPayload := DeleteUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: DeleteUserTarget{
			UserID: userID,
		},
		Force: force,
	}

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("User deleted successfully")
	return nil
}

// GetUserAttributesRequest represents the payload for fetching user attributes.
type GetUserAttributesRequest struct {
	APIKey       string                  `json:"api_key"`
	APIKeySecret string                  `json:"api_key_secret"`
	TargetUser   GetUserAttributesTarget `json:"target_user"`
}

// GetUserAttributesTarget represents the target user details for fetching attributes.
type GetUserAttributesTarget struct {
	UserID string `json:"user_id"`
}

// GetUserAttributes retrieves the attributes of a user.
func (api *KasmAPI) GetUserAttributes(ctx context.Context, userID string) (*UserAttributes, error) {
	endpoint := "/api/public/get_attributes"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("Fetching user attributes")

	// Construct request payload
	requestPayload := GetUserAttributesRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: GetUserAttributesTarget{
			UserID: userID,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to fetch user attributes")
		return nil, fmt.Errorf("failed to fetch user attributes: %w", err)
	}

	// Parse the response into UserAttributes struct
	var attributes UserAttributes
	if err := json.Unmarshal(responseBytes, &attributes); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to decode user attributes response")
		return nil, fmt.Errorf("failed to decode user attributes response: %v", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("User attributes retrieved successfully")
	return &attributes, nil
}

// LogoutUserRequest represents the payload for logging out a user.
type LogoutUserRequest struct {
	APIKey       string           `json:"api_key"`
	APIKeySecret string           `json:"api_key_secret"`
	TargetUser   LogoutUserTarget `json:"target_user"`
}

// LogoutUserTarget represents the target user details for logout.
type LogoutUserTarget struct {
	UserID string `json:"user_id"`
}

// LogoutUser logs a user out by userID.
func (api *KasmAPI) LogoutUser(ctx context.Context, userID string) error {
	endpoint := "/api/public/logout_user"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("Logging out user")

	// Construct request payload
	requestPayload := LogoutUserRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: LogoutUserTarget{
			UserID: userID,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to logout user")
		return fmt.Errorf("failed to logout user: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("User logged out successfully")
	return nil
}

// UpdateUserAttributesRequest represents the payload for updating user attributes.
type UpdateUserAttributesRequest struct {
	APIKey               string         `json:"api_key"`
	APIKeySecret         string         `json:"api_key_secret"`
	TargetUserAttributes UserAttributes `json:"target_user_attributes"`
}

// UpdateUserAttributes updates a user's attributes.
func (api *KasmAPI) UpdateUserAttributes(ctx context.Context, attributes UserAttributes) error {
	endpoint := "/api/public/update_user_attributes"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", attributes.UserID).
		Msg("Updating user attributes")

	// Construct request payload
	requestPayload := UpdateUserAttributesRequest{
		APIKey:               api.APIKey,
		APIKeySecret:         api.APIKeySecret,
		TargetUserAttributes: attributes,
	}

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", attributes.UserID).
			Msg("Failed to update user attributes")
		return fmt.Errorf("failed to update user attributes: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", attributes.UserID).
		Msg("User attributes updated successfully")
	return nil
}

// AddUserToGroupRequest represents the payload for adding a user to a group.
type AddUserToGroupRequest struct {
	APIKey       string               `json:"api_key"`
	APIKeySecret string               `json:"api_key_secret"`
	TargetUser   AddUserToGroupTarget `json:"target_user"`
	TargetGroup  AddUserToGroupTarget `json:"target_group"`
}

// AddUserToGroupTarget represents the target user or group details.
type AddUserToGroupTarget struct {
	UserID  string `json:"user_id,omitempty"`
	GroupID string `json:"group_id,omitempty"`
}

// AddUserToGroup adds a user to a specific group.
func (api *KasmAPI) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	endpoint := "/api/public/add_user_group"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("Adding user to group")

	// Construct request payload
	requestPayload := AddUserToGroupRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: AddUserToGroupTarget{
			UserID: userID,
		},
		TargetGroup: AddUserToGroupTarget{
			GroupID: groupID,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Str("group_id", groupID).
			Msg("Failed to add user to group")
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("User added to group successfully")
	return nil
}

// RemoveUserFromGroupRequest represents the payload for removing a user from a group.
type RemoveUserFromGroupRequest struct {
	APIKey       string                `json:"api_key"`
	APIKeySecret string                `json:"api_key_secret"`
	TargetUser   RemoveUserTargetUser  `json:"target_user"`
	TargetGroup  RemoveUserTargetGroup `json:"target_group"`
}

// RemoveUserTargetUser represents the target user details.
type RemoveUserTargetUser struct {
	UserID string `json:"user_id"`
}

// RemoveUserTargetGroup represents the target group details.
type RemoveUserTargetGroup struct {
	GroupID string `json:"group_id"`
}

// RemoveUserFromGroup removes a user from a specific group.
func (api *KasmAPI) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	endpoint := "/api/public/remove_user_group"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("Removing user from group")

	// Construct request payload
	requestPayload := RemoveUserFromGroupRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: RemoveUserTargetUser{
			UserID: userID,
		},
		TargetGroup: RemoveUserTargetGroup{
			GroupID: groupID,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	_, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Str("group_id", groupID).
			Msg("Failed to remove user from group")
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("User removed from group successfully")
	return nil
}

// GenerateLoginLinkRequest represents the payload for generating a login link.
type GenerateLoginLinkRequest struct {
	APIKey       string                  `json:"api_key"`
	APIKeySecret string                  `json:"api_key_secret"`
	TargetUser   GenerateLoginTargetUser `json:"target_user"`
}

// GenerateLoginTargetUser represents the target user details.
type GenerateLoginTargetUser struct {
	UserID string `json:"user_id"`
}

// GenerateLoginLinkResponse represents the response containing the login URL.
type GenerateLoginLinkResponse struct {
	URL string `json:"url"`
}

// GenerateLoginLink generates a login link for a user.
func (api *KasmAPI) GenerateLoginLink(ctx context.Context, userID string) (string, error) {
	endpoint := "/api/public/get_login"
	log.Info().
		Str("method", "POST").
		Str("endpoint", endpoint).
		Str("user_id", userID).
		Msg("Generating login link")

	// Construct request payload
	requestPayload := GenerateLoginLinkRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
		TargetUser: GenerateLoginTargetUser{
			UserID: userID,
		},
	}

	// Make POST request using the enhanced MakePostRequest method
	responseBytes, err := api.MakePostRequest(ctx, endpoint, requestPayload)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to generate login link")
		return "", fmt.Errorf("failed to generate login link: %w", err)
	}

	// Parse the response into GenerateLoginLinkResponse struct
	var loginResponse GenerateLoginLinkResponse
	if err := json.Unmarshal(responseBytes, &loginResponse); err != nil {
		log.Error().
			Err(err).
			Str("method", "POST").
			Str("endpoint", endpoint).
			Str("user_id", userID).
			Msg("Failed to decode login link response")
		return "", fmt.Errorf("failed to decode login link response: %v", err)
	}

	log.Info().
		Str("user_id", userID).
		Str("login_url", loginResponse.URL).
		Msg("Login link generated successfully")
	return loginResponse.URL, nil
}
