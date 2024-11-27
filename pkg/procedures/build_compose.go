package procedures

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
	"io/fs"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercli"
	"kasmlink/pkg/dockercompose"
	shadowscp "kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/ssh"
	"os"
	"path/filepath"
	"strings"
)

// InitFolder initializes a specified folder by copying embedded templates or Dockerfiles.
func InitFolder(folderPath, subfolder, sourcePath string, embeddedFS fs.FS) error {
	targetFolder := filepath.Join(folderPath, subfolder)
	log.Info().Str("folderPath", folderPath).Str("subfolder", subfolder).Msg("Initializing folder path")

	// Create the target folder if it doesn’t exist
	if err := os.MkdirAll(targetFolder, os.ModePerm); err != nil {
		log.Error().Err(err).Str("path", targetFolder).Msg("Failed to create target folder")
		return fmt.Errorf("failed to create folder %s: %w", targetFolder, err)
	}

	// Copy files from embedded FS to target folder
	if err := copyEmbeddedFiles(embeddedFS, sourcePath, targetFolder); err != nil {
		log.Error().Err(err).Str("subfolder", subfolder).Msg("Error during folder initialization")
		return fmt.Errorf("error initializing folder %s: %w", subfolder, err)
	}

	log.Info().Str("folderPath", targetFolder).Msg("Folder initialization completed successfully")
	return nil
}

// copyEmbeddedFiles copies files from an embedded file system to a target directory.
func copyEmbeddedFiles(embeddedFS fs.FS, sourcePath, targetFolder string) error {
	return fs.WalkDir(embeddedFS, sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Error walking through embedded directory")
			return err
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		// Prepare target file path and create necessary directories
		relativePath := strings.TrimPrefix(path, sourcePath+"/")
		targetPath := filepath.Join(targetFolder, relativePath)
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to create directory for file")
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Copy file content from embedded FS to target path
		content, err := fs.ReadFile(embeddedFS, path)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Failed to read embedded file")
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			log.Error().Err(err).Str("path", targetPath).Msg("Failed to write file to target path")
			return fmt.Errorf("failed to write file to %s: %w", targetPath, err)
		}
		log.Info().Str("file", targetPath).Msg("File initialized successfully")
		return nil
	})
}

// InitTemplatesFolder initializes the templates folder with embedded templates.
func InitTemplatesFolder(folderPath string) error {
	return InitFolder(folderPath, "services", "services", embedfiles.EmbeddedServicesFS)
}

// InitDockerfilesFolder initializes the Dockerfiles folder with embedded Dockerfile templates.
func InitDockerfilesFolder(folderPath string) error {
	return InitFolder(folderPath, "dockerfiles", "dockerfiles", embedfiles.EmbeddedDockerImagesDirectory)
}

