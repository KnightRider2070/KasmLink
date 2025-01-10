package dockercompose

type DockerCompose struct {
	Version  string                       `json:"version,omitempty"`
	Name     string                       `json:"name,omitempty"`
	Include  []Include                    `json:"include,omitempty"`
	Services map[string]ServiceDefinition `json:"services,omitempty"`
	Networks map[string]NetworkDefinition `json:"networks,omitempty"`
	Volumes  map[string]VolumeDefinition  `json:"volumes,omitempty"`
	Secrets  map[string]SecretDefinition  `json:"secrets,omitempty"`
	Configs  map[string]ConfigDefinition  `json:"configs,omitempty"`
}

type Include struct {
	Path             string `json:"path,omitempty"`
	EnvFile          string `json:"env_file,omitempty"`
	ProjectDirectory string `json:"project_directory,omitempty"`
}

type ServiceDefinition struct {
	Image         string                `json:"image,omitempty"`
	ContainerName string                `json:"container_name,omitempty"`
	Build         Build                 `json:"build,omitempty"`
	Environment   map[string]string     `json:"environment,omitempty"`
	Command       []string              `json:"command,omitempty"`
	Entrypoint    []string              `json:"entrypoint,omitempty"`
	Ports         []PortMapping         `json:"ports,omitempty"`
	Volumes       []VolumeMount         `json:"volumes,omitempty"`
	Networks      []string              `json:"networks,omitempty"`
	Restart       string                `json:"restart,omitempty"`
	DependsOn     []DependencyCondition `json:"depends_on,omitempty"`
	Healthcheck   *Healthcheck          `json:"healthcheck,omitempty"`
	Deployment    *DeploymentConfig     `json:"deploy,omitempty"`
	ExtraHosts    map[string]string     `json:"extra_hosts,omitempty"`
	Secrets       []SecretMount         `json:"secrets,omitempty"`
	Configs       []ConfigMount         `json:"configs,omitempty"`
	Labels        map[string]string     `json:"labels,omitempty"`
	DNS           []string              `json:"dns,omitempty"`
	DNSSearch     []string              `json:"dns_search,omitempty"`
	Tmpfs         []string              `json:"tmpfs,omitempty"`
	Isolation     string                `json:"isolation,omitempty"`
	Runtime       string                `json:"runtime,omitempty"`
	Profiles      []string              `json:"profiles,omitempty"`
	Capabilities  *Capabilities         `json:"capabilities,omitempty"`
	Logging       *Logging              `json:"logging,omitempty"`
	Scale         int                   `json:"scale,omitempty"`
	PullPolicy    string                `json:"pull_policy,omitempty"`
	StopSignal    string                `json:"stop_signal,omitempty"`
	Sysctls       map[string]string     `json:"sysctls,omitempty"`
	Pid           string                `json:"pid,omitempty"`
	Ipc           string                `json:"ipc,omitempty"`
	Uts           string                `json:"uts,omitempty"`
	Extends       *Extends              `json:"extends,omitempty"`
}

type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

