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
	GetGroups                    = "api/public/get_groups"        //Undocumented Api
	CreateGroupEndpoint          = "/api/public/create_group"     //Undocumented Api
	AddImagesToGroupEndpoint     = "/api/public/add_images_group" //Undocumented Api

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

func (us *UserService) GetGroups() (models.GroupsResponse, error) {
	// Build the URL
	url := fmt.Sprintf("%s%s", us.BaseURL, GetGroups)
	log.Debug().Str("url", url).Msg("Constructed URL for GetGroups")

	// Build the payload
	payload := us.BuildPayload(nil)
	log.Debug().Interface("payload", payload).Msg("Payload for GetGroups")

	var parsedResponse models.GroupsResponse

	// Execute the request
	if err := us.ExecuteRequest(url, payload, &parsedResponse); err != nil {
		log.Error().
			Err(err).
			Str("url", url).
			Interface("payload", payload).
			Msg("Failed to execute request for GetGroups")
		return models.GroupsResponse{}, fmt.Errorf("failed to get groups: %w", err)
	}

	log.Info().
		Int("group_count", len(parsedResponse.Groups)).
		Msg("Successfully fetched groups")

	return parsedResponse, nil
}

// AddUserToGroup adds a user to a specified group.
func (us *UserService) AddUserToGroup(userID, groupID string) error {
	url := fmt.Sprintf("%s%s", us.BaseURL, AddUserToGroupEndpoint)
	log.Info().
		Str("url", url).
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("Adding user to group.")

	// Construct the payload
	payload := us.BuildPayload(map[string]interface{}{
		"target_user": map[string]string{
			"user_id": userID,
		},
		"target_group": map[string]string{
			"group_id": groupID,
		},
	})

	// Execute the request
	if err := us.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("group_id", groupID).
			Msg("Failed to add user to group.")
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	log.Info().
		Str("user_id", userID).
		Str("group_id", groupID).
		Msg("User successfully added to group.")
	return nil
}

// CreateGroup sends a request to create a new group.
func (us *UserService) CreateGroup(group models.Group) (*models.GroupsResponse, error) {
	url := fmt.Sprintf("%s%s", us.BaseURL, CreateGroupEndpoint)
	log.Info().
		Str("url", url).
		Str("group_name", group.Name).
		Msg("Creating new group.")

	// Construct the payload
	payload := us.BuildPayload(map[string]interface{}{
		"target_group": group,
	})

	var createdGroup models.GroupsResponse
	// Execute the request
	if err := us.ExecuteRequest(url, payload, &createdGroup); err != nil {
		log.Error().
			Err(err).
			Str("group_name", group.Name).
			Msg("Failed to create group.")
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	log.Info().
		Str("group_id", createdGroup.Groups[0].GroupID).
		Str("group_name", createdGroup.Groups[0].Name).
		Msg("Group created successfully.")
	return &createdGroup, nil
}

// AddImageToGroup sends a request to associate an image with a group.
func (us *UserService) AddImageToGroup(groupID, imageID string) error {
	url := fmt.Sprintf("%s%s", us.BaseURL, AddImagesToGroupEndpoint)
	log.Info().
		Str("url", url).
		Str("group_id", groupID).
		Str("image_id", imageID).
		Msg("Adding image to group.")

	// Construct the payload
	payload := us.BuildPayload(map[string]interface{}{
		"target_group": map[string]string{
			"group_id": groupID,
		},
		"target_image": map[string]string{
			"image_id": imageID,
		},
	})

	// Execute the request
	if err := us.ExecuteRequest(url, payload, nil); err != nil {
		log.Error().
			Err(err).
			Str("group_id", groupID).
			Str("image_id", imageID).
			Msg("Failed to add image to group.")
		return fmt.Errorf("failed to add image to group: %w", err)
	}

	log.Info().
		Str("group_id", groupID).
		Str("image_id", imageID).
		Msg("Image successfully added to group.")
	return nil
}
