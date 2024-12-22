package models

type ImageAttribute struct {
	ImageID  string `json:"image_id"`
	AttrID   string `json:"attr_id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Value    string `json:"value"`
}

type RunConfig struct {
	Hostname string `json:"hostname"`
}

type ExecConfig struct {
	FirstLaunch struct {
		Environment map[string]string `json:"environment"`
		Cmd         string            `json:"cmd"`
	} `json:"first_launch"`
	Go struct {
		Cmd string `json:"cmd"`
	} `json:"go"`
}

type Image struct {
	ImageID                 string                 `json:"image_id"`
	FriendlyName            string                 `json:"friendly_name"`
	Description             string                 `json:"description"`
	Memory                  int64                  `json:"memory"`
	Cores                   float64                `json:"cores"`
	XRes                    int                    `json:"x_res"`
	YRes                    int                    `json:"y_res"`
	Enabled                 bool                   `json:"enabled"`
	Available               bool                   `json:"available"`
	ImageAttributes         []ImageAttribute       `json:"imageAttributes"`
	ExecConfig              ExecConfig             `json:"exec_config"`
	RunConfig               RunConfig              `json:"run_config"`
	PersistentProfilePath   *string                `json:"persistent_profile_path,omitempty"`
	DockerRegistry          string                 `json:"docker_registry"`
	DockerToken             *string                `json:"docker_token,omitempty"`
	VolumeMappings          map[string]string      `json:"volume_mappings"`
	RestrictToNetwork       bool                   `json:"restrict_to_network"`
	RestrictToZone          bool                   `json:"restrict_to_zone"`
	RestrictToServer        bool                   `json:"restrict_to_server"`
	ServerID                *string                `json:"server_id,omitempty"`
	ZoneID                  *string                `json:"zone_id,omitempty"`
	ZoneName                *string                `json:"zone_name,omitempty"`
	DockerUser              *string                `json:"docker_user,omitempty"`
	CPUAllocationMethod     string                 `json:"cpu_allocation_method"`
	PersistentProfileConfig map[string]interface{} `json:"persistent_profile_config,omitempty"`
	ImageSrc                string                 `json:"image_src"`
}

type GetImagesResponse struct {
	Images []Image `json:"images"`
}