type Logging struct {
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

type Build struct {
	Context    string            `json:"context,omitempty"`
	Dockerfile string            `json:"dockerfile,omitempty"`
	Args       map[string]string `json:"args,omitempty"`
	CacheFrom  []string          `json:"cache_from,omitempty"`
	Target     string            `json:"target,omitempty"`
	Network    string            `json:"network,omitempty"`
	ExtraHosts map[string]string `json:"extra_hosts,omitempty"`
}

type PortMapping struct {
	Target    int    `json:"target,omitempty"`
	Published int    `json:"published,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

type VolumeMount struct {
	Type        string         `json:"type,omitempty"`
	Source      string         `json:"source,omitempty"`
	Target      string         `json:"target,omitempty"`
	ReadOnly    bool           `json:"read_only,omitempty"`
	Consistency string         `json:"consistency,omitempty"`
	Bind        *BindOptions   `json:"bind,omitempty"`
	Volume      *VolumeOptions `json:"volume,omitempty"`
	Tmpfs       *TmpfsOptions  `json:"tmpfs,omitempty"`
	Selinux     string         `json:"selinux_label,omitempty"`
	SubPath     string         `json:"subpath,omitempty"`
}

type BindOptions struct {
	Propagation string `json:"propagation,omitempty"`
}

type VolumeOptions struct {
	NoCopy bool `json:"nocopy,omitempty"`
}

type TmpfsOptions struct {
	Size int `json:"size,omitempty"`
}

type DependencyCondition struct {
	Service   string `json:"service"`
	Condition string `json:"condition"`
}

type Healthcheck struct {
	Test        []string `json:"test,omitempty"`
	Interval    string   `json:"interval,omitempty"`
	Timeout     string   `json:"timeout,omitempty"`
	Retries     int      `json:"retries,omitempty"`
	StartPeriod string   `json:"start_period,omitempty"`
}

type DeploymentConfig struct {
	Mode          string           `json:"mode,omitempty"`
	Replicas      int              `json:"replicas,omitempty"`
	UpdateConfig  *UpdateConfig    `json:"update_config,omitempty"`
	RestartPolicy *RestartPolicy   `json:"restart_policy,omitempty"`
	Placement     *PlacementConfig `json:"placement,omitempty"`
	Resources     *ResourceConfig  `json:"resources,omitempty"`
}

type UpdateConfig struct {
	Parallelism   int    `json:"parallelism,omitempty"`
	Delay         string `json:"delay,omitempty"`
	FailureAction string `json:"failure_action,omitempty"`
}

type RestartPolicy struct {
	Condition   string `json:"condition,omitempty"`
	Delay       string `json:"delay,omitempty"`
	MaxAttempts int    `json:"max_attempts,omitempty"`
	Window      string `json:"window,omitempty"`
}

type CPUConstraints struct {
	CPUShares    int    `json:"cpu_shares,omitempty"`
	CPUQuota     int    `json:"cpu_quota,omitempty"`
	CPUPeriod    int    `json:"cpu_period,omitempty"`
	CPURtRuntime int    `json:"cpu_rt_runtime,omitempty"`
	CPURtPeriod  int    `json:"cpu_rt_period,omitempty"`
	Cpuset       string `json:"cpuset,omitempty"`
}

type Extends struct {
	Service string `json:"service,omitempty"`
	File    string `json:"file,omitempty"`
}

type PlacementConfig struct {
	Constraints        []string `json:"constraints,omitempty"`
	Preferences        []string `json:"preferences,omitempty"`
	MaxReplicasPerNode int      `json:"max_replicas_per_node,omitempty"`
}

type ResourceConfig struct {
	Limits       *ResourceLimits `json:"limits,omitempty"`
	Reservations *ResourceLimits `json:"reservations,omitempty"`
}

type ResourceLimits struct {
	Cpus   string `json:"cpus,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type SecretMount struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	UID    string `json:"uid,omitempty"`
	GID    string `json:"gid,omitempty"`
	Mode   int    `json:"mode,omitempty"`
}

type ConfigMount struct {
	Source string `json:"source,omitempty"`
	Target string `json:"target,omitempty"`
	UID    string `json:"uid,omitempty"`
	GID    string `json:"gid,omitempty"`
	Mode   int    `json:"mode,omitempty"`
}

type NetworkDefinition struct {
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
	IPAM       *IPAMConfig       `json:"ipam,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Attachable bool              `json:"attachable,omitempty"`
	Internal   bool              `json:"internal,omitempty"`
	EnableIPv6 bool              `json:"enable_ipv6,omitempty"`
	External   *ExternalNetwork  `json:"external,omitempty"`
	Aliases    []string          `json:"aliases,omitempty"`
	Priority   int               `json:"priority,omitempty"`
}

type ExternalNetwork struct {
	Name string `json:"name,omitempty"`
}

type IPAMConfig struct {
	Driver string          `json:"driver,omitempty"`
	Config []IPAMSubConfig `json:"config,omitempty"`
}

type IPAMSubConfig struct {
	Subnet     string            `json:"subnet,omitempty"`
	Gateway    string            `json:"gateway,omitempty"`
	IPRange    string            `json:"ip_range,omitempty"`
	AuxAddress map[string]string `json:"aux_addresses,omitempty"`
}

type VolumeDefinition struct {
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
	External   bool              `json:"external,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

type SecretDefinition struct {
	Name           string            `json:"name,omitempty"`
	External       bool              `json:"external,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Driver         string            `json:"driver,omitempty"`
	DriverOpts     map[string]string `json:"driver_opts,omitempty"`
	TemplateDriver string            `json:"template_driver,omitempty"`
}

type ConfigDefinition struct {
	Name           string            `json:"name,omitempty"`
	File           string            `json:"file,omitempty"`
	External       bool              `json:"external,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	TemplateDriver string            `json:"template_driver,omitempty"`
}
