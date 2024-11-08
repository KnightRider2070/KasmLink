package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/ssh"
	"path/filepath"
	"time"
)

const (
	retryDelay = 2 * time.Second
	retryCount = 3
)

// ImportDockerImageToRemoteNode copies a Docker image tar to the remote node and imports it using SSH.
func ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir string) error {
	log.Info().
		Str("username", username).
		Str("host", host).
		Str("local_tar_file", localTarFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting Docker image import to remote node")

	// Step 1: Retry mechanism to copy the Docker image tar file to the remote node.
	if err := retryOperation(retryCount, retryDelay, func() error {
		return shadowscp.ShadowCopyFile(localTarFilePath, remoteDir)
	}, "copy tar file to remote node"); err != nil {
		return err
	}

	log.Info().
		Str("local_tar_file", localTarFilePath).
		Str("host", host).
		Msg("Docker image tar file copied to remote node successfully")

	// Step 2: Establish SSH connection to target node
	sshConfig := shadowssh.NewSSHConfigFromFlags()

	client, err := shadowssh.NewSSHClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %v", err)
	}
	defer client.Close()

	// Step 3: Execute the Docker import command on the remote node via SSH with retry mechanism.
	remoteTarFilePath := filepath.Join(remoteDir, filepath.Base(localTarFilePath))
	checkCommand := fmt.Sprintf("ls %s && docker load -i %s", remoteTarFilePath, remoteTarFilePath)

	if err := retryOperation(retryCount, retryDelay, func() error {
		_, execErr := shadowssh.ExecuteCommand(client, checkCommand)
		return execErr
	}, "execute Docker load command on remote node"); err != nil {
		return err
	}

	// Step 4: Remove the imported tar file
	deleteCommand := fmt.Sprintf("rm -rf %s", remoteTarFilePath)

	if err := retryOperation(retryCount, retryDelay, func() error {
		_, execErr := shadowssh.ExecuteCommand(client, deleteCommand)
		return execErr
	}, "remove tar file from remote node"); err != nil {
		return err
	}

	log.Info().
		Str("local_tar_file", localTarFilePath).
		Str("host", host).
		Msg("Docker image imported successfully on remote node")
	fmt.Printf("Docker image imported successfully from %s on %s\n", localTarFilePath, host)

	return nil
}

// retryOperation provides a reusable retry mechanism for repeated operations.
func retryOperation(retries int, delay time.Duration, operation func() error, description string) error {
	for retries > 0 {
		err := operation()
		if err != nil {
			retries--
			log.Error().
				Err(err).
				Int("retries_left", retries).
				Msg(fmt.Sprintf("Failed to %s, retrying", description))
			time.Sleep(delay)
			if retries == 0 {
				return fmt.Errorf("failed to %s after retries: %v", description, err)
			}
		} else {
			break
		}
	}
	return nil
}
