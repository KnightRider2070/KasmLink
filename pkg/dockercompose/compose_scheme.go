package dockercompose

// ComposeFile represents the structure of a Docker Compose file.
type ComposeFile struct {
	Version  string             `yaml:"version,omitempty"`  // Optional: specifies the version of the Compose file format
	Services map[string]Service `yaml:"services"`           // Required: service configurations
	Networks map[string]Network `yaml:"networks,omitempty"` // Optional: network configurations
	Volumes  map[string]Volume  `yaml:"volumes,omitempty"`  // Optional: volume configurations
	Configs  map[string]Config  `yaml:"configs,omitempty"`  // Optional: config configurations
	Secrets  map[string]Secret  `yaml:"secrets,omitempty"`  // Optional: secret configurations
}

// Service represents an individual service configuration within the Compose file.
type Service struct {
	ContainerName   string            `yaml:"container_name,omitempty"`    // Optional: name of the container
	Image           string            `yaml:"image,omitempty"`             // Optional: Docker image to use
	Build           *BuildConfig      `yaml:"build,omitempty"`             // Optional: build context or options
	Environment     map[string]string `yaml:"environment,omitempty"`       // Optional: environment variables
	Ports           []string          `yaml:"ports,omitempty"`             // Optional: port mappings
	Volumes         []string          `yaml:"volumes,omitempty"`           // Optional: volume mounts
	Secrets         []string          `yaml:"secrets,omitempty"`           // Optional: secret references
	Configs         []string          `yaml:"configs,omitempty"`           // Optional: config references
	Labels          map[string]string `yaml:"labels,omitempty"`            // Optional: service labels
	Command         []string          `yaml:"command,omitempty"`           // Optional: command to execute
	Entrypoint      []string          `yaml:"entrypoint,omitempty"`        // Optional: entrypoint for the container
	User            string            `yaml:"user,omitempty"`              // Optional: user for the container
	GroupAdd        []string          `yaml:"group_add,omitempty"`         // Optional: additional groups
	WorkingDir      string            `yaml:"working_dir,omitempty"`       // Optional: working directory inside the container
	RestartPolicy   string            `yaml:"restart,omitempty"`           // Optional: restart policy
	StopGracePeriod string            `yaml:"stop_grace_period,omitempty"` // Optional: stop grace period
	DependsOn       []string          `yaml:"depends_on,omitempty"`        // Optional: service dependencies
	Healthcheck     *Healthcheck      `yaml:"healthcheck,omitempty"`       // Optional: health check configuration
	Logging         *Logging          `yaml:"logging,omitempty"`           // Optional: logging configuration
	ExtraHosts      []string          `yaml:"extra_hosts,omitempty"`       // Optional: additional hostnames
	DNSConfig       DNSConfig         `yaml:",inline"`                     // Optional: DNS configuration
	Attach          bool              `yaml:"attach,omitempty"`            // Optional: collect service logs
	Privileged      bool              `yaml:"privileged,omitempty"`        // Optional: privileged mode
	Tty             bool              `yaml:"tty,omitempty"`               // Optional: allocate a TTY
	StdinOpen       bool              `yaml:"stdin_open,omitempty"`        // Optional: keep STDIN open
	Annotations     map[string]string `yaml:"annotations,omitempty"`       // Optional: custom annotations
	Devices         []string          `yaml:"devices,omitempty"`           // Optional: list of devices
	Ulimits         []string          `yaml:"ulimits,omitempty"`           // Optional: ulimit options
	Init            bool              `yaml:"init,omitempty"`              // Optional: run init within the container

	// Inline embedded configurations
	CPUConfig     CPUConfig     `yaml:",inline"` // Inline: CPU-related settings for the service
	MemoryConfig  MemoryConfig  `yaml:",inline"` // Inline: Memory-related settings for the service
	Capabilities  Capabilities  `yaml:",inline"` // Inline: Capabilities for the service
	NetworkConfig NetworkConfig `yaml:",inline"` // Inline: Network configuration for the service
}

