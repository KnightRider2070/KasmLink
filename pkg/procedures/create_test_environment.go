package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kasmlink/pkg/api"
	"kasmlink/pkg/dockercli"
	shadowssh "kasmlink/pkg/ssh"
	"kasmlink/pkg/userParser"
	"os"
)

// Config represents the user YAML configuration
// It holds environment arguments and other configurations
type Config struct {
	EnvironmentArgs map[string]string `yaml:"environmentArgs"`
}

type UserYaml struct {
	UserID   string `yaml:"user_id"`
	UserName string `yaml:"user_name"`
}

// StartTestSession starts KASM sessions for each user, attaches containers to a network, and handles retries
func StartTestSession(imageTagCore, imageTagFrontend, composeFilePath, remoteNodePath, networkName, yamlFilePath, userYamlPath string, kasmAPI *api.KasmAPI) (map[string]string, error) {
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

	// Step 2: Load environment arguments from the general YAML file
	log.Info().Str("yamlFilePath", yamlFilePath).Msg("Parsing environment arguments from YAML file")
	config := &Config{}
	yamlContent, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment YAML file: %v", err)
	}

	err = yaml.Unmarshal(yamlContent, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment YAML file: %v", err)
	}
	log.Info().Interface("environmentArgs", config.EnvironmentArgs).Msg("Parsed environment arguments successfully")

	// Step 3: Load user-specific details from the user YAML file
	users, err := userParser.LoadUsersFromYaml(userYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load users from YAML: %v", err)
	}

	// Step 4: Check if the core Docker image is available on the remote node
	log.Info().Str("imageTag", imageTagCore).Msg("Checking availability of core Docker image on remote node")
	command := fmt.Sprintf("docker images -q %s", imageTagCore)
	coreImageID, err := shadowssh.ExecuteCommand(sshClient, command)
	if err != nil {
		return nil, fmt.Errorf("failed to check core image availability on remote node: %v", err)
	}

	// Step 5: If the image is not found on the remote node, check for a local tar
	if coreImageID == "" {
		tarFilePath := fmt.Sprintf("./%s.tar", imageTagCore)
		if _, err := os.Stat(tarFilePath); err == nil {
			log.Info().Str("tarFilePath", tarFilePath).Msg("Core Docker image tar found locally, uploading to remote node")
			if err := ImportDockerImageToRemoteNode(sshConfig.Username, sshConfig.Password, sshConfig.NodeAddress, tarFilePath, remoteNodePath); err != nil {
				return nil, fmt.Errorf("failed to import core image from local tar to remote node: %v", err)
			}
		} else {
			// Step 5b: If no local tar is found, build the image
			log.Info().Str("imageTag", imageTagCore).Msg("Core image not found locally or remotely; building image")
			if err := BuildCoreImageKasm(imageTagCore, "opensuse/leap:15.5"); err != nil {
				return nil, fmt.Errorf("failed to build core image: %v", err)
			}

			// Export core image to tar and upload
			tarFilePath, err = dockercli.ExportImageToTar(ctx, retries, imageTagCore, "")
			if err != nil {
				return nil, fmt.Errorf("failed to export core image as tar: %v", err)
			}
			defer os.Remove(tarFilePath)

			if err := ImportDockerImageToRemoteNode(sshConfig.Username, sshConfig.Password, sshConfig.NodeAddress, tarFilePath, remoteNodePath); err != nil {
				return nil, fmt.Errorf("failed to import core image to remote node: %v", err)
			}
		}
	}

	// Step 6: Use the compose file to create backend services and network
	log.Info().Str("composeFilePath", composeFilePath).Msg("Deploying backend services using Docker Compose")
	if err := DeployComposeFile(composeFilePath, remoteNodePath); err != nil {
		return nil, fmt.Errorf("failed to deploy backend services with Docker Compose: %v", err)
	}

	// Step 7: Retrieve the Docker image ID for the frontend image tag via the KASM API
	log.Info().Str("imageTag", imageTagFrontend).Msg("Searching for KASM image ID using image tag")
	images, err := kasmAPI.ListImages()
	if err != nil {
		return nil, fmt.Errorf("failed to list images from KASM API: %v", err)
	}

	dockerImageId, err := dockercli.GetImageIDByTag(ctx, retries, imageTagFrontend)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker image ID for tag %s: %v", imageTagFrontend, err)
	}

	var kasmImageID string
	for _, image := range images {
		if image.ImageID == dockerImageId {
			kasmImageID = image.ImageID
			break
		}
	}
	if kasmImageID == "" {
		return nil, fmt.Errorf("could not find matching KASM image for Docker image ID: %s", dockerImageId)
	}

	// Step 8: For each user, request a KASM session with the specified environment args and attach containers to the network
	log.Info().Str("network", networkName).Msg("Starting KASM sessions for users using Kasm API")
	var kasmSessionIDs []string

	for _, user := range users {
		req := api.RequestKasmRequest{
			UserID:      user.UserID,
			ImageID:     kasmImageID,
			Environment: user.EnvironmentArgs, // Use user-specific environment args
		}
		kasmResponse, err := kasmAPI.RequestKasmSession(req)
		if err != nil {
			return nil, fmt.Errorf("failed to request KASM session for user %s: %v", user.UserID, err)
		}
		sessionIDs[user.UserName] = kasmResponse.KasmID
		kasmSessionIDs = append(kasmSessionIDs, kasmResponse.KasmID)
	}

	for _, kasmSessionID := range kasmSessionIDs {
		getStatusRequest := api.GetKasmStatusRequest{KasmID: kasmSessionID}
		kasmStatusResponse, err := kasmAPI.GetKasmStatus(getStatusRequest)
		if err != nil || kasmStatusResponse.Kasm == nil || kasmStatusResponse.Kasm.ContainerID == "" {
			continue
		}

		containerID := kasmStatusResponse.Kasm.ContainerID
		if kasmStatusResponse.OperationalStatus == "running" {
			attachCommand := fmt.Sprintf("docker network connect %s %s", networkName, containerID)
			_, err = shadowssh.ExecuteCommand(sshClient, attachCommand)
			if err != nil {
				return nil, fmt.Errorf("failed to attach container %s to network %s: %v", containerID, networkName, err)
			}
		}
	}

	log.Info().Msg("All KASM sessions started and containers attached to the network successfully")
	return sessionIDs, nil
}
