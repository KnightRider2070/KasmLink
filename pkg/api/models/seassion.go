package models

type RequestKasm struct {
	UserID                string            `json:"user_id"`
	ImageID               string            `json:"image_id"`
	EnableSharing         bool              `json:"enable_sharing"`
	Environment           map[string]string `json:"environment"`
	ConnectionInfo        map[string]string `json:"connection_info"`
	ClientLanguage        string            `json:"client_language"`
	ClientTimezone        string            `json:"client_timezone"`
	EgressGatewayId       string            `json:"egress_gateway_id"`
	PersistentProfileMode string            `json:"persistent_profile_mode"`
	RdpClientType         string            `json:"rdp_client_type"`
}

type RequestKasmResponse struct {
	RequestKasm
	KasmID       string `json:"kasm_id"`
	Username     string `json:"username"`
	Status       string `json:"status"`
	ShareID      string `json:"share_id"`
	SessionToken string `json:"session_token"`
	KasmURL      string `json:"kasm_url"`
}

type GetKasmStatus struct {
	UserID         string `json:"user_id"`
	KasmID         string `json:"kasm_id"`
	SkipAgentCheck bool   `json:"skip_agent_check"`
}

type GetKasmStatusResponse struct {
	OperationalMessage  string   `json:"operational_message"`
	OperationalProgress int      `json:"operational_progress"`
	OperationalStatus   string   `json:"operational_status"`
	CurrentTime         string   `json:"current_time"`
	KasmUrl             string   `json:"kasm_url"`
	Kasm                KasmInfo `json:"kasm"`
}

type KasmInfo struct {
	ExpirationDate        string          `json:"expiration_date"`
	ContainerIP           string          `json:"container_ip"`
	StartDate             string          `json:"start_date"`
	PointOfPresence       *string         `json:"point_of_presence"`
	Token                 string          `json:"token"`
	ImageID               string          `json:"image_id"`
	ViewOnlyToken         string          `json:"view_only_token"`
	Cores                 float64         `json:"cores"`
	Hostname              string          `json:"hostname"`
	KasmID                string          `json:"kasm_id"`
	PortMap               map[string]Port `json:"port_map"`
	Image                 ImageInfo       `json:"image"`
	IsPersistentProfile   string          `json:"is_persistent_profile"`
	Memory                int64           `json:"memory"`
	OperationalStatus     string          `json:"operational_status"`
	ClientSettings        ClientSettings  `json:"client_settings"`
	ContainerId           string          `json:"container_id"`
	Port                  int             `json:"port"`
	KeepAliveDate         string          `json:"keep_alive_date"`
	UserID                string          `json:"user_id"`
	PersistentProfileMode *string         `json:"persistent_profile_mode"`
	ShareID               string          `json:"share_id"`
	Host                  string          `json:"host"`
	ServerId              string          `json:"server_id"`
}

type ImageInfo struct {
	ImageId      string `json:"image_id"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	ImageSrc     string `json:"image_src"`
}

type Port struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}
type ClientSettings struct {
	AllowKasmAudio             bool `json:"allow_kasm_audio"`
	IdleDisconnect             int  `json:"idle_disconnect"`
	LockSharingVideoMode       bool `json:"lock_sharing_video_mode"`
	AllowPersistentProfile     bool `json:"allow_persistent_profile"`
	AllowKasmClipboardDown     bool `json:"allow_kasm_clipboard_down"`
	AllowKasmMicrophone        bool `json:"allow_kasm_microphone"`
	AllowKasmDownloads         bool `json:"allow_kasm_downloads"`
	KasmAudioDefaultOn         bool `json:"kasm_audio_default_on"`
	AllowPointOfPresence       bool `json:"allow_point_of_presence"`
	AllowKasmUploads           bool `json:"allow_kasm_uploads"`
	AllowKasmClipboardUp       bool `json:"allow_kasm_clipboard_up"`
	EnableWebp                 bool `json:"enable_webp"`
	AllowKasmSharing           bool `json:"allow_kasm_sharing"`
	AllowKasmClipboardSeamless bool `json:"allow_kasm_clipboard_seamless"`
}

type DestroyKasmRequest struct {
	UserID string `json:"user_id"`
	KasmID string `json:"kasm_id"`
}

type ExecCommandRequest struct {
	UserID            string `json:"user_id"`
	KasmID            string `json:"kasm_id"`
	ExecConfigRequest string `json:"exec_config"`
}
