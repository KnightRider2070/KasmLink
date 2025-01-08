package userService

import (
	"fmt"
	"kasmlink/pkg/api/base"
	"kasmlink/pkg/api/http"
	"kasmlink/pkg/api/models"

	"github.com/rs/zerolog/log"
)

const (
	CreateUserEndpoint           = "/api/public/create_user"
	GetUsersEndpoint             = "/api/public/get_users"
	GetUserEndpoint              = "/api/public/get_user"
	UpdateUserEndpoint           = "/api/public/update_user"
	DeleteUserEndpoint           = "/api/public/delete_user"
	LogoutUserEndpoint           = "/api/public/logout_user"
	GetUserAttributesEndpoint    = "/api/public/get_attributes"
	UpdateUserAttributesEndpoint = "/api/public/update_attributes"
	AddUserToGroupEndpoint       = "/api/public/add_user_group"
	RemoveUserFromGroupEndpoint  = "/api/public/remove_user_group"
	GenerateLoginLinkEndpoint    = "/api/public/get_login"
)

// UserService provides methods to manage users.
type UserService struct {
	*base.BaseService
}

// NewUserService creates a new instance of UserService.
func NewUserService(handler http.RequestHandler) *UserService {
	return &UserService{
		BaseService: base.NewBaseService(handler),
	}
}

// CreateUser sends a request to create a userService.
func (us *UserService) CreateUser(user models.TargetUser) (*models.UserResponse, error) {
	url := fmt.Sprintf("%s%s", us.BaseURL, CreateUserEndpoint)
	log.Info().Str("url", url).Str("username", user.Username).Msg("Creating new userService.")

	payload := us.BuildPayload(map[string]interface{}{
		"target_user": user,
	})

	var createdUser models.UserResponse
	if err := us.ExecuteRequest(url, payload, &createdUser); err != nil {
		log.Error().Err(err).Msg("Failed to create userService.")
		return nil, err
	}

	log.Info().Str("user_id", createdUser.UserID).Msg("User created successfully.")
	return &createdUser, nil
}

// GetUsers retrieves a list of all users.
func (us *UserService) GetUsers() ([]models.UserResponse, error) {
	url := fmt.Sprintf("%s%s", us.BaseURL, GetUsersEndpoint)
	log.Info().Str("url", url).Msg("Fetching all users.")

	payload := us.BuildPayload(nil)

	var parsedResponse struct {
		Users []models.UserResponse `json:"users"`
	}
	if err := us.ExecuteRequest(url, payload, &parsedResponse); err != nil {
		log.Error().Err(err).Msg("Failed to fetch users.")
		return nil, err
	}

	log.Info().Int("user_count", len(parsedResponse.Users)).Msg("Users retrieved successfully.")
	return parsedResponse.Users, nil
}

// GetUser retrieves userService details by userID or username.
func (us *UserService) GetUser(userID, username string) (*models.UserResponse, error) {
	url := fmt.Sprintf("%s%s", us.BaseURL, GetUserEndpoint)
	log.Info().
		Str("url", url).
		Str("user_id", userID).
		Str("username", username).
		Msg("Fetching userService details.")

	payload := us.BuildPayload(map[string]interface{}{
		"target_user": map[string]string{
			"user_id":  userID,
			"username": username,
		},
	})

	var user models.UserResponse
	if err := us.ExecuteRequest(url, payload, &user); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to fetch userService details.")
		return nil, err
	}

	log.Info().Str("user_id", user.UserID).Msg("User details retrieved successfully.")
	return &user, nil
}

// UpdateUser updates an existing userService's details.
func (us *UserService) UpdateUser(user models.TargetUser) (*models.UserResponse, error) {
	url := fmt.Sprintf("%s%s", us.BaseURL, UpdateUserEndpoint)
	log.Info().
		Str("url", url).
		Str("user_id", user.UserID).
		Msg("Updating userService details.")

	payload := us.BuildPayload(map[string]interface{}{
		"target_user": user,
	})

	var updatedUser models.UserResponse
	if err := us.ExecuteRequest(url, payload, &updatedUser); err != nil {
		log.Error().Err(err).Str("user_id", user.UserID).Msg("Failed to update userService.")
		return nil, err
	}

	log.Info().Str("user_id", updatedUser.UserID).Msg("User updated successfully.")
	return &updatedUser, nil
}

// DeleteUser removes a userService by userID with optional force.
func (us *UserService) DeleteUser(userID string, force bool) error {
	url := fmt.Sprintf("%s%s", us.BaseURL, DeleteUserEndpoint)
	log.Info().
		Str("url", url).
		Str("user_id", userID).
		Bool("force", force).
		Msg("Deleting userService.")

	payload := us.BuildPayload(map[string]interface{}{
		"target_user": map[string]string{
			"user_id": userID,
		},
		"force": force,
	})

	if err := us.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete userService.")
		return err
	}

	log.Info().Str("user_id", userID).Msg("User deleted successfully.")
	return nil
}
