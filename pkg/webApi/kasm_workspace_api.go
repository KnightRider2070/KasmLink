package webApi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

//NOTE: Using undocumented API endpoints. This might require changes for new versions of Kasm.
// Reference for further information https://kasmweb.atlassian.net/servicedesk/customer/portal/3/article/10682377

// TargetImage represents the structure for the "target_image" object used
// in create, update, and other image-related requests.
type TargetImage struct {
	AllowNetworkSelection  bool            `json:"allow_network_selection,omitempty"`
	Categories             string          `json:"categories,omitempty"`
	Cores                  float64         `json:"cores"`
	CPUAllocationMethod    string          `json:"cpu_allocation_method"`
	Description            string          `json:"description"`
	DockerRegistry         string          `json:"docker_registry,omitempty"`
	DockerToken            string          `json:"docker_token,omitempty"`
	DockerUser             string          `json:"docker_user,omitempty"`
	Enabled                bool            `json:"enabled"`
	ExecConfig             string          `json:"exec_config,omitempty"`
	FilterPolicyID         *string         `json:"filter_policy_id,omitempty"`
	FriendlyName           string          `json:"friendly_name"`
	GPUCount               float64         `json:"gpu_count"`
	Hash                   string          `json:"hash,omitempty"`
	Hidden                 bool            `json:"hidden,omitempty"`
	ImageID                string          `json:"image_id,omitempty"`
	ImageSrc               *string         `json:"image_src,omitempty"`
	ImageType              string          `json:"image_type"`
	IsRemoteApp            bool            `json:"is_remote_app,omitempty"`
	LaunchConfig           json.RawMessage `json:"launch_config,omitempty"`
	LinkURL                *string         `json:"link_url,omitempty"`
	Memory                 int             `json:"memory"`
	Name                   string          `json:"name"`
	Notes                  string          `json:"notes,omitempty"`
	OverrideEgressGateways bool            `json:"override_egress_gateways,omitempty"`
	PersistentProfilePath  *string         `json:"persistent_profile_path,omitempty"`
	RDPClientType          *string         `json:"rdp_client_type,omitempty"`
	RemoteAppArgs          *string         `json:"remote_app_args,omitempty"`
	RemoteAppName          *string         `json:"remote_app_name,omitempty"`
	RemoteAppProgram       *string         `json:"remote_app_program,omitempty"`
	RequireGPU             bool            `json:"require_gpu,omitempty"`
	RestrictNetworkNames   []string        `json:"restrict_network_names,omitempty"`
	RestrictToNetwork      bool            `json:"restrict_to_network,omitempty"`
	RestrictToServer       bool            `json:"restrict_to_server,omitempty"`
	RestrictToZone         bool            `json:"restrict_to_zone,omitempty"`
	RunConfig              string          `json:"run_config,omitempty"`
	ServerID               string          `json:"server_id,omitempty"`
	ServerPoolID           *string         `json:"server_pool_id,omitempty"`
	SessionTimeLimit       string          `json:"session_time_limit,omitempty"`
	UncompressedSizeMB     int             `json:"uncompressed_size_mb,omitempty"`
	VolumeMappings         string          `json:"volume_mappings,omitempty"`
	ZoneID                 string          `json:"zone_id,omitempty"`
}

// CreateImageRequest represents the request structure for creating/updating an image.
// Now includes APIKey and APIKeySecret.
type CreateImageRequest struct {
	APIKey       string      `json:"api_key"`
	APIKeySecret string      `json:"api_key_secret"`
	TargetImage  TargetImage `json:"target_image"`
}