func MergeComposeFiles(file1, file2 dockercompose.ComposeFile) (dockercompose.ComposeFile, error) {
	// Check if versions are compatible
	if file1.Version != "" && file2.Version != "" && file1.Version != file2.Version {
		log.Error().
			Str("file1_version", file1.Version).
			Str("file2_version", file2.Version).
			Msg("incompatible compose file versions")
		return dockercompose.ComposeFile{}, fmt.Errorf("incompatible compose file versions: %s and %s", file1.Version, file2.Version)
	}

	// Use file2's version if file1's version is empty
	version := file1.Version
	if version == "" {
		version = file2.Version
	}

	// Initialize the merged ComposeFile
	merged := dockercompose.ComposeFile{
		Version:  version,
		Services: make(map[string]dockercompose.Service),
		Networks: make(map[string]dockercompose.Network),
		Volumes:  make(map[string]dockercompose.Volume),
		Configs:  make(map[string]dockercompose.Config),
		Secrets:  make(map[string]dockercompose.Secret),
	}

	// Merge services
	log.Debug().
		Interface("file1_services", file1.Services).
		Interface("file2_services", file2.Services).
		Msg("merging services")
	for name, service := range file1.Services {
		merged.Services[name] = service
	}
	for name, service := range file2.Services {
		if existingService, exists := merged.Services[name]; exists {
			log.Debug().Str("service_name", name).Msg("merging existing service")
			merged.Services[name] = existingService
		} else {
			merged.Services[name] = service
		}
	}

	// Merge networks
	log.Debug().
		Interface("file1_networks", file1.Networks).
		Interface("file2_networks", file2.Networks).
		Msg("merging networks")
	for name, network := range file1.Networks {
		merged.Networks[name] = network
	}
	for name, network := range file2.Networks {
		if _, exists := merged.Networks[name]; !exists {
			merged.Networks[name] = network
		}
	}

	// Merge volumes
	log.Debug().
		Interface("file1_volumes", file1.Volumes).
		Interface("file2_volumes", file2.Volumes).
		Msg("merging volumes")
	for name, volume := range file1.Volumes {
		merged.Volumes[name] = volume
	}
	for name, volume := range file2.Volumes {
		if _, exists := merged.Volumes[name]; !exists {
			merged.Volumes[name] = volume
		}
	}

	// Merge configs
	for name, config := range file1.Configs {
		merged.Configs[name] = config
	}
	for name, config := range file2.Configs {
		if _, exists := merged.Configs[name]; !exists {
			merged.Configs[name] = config
		}
	}

	// Merge secrets
	for name, secret := range file1.Secrets {
		merged.Secrets[name] = secret
	}
	for name, secret := range file2.Secrets {
		if _, exists := merged.Secrets[name]; !exists {
			merged.Secrets[name] = secret
		}
	}

	log.Debug().Interface("merged_compose_file", merged).Msg("merge completed")
	return merged, nil
}

// CreateServiceReplicas creates replicas of a service with modified names in a Compose file.
func CreateServiceReplicas(composeFile *dockercompose.ComposeFile, replicas int, inputNames []string) error {
	// Ensure there is exactly one service in the compose file
	if len(composeFile.Services) != 1 {
		return fmt.Errorf("expected exactly one service in the compose file, found %d", len(composeFile.Services))
	}

	// Retrieve the single service
	var originalServiceName string
	var originalService dockercompose.Service
	for name, service := range composeFile.Services {
		originalServiceName = name
		originalService = service
		break
	}

	// Validate input names
	if len(inputNames) == 0 {
		return fmt.Errorf("no input names provided")
	}

	// Generate names for replicas
	replicaNames := make([]string, replicas)
	if len(inputNames) == 1 {
		// If only one name is provided, generate names using name-i format
		baseName := inputNames[0]
		for i := 0; i < replicas; i++ {
			replicaNames[i] = fmt.Sprintf("%s-%d", baseName, i+1)
		}
	} else if len(inputNames) == replicas {
		// If enough names are provided, use them directly
		copy(replicaNames, inputNames)
	} else {
		return fmt.Errorf("number of input names (%d) does not match the number of replicas (%d)", len(inputNames), replicas)
	}

	// Create replicas in the compose file
	for _, replicaName := range replicaNames {
		newService := originalService
		newService.ContainerName = replicaName // Update the container name
		composeFile.Services[replicaName] = newService
	}

	// Remove the original service
	delete(composeFile.Services, originalServiceName)

	return nil
}

// WriteComposeFile writes the provided ComposeFile object to a specified file path.
func WriteComposeFile(composeFile *dockercompose.ComposeFile, filePath string) error {
	// Open the file for writing (create or truncate)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Encode the ComposeFile as YAML
	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(composeFile); err != nil {
		return fmt.Errorf("failed to write compose file to %s: %w", filePath, err)
	}

	return nil
}

