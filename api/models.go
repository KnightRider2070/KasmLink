package api

// Image represents an image object.
type Image struct {
	ImageID        string `json:"image_id"`
	FriendlyName   string `json:"friendly_name"`
	DockerRegistry string `json:"docker_registry"`
	Available      bool   `json:"available"`
}

// CreateUserRequest represents the request to create a user.
type CreateUserRequest struct {
	APIKey       string   `json:"api_key"`
	APIKeySecret string   `json:"api_key_secret"`
	TargetUser   UserInfo `json:"target_user"`
}

// UpdateUserRequest represents the request to update a user.
type UpdateUserRequest struct {
	APIKey       string   `json:"api_key"`
	APIKeySecret string   `json:"api_key_secret"`
	TargetUser   UserInfo `json:"target_user"`
}

// UpdateUserAttributesRequest represents the request to update user attributes.
type UpdateUserAttributesRequest struct {
	APIKey             string `json:"api_key"`
	APIKeySecret       string `json:"api_key_secret"`
	UserID             string `json:"user_id"`
	ToggleControlPanel bool   `json:"toggle_control_panel"`
	AutoLoginKasm      bool   `json:"auto_login_kasm"`
	ShowTips           bool   `json:"show_tips"`
	DefaultImage       string `json:"default_image"`
}

// UserInfo contains user details required to create or update a user.
type UserInfo struct {
	UserID       string `json:"user_id,omitempty"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Locked       bool   `json:"locked"`
	Disabled     bool   `json:"disabled"`
	Organization string `json:"organization"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
}

// UserResponse represents the response containing user details.
type UserResponse struct {
	UserID       string      `json:"user_id"`
	Username     string      `json:"username"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	Locked       bool        `json:"locked"`
	Disabled     bool        `json:"disabled"`
	Organization string      `json:"organization"`
	Phone        string      `json:"phone"`
	Groups       []UserGroup `json:"groups"`
	Notes        string      `json:"notes"`
	Realm        string      `json:"realm"`
	LastSession  string      `json:"last_session"`
}

// UserGroup represents the group information for a user.
type UserGroup struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
}

// UserAttributesResponse represents the response containing user attributes.
type UserAttributesResponse struct {
	UserAttributes UserAttributes `json:"user_attributes"`
}

// UserAttributes represents the user attribute settings.
type UserAttributes struct {
	UserID             string `json:"user_id"`
	ToggleControlPanel bool   `json:"toggle_control_panel"`
	AutoLoginKasm      bool   `json:"auto_login_kasm"`
	ShowTips           bool   `json:"show_tips"`
	DefaultImage       string `json:"default_image"`
}

// KasmSessionResponse represents the response containing Kasm session details.
type KasmSessionResponse struct {
	KasmID       string `json:"kasm_id"`
	Username     string `json:"username"`
	Status       string `json:"status"`
	SessionToken string `json:"session_token"`
	KasmURL      string `json:"kasm_url"`
}

// KasmStatusResponse represents the response containing the status of a Kasm session.
type KasmStatusResponse struct {
	OperationalStatus   string `json:"operational_status"`
	OperationalMessage  string `json:"operational_message"`
	OperationalProgress int    `json:"operational_progress"`
}

// ExecConfig represents the configuration for executing a command in a Kasm session.
type ExecConfig struct {
	Cmd         string            `json:"cmd"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkDir     string            `json:"workdir,omitempty"`
	Privileged  bool              `json:"privileged,omitempty"`
	User        string            `json:"user,omitempty"`
}

// ExecCommandResponse represents the response after executing a command in a Kasm session.
type ExecCommandResponse struct {
	KasmID      string `json:"kasm_id"`
	CurrentTime string `json:"current_time"`
}