// ImageDetail represents the structure of the "image" object in the response.
type ImageDetail struct {
	ImageID                   string                 `json:"image_id"`
	Cores                     float64                `json:"cores"`
	Description               string                 `json:"description"`
	DockerRegistry            *string                `json:"docker_registry,omitempty"`
	DockerToken               *string                `json:"docker_token,omitempty"`
	DockerUser                *string                `json:"docker_user,omitempty"`
	Enabled                   bool                   `json:"enabled"`
	FriendlyName              string                 `json:"friendly_name"`
	Hash                      *string                `json:"hash,omitempty"`
	Memory                    int                    `json:"memory"`
	Name                      string                 `json:"name"`
	XRes                      int                    `json:"x_res"`
	YRes                      int                    `json:"y_res"`
	ImageAttributes           []string               `json:"imageAttributes,omitempty"`
	RestrictToNetwork         bool                   `json:"restrict_to_network"`
	RestrictNetworkNames      []string               `json:"restrict_network_names,omitempty"`
	RestrictToServer          bool                   `json:"restrict_to_server"`
	ServerID                  *string                `json:"server_id,omitempty"`
	PersistentProfileConfig   map[string]interface{} `json:"persistent_profile_config,omitempty"`
	VolumeMappings            json.RawMessage        `json:"volume_mappings,omitempty"`
	RunConfig                 DockerRunConfig        `json:"run_config,omitempty"`
	ImageSrc                  *string                `json:"image_src,omitempty"`
	Available                 bool                   `json:"available"`
	ExecConfig                map[string]interface{} `json:"exec_config,omitempty"`
	RestrictToZone            bool                   `json:"restrict_to_zone"`
	ZoneID                    *string                `json:"zone_id,omitempty"`
	ZoneName                  *string                `json:"zone_name,omitempty"`
	PersistentProfilePath     *string                `json:"persistent_profile_path,omitempty"`
	FilterPolicyID            *string                `json:"filter_policy_id,omitempty"`
	FilterPolicyName          *string                `json:"filter_policy_name,omitempty"`
	FilterPolicyForceDisabled bool                   `json:"filter_policy_force_disabled"`
	SessionTimeLimit          *string                `json:"session_time_limit,omitempty"`
	Categories                []string               `json:"categories,omitempty"`
	DefaultCategory           string                 `json:"default_category,omitempty"`
	AllowNetworkSelection     bool                   `json:"allow_network_selection"`
	RequireGPU                bool                   `json:"require_gpu"`
	GPUCount                  float64                `json:"gpu_count"`
	Hidden                    bool                   `json:"hidden"`
	Notes                     *string                `json:"notes,omitempty"`
	ImageType                 string                 `json:"image_type"`
	ServerPoolID              *string                `json:"server_pool_id,omitempty"`
	LinkURL                   *string                `json:"link_url,omitempty"`
	CPUAllocationMethod       string                 `json:"cpu_allocation_method"`
	UncompressedSizeMB        int                    `json:"uncompressed_size_mb,omitempty"`
	LaunchConfig              map[string]interface{} `json:"launch_config,omitempty"`
	RDPClientType             *string                `json:"rdp_client_type,omitempty"`
	OverrideEgressGateways    bool                   `json:"override_egress_gateways"`
	RemoteAppName             *string                `json:"remote_app_name,omitempty"`
	RemoteAppArgs             *string                `json:"remote_app_args,omitempty"`
	RemoteAppIcon             *string                `json:"remote_app_icon,omitempty"`
	RemoteAppProgram          *string                `json:"remote_app_program,omitempty"`
	IsRemoteApp               bool                   `json:"is_remote_app"`
}

