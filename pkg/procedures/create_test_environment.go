package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api"
	"kasmlink/pkg/userParser"
)

func processUsers(kapi *api.KasmAPI, yamlFilePath string, groupID string) error {
	log.Info().Str("yamlFilePath", yamlFilePath).Msg("Starting user processing")

	// Step 1: Load configuration
	config, err := userParser.LoadConfig(yamlFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load YAML configuration")
		return fmt.Errorf("failed to load YAML configuration: %w", err)
	}
	log.Info().Int("userCount", len(config.UserDetails)).Msg("Loaded user configuration successfully")

	// Step 2: Process each user
	for _, user := range config.UserDetails {
		log.Debug().Str("username", user.Username).Msg("Processing user")
		var userID string

		// Step 2a: Check if the user already exists
		existingUser, err := kapi.GetUser("", user.Username)
		if err != nil {
			// If user does not exist, create it
			log.Debug().Str("username", user.Username).Msg("User not found, creating new user")
			newUser := api.TargetUser{
				Username:     user.Username,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
				Organization: user.Organization,
				Phone:        user.Phone,
				Password:     user.Password,
				Locked:       user.Locked,
				Disabled:     user.Disabled,
			}
			createdUser, err := kapi.CreateUser(newUser)
			if err != nil {
				log.Error().Err(err).Str("username", user.Username).Msg("Error creating user")
				continue
			}
			userID = createdUser.UserID
			log.Info().Str("username", user.Username).Str("userID", userID).Msg("User created successfully")
		} else {
			// If user exists, retrieve their userId
			userID = existingUser.UserID
			log.Info().Str("username", user.Username).Str("userID", userID).Msg("User already exists")
		}

		// Step 2b: Add user to the specified group
		err = kapi.AddUserToGroup(userID, groupID)
		if err != nil {
			log.Error().Err(err).Str("username", user.Username).Str("groupID", groupID).Msg("Error adding user to group")
		} else {
			log.Info().Str("username", user.Username).Str("groupID", groupID).Msg("User added to group successfully")
		}

		// Step 2c: Update the YAML file with userId and KasmSessionOfContainer
		err = userParser.UpdateUserConfig(yamlFilePath, user.Username, userID, user.KasmSessionOfContainer)
		if err != nil {
			log.Error().Err(err).Str("username", user.Username).Msg("Failed to update user configuration in YAML")
		} else {
			log.Debug().Str("username", user.Username).Msg("User configuration updated successfully in YAML")
		}
	}

	log.Info().Str("yamlFilePath", yamlFilePath).Msg("User processing completed successfully")
	return nil
}
