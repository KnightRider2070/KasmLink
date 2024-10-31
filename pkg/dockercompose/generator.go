package dockercompose

import (
	"fmt"
	"os"
	"text/template"
)

// HealthCheck struct for full customization
type HealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

// ULimit defines customizable ulimit settings
type ULimit struct {
	Name string `yaml:"name"`
	Soft int    `yaml:"soft"`
	Hard int    `yaml:"hard"`
}

// ResourceLimits for defining limits and reservations
type ResourceLimits struct {
	MemoryLimit       string `yaml:"memory_limit"`
	CPULimit          string `yaml:"cpu_limit"`
	MemoryReservation string `yaml:"memory_reservation"`
}

// LoggingConfig for logging driver options
type LoggingConfig struct {
	Driver  string            `yaml:"driver"`
	MaxSize string            `yaml:"max_size"`
	MaxFile string            `yaml:"max_file"`
	Labels  map[string]string `yaml:"labels"`
}

// RestartPolicy defines service restart conditions
type RestartPolicy struct {
	RestartCondition string `yaml:"restart_condition"`
	RestartDelay     string `yaml:"restart_delay"`
	MaxAttempts      int    `yaml:"max_attempts"`
	RestartWindow    string `yaml:"restart_window"`
}

// VolumeDefinition defines a volume
type VolumeDefinition struct {
	Name     string            `yaml:"name"`
	External bool              `yaml:"external"`
	Driver   string            `yaml:"driver"`
	Labels   map[string]string `yaml:"labels"`
}

// SecretsDefinition defines a secret
type SecretsDefinition struct {
	Name     string `yaml:"name"`
	FilePath string `yaml:"file_path"`
}

// ConfigsDefinition defines a config
type ConfigsDefinition struct {
	Name     string `yaml:"name"`
	FilePath string `yaml:"file_path"`
}

// ServiceInput holds the dynamic values needed to generate a service in the Docker Compose file.
type ServiceInput struct {
	ServiceName          string            `yaml:"service_name"`
	BuildContext         string            `yaml:"build_context"`
	ContainerName        string            `yaml:"container_name"`
	ContainerIP          string            `yaml:"container_ip"`
	EnvironmentVariables map[string]string `yaml:"environment_variables"`
	HealthCheck          HealthCheck       `yaml:"health_check"`
	ULimits              []ULimit          `yaml:"ulimits"`
	Command              string            `yaml:"command"`
	Volumes              []string          `yaml:"volumes"`
	Resources            ResourceLimits    `yaml:"resources"`
	Logging              LoggingConfig     `yaml:"logging"`
	Deploy               RestartPolicy     `yaml:"deploy"`
	DependsOn            []string          `yaml:"depends_on"`
	Labels               map[string]string `yaml:"labels"`
	TTY                  bool              `yaml:"tty"`
	ShmSize              string            `yaml:"shm_size"`
}

// ComposeFile holds the overall configuration for generating the Docker Compose file.
type ComposeFile struct {
	Services          []ServiceInput      `yaml:"services"`
	VolumesDefinition []VolumeDefinition  `yaml:"volumes_definition"`
	NetworkName       string              `yaml:"network_name"`
	NetworkSubnet     string              `yaml:"network_subnet"`
	NetworkLabels     map[string]string   `yaml:"network_labels"`
	SecretsDefinition []SecretsDefinition `yaml:"secrets_definition"`
	ConfigsDefinition []ConfigsDefinition `yaml:"configs_definition"`
}

// GenerateDockerComposeFile applies the user inputs to the Docker Compose template and writes it to the specified output path.
func GenerateDockerComposeFile(tmpl *template.Template, composeFile ComposeFile, outputPath string) error {
	// Create the output file at the specified path
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Apply the user input to the template and generate the Docker Compose file
	err = tmpl.Execute(outputFile, composeFile)
	if err != nil {
		return fmt.Errorf("failed to apply template: %v", err)
	}

	fmt.Printf("docker-compose.yml has been generated successfully at %s.\n", outputPath)
	return nil
}