// DockerRunConfig represents the Docker Run Config Override structure
// Naming and descriptions are from https://docker-py.readthedocs.io/en/stable/index.html used by Kasm.
type DockerRunConfig struct {
	// Basic config
	Image      string            `json:"image,omitempty"`       // The image to run.
	Command    []string          `json:"command,omitempty"`     // The command to run in the container.
	Name       string            `json:"name,omitempty"`        // The name for this container.
	Entrypoint []string          `json:"entrypoint,omitempty"`  // The entrypoint for the container.
	WorkingDir string            `json:"working_dir,omitempty"` // Path to the working directory.
	User       string            `json:"user,omitempty"`        // Username or UID to run commands as inside the container.
	Hostname   string            `json:"hostname,omitempty"`    // Additional hostnames to resolve inside the container, as a mapping of hostname to IP address.
	Domainname string            `json:"domainname,omitempty"`  // Set custom DNS search domains.
	Platform   string            `json:"platform,omitempty"`    // Platform in the format os[/arch[/variant]]. Only used if the method needs to pull the requested image.
	Labels     map[string]string `json:"labels,omitempty"`      // A dictionary of name-value labels (e.g. {"label1": "value1", "label2": "value2"}) or a list of names of labels to set with empty values (e.g. ["label1", "label2"])

	// Networking
	Network         string            `json:"network,omitempty"`          // Name of the network this container will be connected to at creation time. You can connect to additional networks using Network.connect(). Incompatible with network_mode.
	NetworkMode     string            `json:"network_mode,omitempty"`     // One of: bridge Create a new network stack for the container on the bridge network. none No networking for this container.	container:<name|id> Reuse another container’s network stack.	host Use the host network stack. This mode is incompatible with ports.	Incompatible with network.
	DNS             []string          `json:"dns,omitempty"`              // Set custom DNS servers.
	DNSOpt          []string          `json:"dns_opt,omitempty"`          // Additional options to be added to the container’s resolv.conf file.
	DNSSearch       []string          `json:"dns_search,omitempty"`       // DNS search domains.
	ExtraHosts      map[string]string `json:"extra_hosts,omitempty"`      // Additional hostnames to resolve inside the container, as a mapping of hostname to IP address.
	NetworkDisabled bool              `json:"network_disabled,omitempty"` // Disable networking.

	// Resources & Limits
	CPUShares      int    `json:"cpu_shares,omitempty"`      // CPU shares (relative weight).
	CPUPeriod      int    `json:"cpu_period,omitempty"`      // The length of a CPU period in microseconds.
	CPUQuota       int    `json:"cpu_quota,omitempty"`       // Microseconds of CPU time that the container can get in a CPU period.
	CPURtPeriod    int    `json:"cpu_rt_period,omitempty"`   // Limit CPU real-time period in microseconds.
	CPURtRuntime   int    `json:"cpu_rt_runtime,omitempty"`  // Limit CPU real-time runtime in microseconds.
	CPUSetCpus     string `json:"cpuset_cpus,omitempty"`     // CPUs in which to allow execution (0-3, 0,1).
	CPUSetMems     string `json:"cpuset_mems,omitempty"`     // Memory nodes (MEMs) in which to allow execution (0-3, 0,1). Only effective on NUMA systems.
	CPUCount       int    `json:"cpu_count,omitempty"`       // Number of usable CPUs (Windows only).
	CPUPercent     int    `json:"cpu_percent,omitempty"`     // Usable percentage of the available CPUs (Windows only).
	MemLimit       string `json:"mem_limit,omitempty"`       // Memory limit. Accepts float values (which represent the memory limit of the created container in bytes) or a string with a units identification char (100000b, 1000k, 128m, 1g). If a string is specified without a units character, bytes are assumed as an intended unit.
	MemReservation string `json:"mem_reservation,omitempty"` // Memory soft limit.
	MemSwappiness  int    `json:"mem_swappiness,omitempty"`  // Tune a container’s memory swappiness behavior. Accepts number between 0 and 100.
	MemswapLimit   string `json:"memswap_limit,omitempty"`   // Maximum amount of memory + swap a container is allowed to consume.
	PidsLimit      int    `json:"pids_limit,omitempty"`      //Tune a container’s pids limit. Set -1 for unlimited.

	// Mounts & Volumes
	Volumes      map[string]VolumeMapping `json:"volumes,omitempty"`       // A dictionary to configure volumes mounted inside the container.
	VolumesFrom  []string                 `json:"volumes_from,omitempty"`  // List of container names or IDs to get volumes from.
	VolumeDriver string                   `json:"volume_driver,omitempty"` // The name of a volume driver/plugin.
	Mounts       []MountConfig            `json:"mounts,omitempty"`        // Specification for mounts to be added to the container. More powerful alternative to volumes. Each item in the list is expected to be a docker.types.Mount object.
	Tmpfs        map[string]string        `json:"tmpfs,omitempty"`         // Temporary filesystems to mount, as a dictionary mapping a path inside the container to options for that path.

	// Devices & Caps
	Devices           []string        `json:"devices,omitempty"`             // Expose host devices to the container, as a list of strings in the form <path_on_host>:<path_in_container>:<cgroup_permissions>.	For example, /dev/sda:/dev/xvda:rwm allows the container to have read-write access to the host’s /dev/sda via a node named /dev/xvda inside the container.
	DeviceRequests    []DeviceRequest `json:"device_requests,omitempty"`     // Expose host resources such as GPUs to the container, as a list of docker.types.DeviceRequest instances.
	DeviceCgroupRules []string        `json:"device_cgroup_rules,omitempty"` // A list of cgroup rules to apply to the container.
	CapAdd            []string        `json:"cap_add,omitempty"`             // Add kernel capabilities. For example, ["SYS_ADMIN", "MKNOD"].
	CapDrop           []string        `json:"cap_drop,omitempty"`            // Drop kernel capabilities.

	// Security
	SecurityOpt []string          `json:"security_opt,omitempty"` // A list of string values to customize labels for MLS systems, such as SELinux.
	Privileged  bool              `json:"privileged,omitempty"`   // Give extended privileges to this container.
	UsernsMode  string            `json:"userns_mode,omitempty"`  // Sets the user namespace mode for the container when user namespace remapping option is enabled. Supported values are: host
	IpcMode     string            `json:"ipc_mode,omitempty"`     // Set the IPC mode for the container.
	PidMode     string            `json:"pid_mode,omitempty"`     // If set to host, use the host PID namespace inside the container.
	UtsMode     string            `json:"uts_mode,omitempty"`     // Sets the UTS namespace mode for the container. Supported values are: host
	Isolation   string            `json:"isolation,omitempty"`    // Isolation technology to use. Default: None.
	ShmSize     string            `json:"shm_size,omitempty"`     // Size of /dev/shm (e.g. 1G).
	Sysctls     map[string]string `json:"sysctls,omitempty"`      // Kernel parameters to set in the container.
	GroupAdd    []string          `json:"group_add,omitempty"`    // List of additional group names and/or IDs that the container process will run as.

	// Environment
	Environment map[string]string `json:"environment,omitempty"` // Environment variables to set inside the container, as a dictionary or a list of strings in the format ["SOMEVARIABLE=xxx"].

	// Healthcheck
	Healthcheck *HealthcheckConfig `json:"healthcheck,omitempty"` // Specify a test to perform to check that the container is healthy.

	// Other runtime configs
	CgroupParent    string                 `json:"cgroup_parent,omitempty"`     // Override the default parent cgroup.
	Cgroupns        string                 `json:"cgroupns,omitempty"`          // Override the default cgroup namespace mode for the container. Supported values are: private, host
	AutoRemove      bool                   `json:"auto_remove,omitempty"`       // enable auto-removal of the container on daemon side when the container’s process exits
	Remove          bool                   `json:"remove,omitempty"`            // If set, remove container when done default: false
	Detach          bool                   `json:"detach,omitempty"`            // Run container in the background and return a Container object.
	StdinOpen       bool                   `json:"stdin_open,omitempty"`        // Keep STDIN open even if not attached.
	Tty             bool                   `json:"tty,omitempty"`               // Allocate a pseudo-TTY.
	Stdout          bool                   `json:"stdout,omitempty"`            // Return logs from STDOUT when detach=False. Default: True.
	Stderr          bool                   `json:"stderr,omitempty"`            // Return logs from STDERR when detach=False. Default: False.
	Stream          bool                   `json:"stream,omitempty"`            // If true and detach is false, return a log generator instead of a string. Ignored if detach is true. Default: False.
	PublishAllPorts bool                   `json:"publish_all_ports,omitempty"` // Publish all ports to the host.
	Ports           map[string]interface{} `json:"ports,omitempty"`             // Ports to bind inside the container.	The keys of the dictionary are the ports to bind inside the container, either as an integer or a string in the form port/protocol, where the protocol is either tcp, udp, or sctp.
	RestartPolicy   *RestartPolicy         `json:"restart_policy,omitempty"`    // Restart the container when it exits.
	Runtime         string                 `json:"runtime,omitempty"`           // Runtime to use with this container.
	StorageOpt      map[string]string      `json:"storage_opt,omitempty"`       // Storage driver options per container as a key-value mapping.
	Ulimits         []UlimitConfig         `json:"ulimits,omitempty"`           // Ulimits to set inside the container, as a list of docker.types.Ulimit instances.
	Init            bool                   `json:"init,omitempty"`              // Run an init inside the container that forwards signals and reaps processes
	UseConfigProxy  bool                   `json:"use_config_proxy,omitempty"`  // If True, and if the docker client configuration file (~/.docker/config.json by default) contains a proxy configuration, the corresponding environment variables will be set in the container being built.
}

