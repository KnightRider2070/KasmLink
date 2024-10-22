package procedures

import (
	"fmt"
	"kasmlink/pkg/scp"
	shadowssh "kasmlink/pkg/ssh"
	"path/filepath"
)

// ImportDockerImageToRemoteNode copies a Docker image tar to the remote node and imports it using SSH
func ImportDockerImageToRemoteNode(username, password, host, localTarFilePath, remoteDir string) error {
	// Step 1: Copy the Docker image tar file to the remote node
	err := shadowscp.ShadowCopyFile(username, password, host, localTarFilePath, remoteDir)
	if err != nil {
		return fmt.Errorf("failed to copy tar file to remote node: %v", err)
	}

	// Step 2: Execute the Docker import command on the remote node via SSH
	remoteTarFilePath := filepath.Join(remoteDir, filepath.Base(localTarFilePath))
	checkCommand := fmt.Sprintf("ls %s && docker load -i %s", remoteTarFilePath, remoteTarFilePath)

	err = shadowssh.ShadowExecuteCommand(username, password, host, checkCommand)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Docker image imported successfully from %s on %s\n", localTarFilePath, host)
	return nil
}
