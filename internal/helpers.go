package internal

import (
	"context"
	"fmt"
	"kasmlink/pkg/api/models"
	"kasmlink/pkg/api/user"
	"strings"

	"github.com/rs/zerolog/log"
	"kasmlink/pkg/shadowssh"
	"kasmlink/pkg/userParser"
)

// CheckRemoteImages verifies missing Docker images on the remote server.
func CheckRemoteImages(ctx context.Context, client *shadowssh.Client, images []string) ([]string, error) {
	log.Debug().Msg("Executing command to list available Docker images on remote node.")

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	output, err := client.ExecuteCommand(ctx, cmd)
	if err != nil {
		log.Error().Err(err).Str("command", cmd).Msg("Failed to list remote Docker images.")
		return nil, fmt.Errorf("failed to execute remote Docker images command: %w", err)
	}

	remoteImages := strings.Split(output, "\n")
	var missing []string
	imageSet := make(map[string]struct{})

	for _, img := range remoteImages {
		trimmedImg := strings.TrimSpace(img)
		if trimmedImg != "" {
			imageSet[trimmedImg] = struct{}{}
		}
	}

	for _, img := range images {
		if _, exists := imageSet[img]; !exists {
			missing = append(missing, img)
			log.Debug().Str("image", img).Msg("Image is missing on the remote node.")
		}
	}

	return missing, nil
}

// CreateOrGetUser ensures a user exists in the system, creating if necessary.
func CreateOrGetUser(ctx context.Context, api *user.UserService, user userParser.UserDetails) (string, error) {
	log.Info().Str("username", user.TargetUser.Username).Msg("Ensuring user exists via API.")

	userExisting, err := api.GetUser(user.TargetUser.UserID, user.TargetUser.Username)
	if err == nil {
		log.Info().Str("username", userExisting.Username).Str("user_id", userExisting.UserID).Msg("User already exists.")
		return userExisting.UserID, nil
	}

	log.Info().Str("username", user.TargetUser.Username).Msg("User not found. Proceeding with creation.")
	targetUser := models.TargetUser{
		Username:     user.TargetUser.Username,
		FirstName:    user.TargetUser.FirstName,
		LastName:     user.TargetUser.LastName,
		Locked:       user.TargetUser.Locked,
		Disabled:     user.TargetUser.Disabled,
		Organization: user.TargetUser.Organization,
		Phone:        user.TargetUser.Phone,
		Password:     user.TargetUser.Password,
	}

	createdUser, err := api.CreateUser(targetUser)
	if err != nil {
		log.Error().Err(err).Str("username", user.TargetUser.Username).Msg("Failed to create user.")
		return "", fmt.Errorf("failed to create user %s: %w", user.TargetUser.Username, err)
	}

	log.Info().Str("username", createdUser.Username).Str("user_id", createdUser.UserID).Msg("User created successfully.")
	return createdUser.UserID, nil
}

// ParseVolumeMounts validates and converts volume mounts.
func ParseVolumeMounts(details userParser.UserDetails) (map[string]models.VolumeMapping, error) {
	volumeMappings := make(map[string]models.VolumeMapping)

	for hostPath, containerPathAndMode := range details.VolumeMounts {
		parts := strings.Split(containerPathAndMode, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid volume mount format: %s, expected 'containerPath:mode'", containerPathAndMode)
		}

		containerPath, mode := parts[0], parts[1]
		if mode != "rw" && mode != "ro" {
			return nil, fmt.Errorf("invalid volume mount mode: %s, expected 'rw' or 'ro'", mode)
		}

		volumeMappings[containerPath] = models.VolumeMapping{
			Bind: hostPath,
			Mode: mode,
			Gid:  1000,
			Uid:  1000,
		}
	}

	return volumeMappings, nil
}