type DeviceRequest struct {
	Driver       string            `json:"Driver,omitempty"`
	Count        int               `json:"Count,omitempty"`
	DeviceIDs    []string          `json:"DeviceIDs,omitempty"`
	Capabilities [][]string        `json:"Capabilities,omitempty"`
	Options      map[string]string `json:"Options,omitempty"`
}

type HealthcheckConfig struct {
	Test     []string `json:"test,omitempty"`     // Command to run to check health
	Interval int64    `json:"interval,omitempty"` // in nanoseconds
	Timeout  int64    `json:"timeout,omitempty"`  // in nanoseconds
	Retries  int      `json:"retries,omitempty"`  // The number of consecutive failures needed to consider a container as unhealthy.

	StartPeriod int64 `json:"start_period,omitempty"` // in nanoseconds
}

type MountConfig struct {
	Type        string `json:"type,omitempty"`        // The mount type (bind / volume / tmpfs / npipe). Default: volume.
	Source      string `json:"source,omitempty"`      // Mount source (e.g. a volume name or a host path).
	Target      string `json:"target,omitempty"`      // Container path.
	ReadOnly    bool   `json:"read_only,omitempty"`   //  Whether the mount should be read-only.
	Propagation string `json:"propagation,omitempty"` // A propagation mode with the value [r]private, [r]shared, or [r]slave. Only valid for the bind type.
	NoCopy      bool   `json:"no_copy,omitempty"`     // False if the volume should be populated with the data from the target. Default: False. Only valid for the volume type.
}

