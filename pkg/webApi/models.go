package webApi

// USER API STRUCTS

// TargetUser represents the target user data for create, get, and update operations.
type TargetUser struct {
	UserID       string `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Locked       bool   `json:"locked,omitempty"`
	Disabled     bool   `json:"disabled,omitempty"`
	Organization string `json:"organization,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Password     string `json:"password,omitempty"`
	// Add other necessary fields as per API specifications
}

// UserGroup represents a user's group in the response.
type UserGroup struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
}

// KasmSession represents a user's active Kasm session.
type KasmSession struct {
	KasmID         string         `json:"kasm_id"`
	StartDate      string         `json:"start_date"`
	KeepaliveDate  string         `json:"keepalive_date"`
	ExpirationDate string         `json:"expiration_date"`
	Server         KasmServerInfo `json:"server"`
}

// KasmServerInfo represents the server information for a Kasm session.
type KasmServerInfo struct {
	ServerID string `json:"server_id"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

// UserResponse represents the response for user-related endpoints.
type UserResponse struct {
	UserID       string        `json:"user_id"`
	Username     string        `json:"username"`
	FirstName    *string       `json:"first_name,omitempty"` // Pointer to handle null
	LastName     *string       `json:"last_name,omitempty"`  // Pointer to handle null
	Phone        *string       `json:"phone,omitempty"`      // Pointer to handle null
	Organization *string       `json:"organization,omitempty"`
	Realm        string        `json:"realm"`
	LastSession  *string       `json:"last_session,omitempty"`
	Groups       []UserGroup   `json:"groups"`
	Kasms        []KasmSession `json:"kasms"`
	Disabled     bool          `json:"disabled"`
	Locked       bool          `json:"locked"`
	Created      string        `json:"created"`
	Notes        *string       `json:"notes,omitempty"` // Added Notes field based on new API
}

type GetUserResponse struct {
	User UserResponse `json:"user"`
}

// UserAttributes represents a user's attributes (preferences).
type UserAttributes struct {
	SSHPublicKey       string  `json:"ssh_public_key"`
	ShowTips           bool    `json:"show_tips"`
	UserID             string  `json:"user_id"`
	ToggleControlPanel bool    `json:"toggle_control_panel"`
	ChatSFX            bool    `json:"chat_sfx"`
	DefaultImage       *string `json:"default_image,omitempty"`
	AutoLoginKasm      *bool   `json:"auto_login_kasm,omitempty"`
	// Add other necessary fields as per API specifications
}

// KASM API STRUCTS

// RequestKasmRequest represents the request to start a Kasm session.
type RequestKasmRequest struct {
	APIKey         string            `json:"api_key"`
	APIKeySecret   string            `json:"api_key_secret"`
	UserID         string            `json:"user_id"`
	ImageID        string            `json:"image_id"`
	EnableSharing  bool              `json:"enable_sharing"`
	Environment    map[string]string `json:"environment,omitempty"`
	ClientLanguage *string           `json:"client_language,omitempty"`
	ClientTimezone *string           `json:"client_timezone,omitempty"`
	KasmURL        *string           `json:"kasm_url,omitempty"`
}

// RequestKasmResponse represents the response when a Kasm session is requested.
type RequestKasmResponse struct {
	KasmID       string `json:"kasm_id"`
	Username     string `json:"username"`
	Status       string `json:"status"`
	ShareID      string `json:"share_id"`
	UserID       string `json:"user_id"`
	SessionToken string `json:"session_token"`
	KasmURL      string `json:"kasm_url"`
}

// GetKasmStatusRequest represents the request to get the status of a Kasm session.
type GetKasmStatusRequest struct {
	APIKey         string `json:"api_key"`
	APIKeySecret   string `json:"api_key_secret"`
	UserID         string `json:"user_id"`
	KasmID         string `json:"kasm_id"`
	SkipAgentCheck bool   `json:"skip_agent_check,omitempty"`
}

// GetKasmStatusResponse represents the response for Kasm session status.
type GetKasmStatusResponse struct {
	OperationalMessage  string    `json:"operational_message"`
	OperationalProgress int       `json:"operational_progress"`
	OperationalStatus   string    `json:"operational_status"`
	Kasm                *KasmInfo `json:"kasm,omitempty"`
}

// KasmInfo represents the detailed Kasm session info.
type KasmInfo struct {
	ExpirationDate    string          `json:"expiration_date"`
	ContainerIP       string          `json:"container_ip"`
	ImageID           string          `json:"image_id"`
	OperationalStatus string          `json:"operational_status"`
	PortMap           map[string]Port `json:"port_map"`
	Hostname          string          `json:"hostname"`
	KasmID            string          `json:"kasm_id"`
	UserID            string          `json:"user_id"`
	Memory            int64           `json:"memory"`
	ShareID           string          `json:"share_id"`
	ClientSettings    ClientSettings  `json:"client_settings"`
	ContainerID       string          `json:"container_id"`
}

// Port represents port mappings in a Kasm session.
type Port struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}

// ClientSettings represents client-specific settings for a Kasm session.
type ClientSettings struct {
	AllowKasmAudio         bool `json:"allow_kasm_audio"`
	IdleDisconnect         int  `json:"idle_disconnect"`
	AllowKasmMicrophone    bool `json:"allow_kasm_microphone"`
	AllowPersistentProfile bool `json:"allow_persistent_profile"`
}

// DestroyKasmRequest represents the request to destroy a Kasm session.
type DestroyKasmRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
	KasmID       string `json:"kasm_id"`
	UserID       string `json:"user_id"`
}

// ExecCommandRequest represents the request to execute a command inside a Kasm session.
type ExecCommandRequest struct {
	APIKey       string            `json:"api_key"`
	APIKeySecret string            `json:"api_key_secret"`
	KasmID       string            `json:"kasm_id"`
	UserID       string            `json:"user_id"`
	ExecConfig   ExecConfigRequest `json:"exec_config"`
}

// ExecConfigRequest contains the execution configuration for a command.
type ExecConfigRequest struct {
	Cmd         string            `json:"cmd"`
	Environment map[string]string `json:"environment,omitempty"`
	Workdir     string            `json:"workdir,omitempty"`
	Privileged  bool              `json:"privileged,omitempty"`
	User        string            `json:"user,omitempty"`
}

// IMAGE API STRUCTS

// GetImagesRequest represents the request to retrieve available images.
type GetImagesRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
}

// ImageAttribute represents an attribute of a Kasm image.
type ImageAttribute struct {
	ImageID  string `json:"image_id"`
	AttrID   string `json:"attr_id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Value    string `json:"value"`
}