// BuildConfig represents the build configuration for a service.
type BuildConfig struct {
	Context    string            `yaml:"context,omitempty"`    // Build context (path)
	Dockerfile string            `yaml:"dockerfile,omitempty"` // Dockerfile to use
	Args       map[string]string `yaml:"args,omitempty"`       // Build arguments
	Target     string            `yaml:"target,omitempty"`     // Build target stage
	Labels     map[string]string `yaml:"labels,omitempty"`     // Build labels
}

// CPUConfig holds CPU-related settings for a service.
type CPUConfig struct {
	Count     string `yaml:"cpu_count,omitempty"`      // Optional: number of CPUs
	Percent   string `yaml:"cpu_percent,omitempty"`    // Optional: CPU percentage
	Shares    string `yaml:"cpu_shares,omitempty"`     // Optional: CPU shares
	Period    string `yaml:"cpu_period,omitempty"`     // Optional: CPU period
	Quota     string `yaml:"cpu_quota,omitempty"`      // Optional: CPU quota
	RTPeriod  string `yaml:"cpu_rt_period,omitempty"`  // Optional: Real-time period
	RTRuntime string `yaml:"cpu_rt_runtime,omitempty"` // Optional: Real-time runtime
	Set       string `yaml:"cpuset,omitempty"`         // Optional: CPUs allowed for execution
}

// MemoryConfig holds memory-related settings for a service.
type MemoryConfig struct {
	Limit       string `yaml:"mem_limit,omitempty"`       // Optional: memory limit
	Reservation string `yaml:"mem_reservation,omitempty"` // Optional: memory reservation
	SwapLimit   string `yaml:"memswap_limit,omitempty"`   // Optional: memory swap limit
	Swappiness  string `yaml:"mem_swappiness,omitempty"`  // Optional: memory swappiness
}

// Capabilities holds capability-related settings for a service.
type Capabilities struct {
	Add  []string `yaml:"cap_add,omitempty"`  // Optional: capabilities to add
	Drop []string `yaml:"cap_drop,omitempty"` // Optional: capabilities to drop
}

// NetworkConfig holds network-related settings for a service.
type NetworkConfig struct {
	Mode       string   `yaml:"network_mode,omitempty"` // Optional: network mode
	Networks   []string `yaml:"networks,omitempty"`     // Optional: list of networks
	Links      []string `yaml:"links,omitempty"`        // Optional: links to other services
	MacAddress string   `yaml:"mac_address,omitempty"`  // Optional: MAC address
}

// Healthcheck holds health check settings for a service.
type Healthcheck struct {
	Test        []string `yaml:"test,omitempty"`         // Health check command
	Interval    string   `yaml:"interval,omitempty"`     // Interval between health checks
	Timeout     string   `yaml:"timeout,omitempty"`      // Timeout for health check
	Retries     int      `yaml:"retries,omitempty"`      // Number of retries before failure
	StartPeriod string   `yaml:"start_period,omitempty"` // Start period before health check
}

// DNSConfig holds DNS-related settings for a service.
type DNSConfig struct {
	Servers []string `yaml:"dns,omitempty"`        // Optional: list of DNS servers
	Search  []string `yaml:"dns_search,omitempty"` // Optional: DNS search domains
	Options []string `yaml:"dns_opt,omitempty"`    // Optional: DNS options
}

// Logging represents the logging configuration for a service.
type Logging struct {
	Driver  string            `yaml:"driver,omitempty"`  // Logging driver (e.g., "json-file", "syslog")
	Options map[string]string `yaml:"options,omitempty"` // Options for the logging driver
}

// Network represents the network configuration in a Compose file.
type Network struct {
	Driver     string            `yaml:"driver,omitempty"`      // Network driver
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"` // Driver options
	External   bool              `yaml:"external,omitempty"`    // External network flag
}

// Volume represents the volume configuration in a Compose file.
type Volume struct {
	Driver     string            `yaml:"driver,omitempty"`      // Volume driver
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"` // Driver options
	External   bool              `yaml:"external,omitempty"`    // External volume flag
}

// Config represents a config entry in a Compose file.
type Config struct {
	File     string `yaml:"file,omitempty"`     // Path to the config file
	External bool   `yaml:"external,omitempty"` // External config flag
}

// Secret represents a secret entry in a Compose file.
type Secret struct {
	File     string `yaml:"file,omitempty"`     // Path to the secret file
	External bool   `yaml:"external,omitempty"` // External secret flag
}