// DeployBackendServices deploys backend services based on the provided Docker Compose file and SSH configuration.
func DeployBackendServices(backendComposePath string, sshConfig shadowssh.SSHConfig) error {
	// Step 1: Check if the Docker Compose file exists locally
	log.Info().
		Str("path", backendComposePath).
		Msg("Checking existence of Docker Compose file")

	if _, err := os.Stat(backendComposePath); os.IsNotExist(err) {
		log.Error().
			Err(err).
			Str("path", backendComposePath).
			Msg("Compose file does not exist")
		return fmt.Errorf("compose file does not exist at path: %s", backendComposePath)
	}

	// Step 2: Establish SSH connection with remote node using sshConfig
	log.Info().
		Str("host", sshConfig.NodeAddress).
		Str("user", sshConfig.Username).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
			Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().
				Err(cerr).
				Msg("Failed to close SSH connection gracefully")
		} else {
			log.Debug().
				Msg("SSH connection closed")
		}
	}()

	// Step 3: Load compose into structs
	log.Info().
		Str("path", backendComposePath).
		Msg("Loading Docker Compose file")

	compose, err := dockercompose.LoadComposeFile(backendComposePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", backendComposePath).
			Msg("Failed to load Docker Compose file")
		return fmt.Errorf("failed to load compose file: %w", err)
	}

	// Step 3.2: Extract image names used
	imageNames := make([]string, 0)
	serviceNames := make([]string, 0)
	for serviceName, service := range compose.Services {
		serviceNames = append(serviceNames, serviceName)
		imageNames = append(imageNames, service.Image)
	}
	log.Debug().
		Int("service_count", len(serviceNames)).
		Int("image_count", len(imageNames)).
		Msg("Extracted service and image names from Compose file")

	// Step 3.3: Check which images are present on the remote node
	log.Info().
		Msg("Checking for missing Docker images on the remote node")

	missingImages, err := checkRemoteImages(client, imageNames)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check remote Docker images")
		return fmt.Errorf("failed to check remote images: %w", err)
	}

	if len(missingImages) > 0 {
		log.Info().
			Int("missing_images_count", len(missingImages)).
			Msg("Identified missing Docker images on remote node")

		// Ensure buildTars directory exists locally
		buildTarsDir := "./buildTars"
		if _, err := os.Stat(buildTarsDir); os.IsNotExist(err) {
			log.Info().
				Str("directory", buildTarsDir).
				Msg("Creating buildTars directory")

			err := os.Mkdir(buildTarsDir, 0755)
			if err != nil {
				log.Error().
					Err(err).
					Str("directory", buildTarsDir).
					Msg("Failed to create buildTars directory")
				return fmt.Errorf("failed to create buildTars directory: %w", err)
			}
		}

		for _, image := range missingImages {
			tarPath := filepath.Join(buildTarsDir, fmt.Sprintf("%s.tar", sanitizeImageName(image)))
			log.Debug().
				Str("image", image).
				Str("tar_path", tarPath).
				Msg("Processing missing image")

			if _, err := os.Stat(tarPath); os.IsNotExist(err) {
				// Step 3.4: Check for Dockerfile and build image if necessary
				log.Info().
					Str("image", image).
					Msg("Docker image tar not found locally, searching for Dockerfile")

				dockerfilePath, err := findDockerfileForService(image)
				if err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to find Dockerfile for image")
					return fmt.Errorf("failed to find Dockerfile for image %s: %w", image, err)
				}

				// Build the image locally
				log.Info().
					Str("image", image).
					Str("dockerfile", dockerfilePath).
					Msg("Building Docker image locally")

				if err := dockercli.BuildDockerImage(dockerfilePath, image); err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Msg("Failed to build Docker image")
					return fmt.Errorf("failed to build image %s: %w", image, err)
				}

				// Export the image to a tar file
				log.Info().
					Str("image", image).
					Str("tar_path", tarPath).
					Msg("Exporting Docker image to tar")

				exportedTar, err := dockercli.ExportImageToTar(context.Background(), 3, image, tarPath)
				if err != nil {
					log.Error().
						Err(err).
						Str("image", image).
						Str("tar_path", tarPath).
						Msg("Failed to export Docker image to tar")
					return fmt.Errorf("failed to export image %s to tar: %w", image, err)
				}
				log.Info().
					Str("image", image).
					Str("tar_path", exportedTar).
					Msg("Successfully exported Docker image to tar")
			}

			// Step 3.5 & 3.4: Copy the tar onto the remote node into /tmp
			remoteTmpDir := "/tmp"
			log.Info().
				Str("tar_path", tarPath).
				Str("remote_dir", remoteTmpDir).
				Msg("Copying tar file to remote node")

			if err := shadowscp.ShadowCopyFile(tarPath, remoteTmpDir); err != nil {
				log.Error().
					Err(err).
					Str("tar_path", tarPath).
					Str("remote_dir", remoteTmpDir).
					Msg("Failed to copy tar file to remote node")
				return fmt.Errorf("failed to copy tar %s to remote: %w", tarPath, err)
			}

			// Step 3.6: Load the docker image on the remote node
			loadCmd := fmt.Sprintf("docker load -i %s/%s.tar", remoteTmpDir, sanitizeImageName(image))
			log.Info().
				Str("image", image).
				Str("command", loadCmd).
				Msg("Loading Docker image on remote node")

			_, err = shadowssh.ExecuteCommand(client, loadCmd)
			if err != nil {
				log.Error().
					Err(err).
					Str("image", image).
					Str("command", loadCmd).
					Msg("Failed to load Docker image on remote node")
				return fmt.Errorf("failed to load image %s on remote: %w", image, err)
			}

			log.Info().
				Str("image", image).
				Msg("Successfully loaded Docker image on remote node")
		}
	} else {
		log.Info().
			Msg("All Docker images are already present on the remote node")
	}

	// Step 3.7: Copy compose file onto the node to /composefiles
	remoteComposeDir := "/composefiles"
	log.Info().
		Str("compose_path", backendComposePath).
		Str("remote_dir", remoteComposeDir).
		Msg("Copying Docker Compose file to remote node")

	if err := shadowscp.ShadowCopyFile(backendComposePath, remoteComposeDir); err != nil {
		log.Error().
			Err(err).
			Str("compose_path", backendComposePath).
			Str("remote_dir", remoteComposeDir).
			Msg("Failed to copy Docker Compose file to remote node")
		return fmt.Errorf("failed to copy compose file to remote: %w", err)
	}

	// Step 3.8: Execute 'docker compose up' on the remote node
	composeUpCmd := fmt.Sprintf("cd %s && docker compose up -d", remoteComposeDir)
	log.Info().
		Str("command", composeUpCmd).
		Msg("Executing 'docker compose up' on remote node")

	_, err = shadowssh.ExecuteCommand(client, composeUpCmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", composeUpCmd).
			Msg("Failed to execute 'docker compose up' on remote node")
		return fmt.Errorf("failed to execute docker compose up: %w", err)
	}

	log.Info().
		Msg("Deployment completed successfully")
	return nil
}