// RunConfig represents the configuration for running the Kasm image.
type RunConfig struct {
	Hostname string `json:"hostname"`
}

// ExecConfig represents the execution configuration for the Kasm image.
type ExecConfig struct {
	FirstLaunch struct {
		Environment map[string]string `json:"environment"`
		Cmd         string            `json:"cmd"`
	} `json:"first_launch"`
	Go struct {
		Cmd string `json:"cmd"`
	} `json:"go"`
}

// Image represents the details of a Kasm image.
type Image struct {
	ImageID                 string                 `json:"image_id"`
	FriendlyName            string                 `json:"friendly_name"`
	ImageTag                string                 `json:"name"`
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

// GetImagesResponse represents the response from the Kasm API when fetching images.
type GetImagesResponse struct {
	Images []Image `json:"images"`
}

// GROUP API STRUCTS

// GroupUserRequest represents the request to add or remove a user from a group.
type GroupUserRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
	TargetUser   struct {
		UserID string `json:"user_id"`
	} `json:"target_user"`
	TargetGroup struct {
		GroupID string `json:"group_id"`
	} `json:"target_group"`
}

// LoginRequest represents the request to generate a login link for a user.
type LoginRequest struct {
	APIKey       string `json:"api_key"`
	APIKeySecret string `json:"api_key_secret"`
	TargetUser   struct {
		UserID string `json:"user_id"`
	} `json:"target_user"`
}

// LoginResponse represents the response containing the login URL.
type LoginResponse struct {
	URL string `json:"url"`
}