type RestartPolicy struct {
	Condition         string `json:"condition,omitempty"`    // Condition for restart (none, on-failure, or any). Default: none.
	Delay             int    `json:"delay,omitempty"`        // Delay between restart attempts.
	MaximumRetryCount int    `json:"max_attempts,omitempty"` // Maximum attempts to restart a given container before giving up. Default value is 0, which is ignored.
	Window            int    `json:"window,omitempty"`       // Time window used to evaluate the restart policy. Default value is 0, which is unbounded.
}

type UlimitConfig struct {
	Name string `json:"name,omitempty"` // Which ulimit will this apply to. The valid names can be found in ‘/etc/security/limits.conf’ on a gnu/linux system.
	Soft int64  `json:"soft,omitempty"` // The soft limit for the ulimit.
	Hard int64  `json:"hard,omitempty"` // The hard limit for the ulimit.
}

type VolumeMapping struct {
	Bind string `json:"bind,omitempty"` // The path to mount the volume inside the container
	Mode string `json:"mode,omitempty"` // Either rw to mount the volume read/write, or ro to mount it read-only.
	Gid  int    `json:"gid,omitempty"`  // Default GID
	Uid  int    `json:"uid,omitempty"`  // Default UID
}

// Response represents the full response structure containing the image details.
type Response struct {
	Image ImageDetail `json:"image"`
}

// DeleteImageRequest represents the payload required to delete an image.
// Now includes APIKey and APIKeySecret.
type DeleteImageRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
	TargetImage  struct {
		ImageID string `json:"image_id"`
	} `json:"target_image"`
}

// CreateImage sends a POST request to /api/public/create_image with the given CreateImageRequest payload.
// On success, it returns the parsed Response object.
func (api *KasmAPI) CreateImage(ctx context.Context, req CreateImageRequest) (*Response, error) {
	endpoint := "/api/public/create_image"

	// Populate API credentials
	req.APIKey = api.APIKey
	req.APIKeySecret = api.APIKeySecret

	respBody, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create image at %s: %w", endpoint, err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().
			Err(err).
			Str("endpoint", endpoint).
			RawJSON("response_body", respBody).
			Msg("Failed to unmarshal create_image response")
		return nil, fmt.Errorf("failed to unmarshal create_image response: %w", err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", response.Image.ImageID).
		Msg("Image created successfully")

	return &response, nil
}

// UpdateImage sends a POST request to /api/public/update_image to update an existing image.
// req.TargetImage.ImageID must be set. On success, it returns the parsed Response object.
func (api *KasmAPI) UpdateImage(ctx context.Context, req CreateImageRequest) (*Response, error) {
	endpoint := "/api/public/update_image"

	// Populate API credentials
	req.APIKey = api.APIKey
	req.APIKeySecret = api.APIKeySecret

	if req.TargetImage.ImageID == "" {
		return nil, fmt.Errorf("image_id must be set in TargetImage before calling UpdateImage")
	}

	respBody, err := api.MakePostRequest(ctx, endpoint, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update image at %s: %w", endpoint, err)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Error().
			Err(err).
			Str("endpoint", endpoint).
			Str("image_id", req.TargetImage.ImageID).
			RawJSON("response_body", respBody).
			Msg("Failed to unmarshal update_image response")
		return nil, fmt.Errorf("failed to unmarshal update_image response: %w", err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", response.Image.ImageID).
		Msg("Image updated successfully")

	return &response, nil
}

// DeleteImage sends a POST request to /api/public/delete_image to remove an existing image.
// imageID must be provided and must be a valid image ID.
func (api *KasmAPI) DeleteImage(ctx context.Context, imageID string) error {
	endpoint := "/api/public/delete_image"

	if imageID == "" {
		return fmt.Errorf("image_id must be provided")
	}

	reqPayload := DeleteImageRequest{
		APIKey:       api.APIKey,
		APIKeySecret: api.APIKeySecret,
	}
	reqPayload.TargetImage.ImageID = imageID

	_, err := api.MakePostRequest(ctx, endpoint, reqPayload)
	if err != nil {
		return fmt.Errorf("failed to delete image (id=%s) at %s: %w", imageID, endpoint, err)
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("image_id", imageID).
		Msg("Image deleted successfully")

	return nil
}