// checkRemoteImages checks which images are missing on the remote node.
func checkRemoteImages(client *ssh.Client, images []string) ([]string, error) {
	log.Debug().
		Msg("Executing remote Docker images command to list available images")

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	output, err := shadowssh.ExecuteCommand(client, cmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", cmd).
			Msg("Failed to execute remote Docker images command")
		return nil, err
	}

	remoteImages := strings.Split(output, "\n")
	missing := []string{}

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
			log.Debug().
				Str("image", img).
				Msg("Image is missing on remote node")
		}
	}

	return missing, nil
}

// findDockerfileForService searches for a Dockerfile in the ./dockerfiles/ directory that contains the serviceName.
func findDockerfileForService(serviceName string) (string, error) {
	log.Debug().
		Str("service_name", serviceName).
		Msg("Searching for Dockerfile matching service name")

	dockerfilesDir := "./dockerfiles"
	pattern := fmt.Sprintf("*%s*", serviceName)

	matchedFiles, err := filepath.Glob(filepath.Join(dockerfilesDir, pattern))
	if err != nil {
		log.Error().
			Err(err).
			Str("pattern", pattern).
			Msg("Failed to glob Dockerfiles")
		return "", fmt.Errorf("failed to glob dockerfiles: %w", err)
	}

	if len(matchedFiles) == 0 {
		log.Error().
			Str("service_name", serviceName).
			Str("directory", dockerfilesDir).
			Msg("No Dockerfile found containing service name")
		return "", fmt.Errorf("no Dockerfile found containing service name '%s' in %s", serviceName, dockerfilesDir)
	}

	if len(matchedFiles) > 1 {
		log.Error().
			Str("service_name", serviceName).
			Strs("matched_files", matchedFiles).
			Msg("Multiple Dockerfiles found for service name")
		return "", fmt.Errorf("multiple Dockerfiles found for service '%s' in %s: %v", serviceName, dockerfilesDir, matchedFiles)
	}

	log.Debug().
		Str("dockerfile", matchedFiles[0]).
		Msg("Found matching Dockerfile")

	return matchedFiles[0], nil
}

