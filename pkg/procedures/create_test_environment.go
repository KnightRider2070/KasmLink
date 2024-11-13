package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/api"
	"kasmlink/pkg/dockercli"
	shadowssh "kasmlink/pkg/ssh"
	"os"
)

// StartTestSession starts KASM sessions for each user, attaches containers to a network, and handles retries.
func StartTestSession(imageTagCore, imageTagFrontend, composeFilePath, remoteNodePath, networkName string, kasmUserMap map[string]string, kasmAPI *api.KasmAPI) (map[string]string, error) {
	ctx := context.Background()
	retries := 3
	sessionIDs := make(map[string]string)

	// Step 1: Establish SSH connection to the target KASM node
	log.Info().Str("nodeAddress", "targetNode").Msg("Establishing SSH connection to the KASM node")
	sshConfig := shadowssh.NewSSHConfigFromFlags()

	sshClient, err := shadowssh.NewSSHClient(sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %v", err)
	}
	defer func() {
		if err := sshClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close SSH client")
		}
	}()
	log.Info().Msg("SSH connection established successfully")

	// Step 2: Check that the latest core image is available
	log.Info().Str("imageTag", imageTagCore).Msg("Checking availability of core Docker image on remote node")
	command := fmt.Sprintf("docker images -q %s", imageTagCore)
	coreImageID, err := shadowssh.ExecuteCommand(sshClient, command)
	if err != nil {
		return nil, fmt.Errorf("failed to check core image availability on remote node: %v", err)
	}

	// Step 3: If the latest core image is not available, build it and load it to the node
	if coreImageID == "" {
		log.Info().Str("imageTag", imageTagCore).Msg("Core image not found, building the core image")
		if err := BuildCoreImageKasm(imageTagCore, "opensuse/leap:15.5"); err != nil {
			return nil, fmt.Errorf("failed to build core image: %v", err)
		}

		// Export core image to tar
		tarFilePath, err := dockercli.ExportImageToTar(ctx, retries, imageTagCore, "")
		if err != nil {
			return nil, fmt.Errorf("failed to export core image as tar: %v", err)
		}
		defer func() {
			if err := os.Remove(tarFilePath); err != nil {
				log.Error().Err(err).Msg("Failed to remove tar file")
			}
		}()

		// Upload tar file to the remote node
		if err := ImportDockerImageToRemoteNode(sshConfig.Username, sshConfig.Password, sshConfig.NodeAddress, tarFilePath, remoteNodePath); err != nil {
			return nil, fmt.Errorf("failed to import core image to remote node: %v", err)
		}
	}

	// Step 4: Use the compose file to create the backend services and create a network
	log.Info().Str("composeFilePath", composeFilePath).Msg("Deploying backend services using Docker Compose")
	if err := DeployComposeFile(composeFilePath, remoteNodePath); err != nil {
		return nil, fmt.Errorf("failed to deploy backend services with Docker Compose: %v", err)
	}

	// Step 5: Find the KASM image ID by using the imageTagFrontend with the KASM API
	log.Info().Str("imageTag", imageTagFrontend).Msg("Searching for KASM image ID using image tag")
	images, err := kasmAPI.ListImages()
	if err != nil {
		return nil, fmt.Errorf("failed to list images from KASM API: %v", err)
	}

	// Step 1: Get the Docker Image ID for the specified tag
	dockerImageId, err := dockercli.GetImageIDByTag(ctx, 3, imageTagFrontend)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker image ID for tag %s: %v", imageTagFrontend, err)
	}

	log.Info().Str("imageTag", imageTagFrontend).Str("dockerImageID", dockerImageId).Msg("Successfully retrieved Docker image ID")

	// Step 2: Search the list of images from KASM API for the matching Image ID
	var kasmImageID string
	for _, image := range images {
		// Compare the Docker Image ID with the KASM API image ImageID
		if image.ImageID == dockerImageId {
			kasmImageID = image.ImageID
			break
		}
	}

	// Step 3: Handle case where no match was found
	if kasmImageID == "" {
		return nil, fmt.Errorf("could not find matching KASM image for Docker image ID: %s", dockerImageId)
	}

	log.Info().Str("kasmImageID", kasmImageID).Msg("Found matching KASM image ID")

	// Step 8: Based on the user map, start KASM sessions for each user using the KasmAPI
	log.Info().Str("network", networkName).Msg("Starting KASM sessions for users using Kasm API")

	// Step 1: Track session IDs for later use
	var kasmSessionIDs []string

	for userID, userName := range kasmUserMap {
		// Create a request for the KASM API
		req := api.RequestKasmRequest{
			UserID:  userID,
			ImageID: kasmImageID, // Use the KASM image ID found from the API
		}

		// Request a new KASM session using the KasmAPI
		kasmResponse, err := kasmAPI.RequestKasmSession(req)
		if err != nil {
			log.Error().Err(err).Str("userID", userID).Msg("Failed to request KASM session for user")
			return nil, fmt.Errorf("failed to request KASM session for user %s: %v", userID, err)
		}

		// Store the session ID for later use
		sessionIDs[userName] = kasmResponse.KasmID
		kasmSessionIDs = append(kasmSessionIDs, kasmResponse.KasmID) // Track the session IDs for the current run

		log.Info().Str("userID", userID).Str("sessionID", kasmResponse.KasmID).Msg("Started KASM session successfully")
	}

	// Step 2: After starting the KASM sessions, attach only the relevant containers to the network
	// Iterate over the list of KASM session IDs created in this run
	for _, kasmSessionID := range kasmSessionIDs {
		log.Info().Str("network", networkName).Str("sessionID", kasmSessionID).Msg("Attaching frontend containers for the current KASM session to the network")

		// Step 1: Get the Kasm session status via the Kasm API
		getStatusRequest := api.GetKasmStatusRequest{
			KasmID: kasmSessionID,
		}

		kasmStatusResponse, err := kasmAPI.GetKasmStatus(getStatusRequest) // Get the status of the KASM session
		if err != nil {
			log.Error().Err(err).Str("kasmSessionID", kasmSessionID).Msg("Failed to retrieve session status")
			return nil, fmt.Errorf("failed to retrieve session status for KASM session %s: %v", kasmSessionID, err)
		}

		// Step 2: Check the operational status of the session
		if kasmStatusResponse.Kasm == nil || kasmStatusResponse.Kasm.ContainerID == "" {
			log.Warn().Str("kasmSessionID", kasmSessionID).Msg("No container found for the specified KASM session")
			continue // Skip if no container ID is found for this session
		}

		containerID := kasmStatusResponse.Kasm.ContainerID

		// Step 3: Check if the session is running, otherwise skip or retry
		switch kasmStatusResponse.OperationalStatus {
		case "starting":
			log.Info().Str("kasmSessionID", kasmSessionID).Msg("Session is starting. Skipping network attachment.")
			continue // Skip if the session is still starting
		case "running":
			log.Info().Str("kasmSessionID", kasmSessionID).Msg("Session is running. Proceeding with network attachment.")
		default:
			log.Warn().Str("kasmSessionID", kasmSessionID).Str("status", kasmStatusResponse.OperationalStatus).Msg("Session is in an unexpected state. Skipping network attachment.")
			continue // Skip if the session is in an unexpected state
		}

		// Step 4: Attach the container to the network
		attachCommand := fmt.Sprintf("docker network connect %s %s", networkName, containerID)
		_, err = shadowssh.ExecuteCommand(sshClient, attachCommand)
		if err != nil {
			log.Error().Err(err).Str("container", containerID).Str("network", networkName).Msg("Failed to attach container to network")
			return nil, fmt.Errorf("failed to attach container %s to network %s: %v", containerID, networkName, err)
		}

		log.Info().Str("container", containerID).Str("network", networkName).Msg("Container attached to network successfully")
	}

	log.Info().Str("network", networkName).Msg("All relevant containers have been attached to the network successfully")

	// Step 10: Return the KASM session IDs for each user
	log.Info().Msg("All KASM sessions started successfully")
	return sessionIDs, nil
}
