package procedures

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/ssh"
	"path/filepath"
)

// ImportDockerImageToRemoteNode copies a Docker image tar to the remote node and imports it using SSH.
func ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir string) error {
	log.Info().
		Str("username", username).
		Str("host", host).
		Str("local_tar_file", localTarFilePath).
		Str("remote_dir", remoteDir).
		Msg("Starting Docker image import to remote node")

	// Step 1: Copy the Docker image tar file to the remote node.
	err := shadowscp.ShadowCopyFile(username, password, host, localTarFilePath, remoteDir)
	if err != nil {
		log.Error().
			Err(err).
			Str("local_tar_file", localTarFilePath).
			Str("host", host).
			Msg("Failed to copy tar file to remote node")
		return fmt.Errorf("failed to copy tar file to remote node: %v", err)
	}
	log.Info().
		Str("local_tar_file", localTarFilePath).
		Str("host", host).
		Msg("Docker image tar file copied to remote node successfully")

	// Step 2: Execute the Docker import command on the remote node via SSH.
	remoteTarFilePath := filepath.Join(remoteDir, filepath.Base(localTarFilePath))
	checkCommand := fmt.Sprintf("ls %s && docker load -i %s", remoteTarFilePath, remoteTarFilePath)

	log.Info().
		Str("host", host).
		Str("command", checkCommand).
		Msg("Executing Docker load command on remote node")

	err = shadowssh.ShadowExecuteCommand(username, password, host, checkCommand)
	if err != nil {
		log.Error().
			Err(err).
			Str("host", host).
			Str("command", checkCommand).
			Msg("Failed to execute Docker load command on remote node")
		return fmt.Errorf("failed to execute Docker load command on remote node: %v", err)
	}

	log.Info().
		Str("local_tar_file", localTarFilePath).
		Str("host", host).
		Msg("Docker image imported successfully on remote node")
	fmt.Printf("Docker image imported successfully from %s on %s\n", localTarFilePath, host)
	return nil
}