// sanitizeImageName sanitizes the image name to create valid filenames.
func sanitizeImageName(imageName string) string {
	// Replace '/' and ':' with underscores to prevent directory traversal or invalid filenames.
	return strings.ReplaceAll(strings.ReplaceAll(imageName, "/", "_"), ":", "_")
}

// DeployImages deploys a Docker image to the remote node based on the provided Dockerfile path.
func DeployImages(dockerFilePath string, imageName string, sshConfig shadowssh.SSHConfig) error {
	// Step 1: Check if the Dockerfile exists locally
	log.Info().
		Str("dockerfile_path", dockerFilePath).
		Msg("Checking existence of Dockerfile")

	if _, err := os.Stat(dockerFilePath); os.IsNotExist(err) {
		log.Error().
			Err(err).
			Str("dockerfile_path", dockerFilePath).
			Msg("Dockerfile does not exist")
		return fmt.Errorf("Dockerfile does not exist at path: %s", dockerFilePath)
	}

	// Step 2: Establish SSH connection with remote node using sshConfig
	log.Info().
		Str("host", sshConfig.NodeAddress).
		Msg("Establishing SSH connection to remote node")

	client, err := shadowssh.NewSSHClient(&sshConfig)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", sshConfig.NodeAddress).
			Msg("Failed to establish SSH connection")
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Warn().
				Err(cerr).
				Msg("Failed to close SSH connection gracefully")
		} else {
			log.Debug().
				Msg("SSH connection closed")
		}
	}()

	// Step 2: Check if the Docker image tar file exists locally
	imageTarExistsLocally, localTarPath := checkLocalImageTarExists(imageName)
	if !imageTarExistsLocally {
		// Step 3: Build the Docker image locally
		log.Info().
			Str("image", imageName).
			Str("dockerfile_path", dockerFilePath).
			Msg("Building Docker image locally")

		if err := dockercli.BuildDockerImage(dockerFilePath, imageName); err != nil {
			log.Error().
				Err(err).
				Str("image", imageName).
				Msg("Failed to build Docker image")
			return fmt.Errorf("failed to build Docker image %s: %w", imageName, err)
		}
		log.Info().
			Str("image", imageName).
			Msg("Successfully built Docker image locally")

		// Step 4: Export the Docker image to a tar file
		log.Info().
			Str("image", imageName).
			Msg("Exporting Docker image to tar")

		buildTarsDir := "./tarfiles"
		if _, err := os.Stat(buildTarsDir); os.IsNotExist(err) {
			log.Info().
				Str("directory", buildTarsDir).
				Msg("Creating tarfiles directory")

			if err := os.MkdirAll(buildTarsDir, 0755); err != nil {
				log.Error().
					Err(err).
					Str("directory", buildTarsDir).
					Msg("Failed to create tarfiles directory")
				return fmt.Errorf("failed to create tarfiles directory: %w", err)
			}
		}

		exportedTar, err := dockercli.ExportImageToTar(context.Background(), 3, imageName, localTarPath)
		if err != nil {
			log.Error().
				Err(err).
				Str("image", imageName).
				Str("tar_path", localTarPath).
				Msg("Failed to export Docker image to tar")
			return fmt.Errorf("failed to export Docker image %s to tar: %w", imageName, err)
		}
		log.Info().
			Str("image", imageName).
			Str("tar_path", exportedTar).
			Msg("Successfully exported Docker image to tar")
	} else {
		log.Info().
			Str("image", imageName).
			Str("tar_path", localTarPath).
			Msg("Image tar already exists locally. Skipping build and export.")
	}

	// Step 4: Copy the tar file onto the remote node into /tmp
	log.Info().
		Str("tar_path", localTarPath).
		Str("remote_dir", "/tmp").
		Msg("Copying tar file to remote node")

	if err := shadowscp.ShadowCopyFile(localTarPath, "/tmp"); err != nil {
		log.Error().
			Err(err).
			Str("tar_path", localTarPath).
			Str("remote_dir", "/tmp").
			Msg("Failed to copy tar file to remote node")
		return fmt.Errorf("failed to copy tar %s to remote: %w", localTarPath, err)
	}
	log.Info().
		Str("tar_path", localTarPath).
		Str("remote_dir", "/tmp").
		Msg("Successfully copied tar file to remote node")

	// Step 5: Load the Docker image on the remote node
	log.Info().
		Str("image", imageName).
		Str("remote_tar_path", "/tmp").
		Msg("Loading Docker image on remote node")

	loadCmd := fmt.Sprintf("docker load -i %s", "/tmp")
	_, err = shadowssh.ExecuteCommand(client, loadCmd)
	if err != nil {
		log.Error().
			Err(err).
			Str("image", imageName).
			Str("command", loadCmd).
			Msg("Failed to load Docker image on remote node")
		return fmt.Errorf("failed to load Docker image %s on remote node: %w", imageName, err)
	}
	log.Info().
		Str("image", imageName).
		Msg("Successfully loaded Docker image on remote node")

	// Step 6: Remove the tar file from the remote node
	log.Info().
		Str("remote_tar_path", "/tmp").
		Msg("Removing tar file from remote node")

	removeCmd := fmt.Sprintf("rm %s", "/tmp")
	_, err = shadowssh.ExecuteCommand(client, removeCmd)
	if err != nil {
		log.Warn().
			Err(err).
			Str("command", removeCmd).
			Msg("Failed to remove tar file from remote node")
		// Not returning error as removal failure is non-critical
	} else {
		log.Info().
			Str("command", removeCmd).
			Msg("Successfully removed tar file from remote node")
	}

	log.Info().
		Str("image", imageName).
		Msg("Image deployment process completed successfully")

	return nil
}

