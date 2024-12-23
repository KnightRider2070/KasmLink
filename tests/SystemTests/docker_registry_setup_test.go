package SystemTests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"kasmlink/pkg/dockerRegistry"
	sshmanager "kasmlink/pkg/shadowssh"
	"net/http"
	"testing"
	"time"
)

func TestDockerRegistrySetup(t *testing.T) {

	var dockerRegistryConfig *dockerRegistry.RegistryConfig
	dockerRegistryConfig = dockerRegistry.NewRegistryConfig()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()

	sshConfig, err := sshmanager.NewSSHConfig(user, password, hostIp, 22, knwHosts, 10*time.Second)
	assert.NoError(t, err)

	// Run SetupRegistry
	err = dockerRegistry.SetupRegistry(ctx, sshConfig, dockerRegistryConfig, "./")
	assert.NoError(t, err)

	// Validate the registry is up by pinging its URL
	registryURL := fmt.Sprintf("http://%s:%d/v2/", hostIp, dockerRegistryConfig.Port)
	resp, err := http.Get(registryURL)

	// Assert HTTP request success
	assert.NoError(t, err, "Failed to ping Docker registry")
	if resp != nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Docker registry did not respond with HTTP 200 OK")
	}
}
