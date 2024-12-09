package Tests

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"kasmlink/pkg/dockercompose"
	"testing"
)

// TestComposeFileWithFullStructure Tests if a full ComposeFile struct with nested structures marshals correctly.
func TestComposeFileWithFullStructure(t *testing.T) {
	composeFile := dockercompose.ComposeFile{
		Version: "3.8",
		Services: map[string]dockercompose.Service{
			"test_service": {
				ContainerName: "test_container",
				Image:         "test_image",
				Build: &dockercompose.BuildConfig{
					Context:    "./context",
					Dockerfile: "Dockerfile",
					Args: map[string]string{
						"ARG1": "value1",
						"ARG2": "value2",
					},
					Target: "target_stage",
					Labels: map[string]string{"label1": "value1"},
				},
				Environment:     map[interface{}]interface{}{"ENV_VAR1": "value1"},
				Ports:           []string{"80:80"},
				Volumes:         []string{"/data:/data"},
				Secrets:         []string{"secret1"},
				Configs:         []string{"config1"},
				Labels:          map[string]string{"label1": "value1"},
				Command:         []string{"echo", "hello world"},
				Entrypoint:      []string{"entrypoint.sh"},
				User:            "1001",
				GroupAdd:        []string{"1000"},
				WorkingDir:      "/app",
				RestartPolicy:   "on-failure",
				StopGracePeriod: "30s",
				DependsOn:       []string{"another_service"},
				Healthcheck: &dockercompose.Healthcheck{
					Test:        []string{"CMD-SHELL", "exit 0"},
					Interval:    "10s",
					Timeout:     "2s",
					Retries:     3,
					StartPeriod: "5s",
				},
				Logging: &dockercompose.Logging{
					Driver: "json-file",
					Options: map[string]string{
						"max-size": "10m",
						"max-file": "3",
					},
				},
				ExtraHosts: []string{"localhost:127.0.0.1"},
				DNSConfig: dockercompose.DNSConfig{
					Servers: []string{"8.8.8.8"},
					Search:  []string{"local"},
					Options: []string{"debug"},
				},
				Attach:     true,
				Privileged: true,
				Tty:        true,
				StdinOpen:  true,
				Annotations: map[string]string{
					"annotation1": "value1",
				},
				Devices: []string{"/dev/null:/dev/null"},
				Ulimits: []string{"nofile=1024:2048"},
				Init:    true,
				CPUConfig: dockercompose.CPUConfig{
					Count:     "2",
					Percent:   "50",
					Shares:    "1024",
					Period:    "100000",
					Quota:     "200000",
					RTPeriod:  "1000000",
					RTRuntime: "950000",
					Set:       "0,1",
				},
				MemoryConfig: dockercompose.MemoryConfig{
					Limit:       "512m",
					Reservation: "256m",
					SwapLimit:   "1024m",
					Swappiness:  "60",
				},
				Capabilities: dockercompose.Capabilities{
					Add:  []string{"NET_ADMIN"},
					Drop: []string{"MKNOD"},
				},
				NetworkConfig: dockercompose.NetworkConfig{
					Mode:       "bridge",
					Networks:   []interface{}{"net1"},
					Links:      []string{"service_a:alias"},
					MacAddress: "02:42:ac:11:00:02",
				},
			},
		},
		Networks: map[string]dockercompose.Network{
			"example_network": {
				Driver: "bridge",
				DriverOpts: map[string]string{
					"com.docker.network.bridge.default_bridge": "false",
				},
				External: false,
			},
		},
		Volumes: map[string]dockercompose.Volume{
			"example_volume": {
				Driver: "local",
				DriverOpts: map[string]string{
					"type":   "none",
					"device": "/path/to/host/dir",
					"o":      "bind",
				},
				External: false,
			},
		},
		Configs: map[string]dockercompose.Config{
			"example_config": {
				File:     "./config.yml",
				External: false,
			},
		},
		Secrets: map[string]dockercompose.Secret{
			"example_secret": {
				File:     "./secret.txt",
				External: true,
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(composeFile)
	assert.NoError(t, err)

	// Unmarshal back to ComposeFile struct
	var unmarshaledComposeFile dockercompose.ComposeFile
	err = yaml.Unmarshal(data, &unmarshaledComposeFile)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "version: \"3.8\"")

	// Assert that the unmarshaled data matches the original
	assert.Equal(t, composeFile, unmarshaledComposeFile)
}

// TestComposeFileWithPartialStructure Tests if a ComposeFile struct with partial structures marshals and unmarshals correctly.
func TestComposeFileWithPartialStructure(t *testing.T) {
	composeFile := dockercompose.ComposeFile{
		Version: "3.8",
		Services: map[string]dockercompose.Service{
			"partial_service": {
				ContainerName: "partial_container",
				Image:         "nginx",
				Ports:         []string{"8080:80"},
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(composeFile)
	assert.NoError(t, err)

	// Unmarshal back to ComposeFile struct
	var unmarshaledComposeFile dockercompose.ComposeFile
	err = yaml.Unmarshal(data, &unmarshaledComposeFile)
	assert.NoError(t, err)

	// Assert that the unmarshaled data matches the original
	assert.Equal(t, composeFile, unmarshaledComposeFile)
}