// checkLocalImageTarExists checks if the image tar file exists locally.
func checkLocalImageTarExists(imageName string) (bool, string) {
	sanitizedImageName := sanitizeImageName(imageName)
	localTarPath := filepath.Join("./tarfiles", fmt.Sprintf("%s.tar", sanitizedImageName))

	_, err := os.Stat(localTarPath)
	if err == nil {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar exists locally")
		return true, localTarPath
	}
	if os.IsNotExist(err) {
		log.Debug().
			Str("local_tar_path", localTarPath).
			Msg("Image tar does not exist locally")
		return false, localTarPath
	}
	// For other errors, log and treat as non-existent
	log.Error().
		Err(err).
		Str("local_tar_path", localTarPath).
		Msg("Error checking local tar file existence")
	return false, localTarPath
}

func CreateTestEnivronment(userConfigurtionFilePath string) {
	// Step 1 Load user configuration from yaml file using  userParser.LoadConfig(yamlFilePath string) (*UsersConfig, error)
	// Step 2 Ensure that each UserDetails.DockerImageTag is present on the remote node, otherwise throw an error
	// Step 3 Create the users through the KASM API using api.CreateUser(user TargetUser) (*UserResponse, error) and use api.GetUsers() ([]UserResponse, error) to check if user exists.
	// Step 4 Add the user to the specified group UserDetails.Role using api.AddUserToGroup(userID, groupID string) error
	// Step 6 Update the YAML file with the userId and KasmSessionOfContainer using userParser.UpdateUserConfig(yamlFilePath, username, userID, kasmSessionOfContainer string) error

}
