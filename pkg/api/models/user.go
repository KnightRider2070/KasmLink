package models

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
}

type UserGroup struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
}

type UserAttributes struct {
	SSHPublicKey       string  `json:"ssh_public_key"`
	ShowTips           bool    `json:"show_tips"`
	UserID             string  `json:"user_id"`
	ToggleControlPanel bool    `json:"toggle_control_panel"`
	ChatSFX            bool    `json:"chat_sfx"`
	DefaultImage       *string `json:"default_image"`
	AutoLoginKasm      *bool   `json:"auto_login_kasm"`
}

type KasmSession struct {
	KasmID         string         `json:"kasm_id"`
	StartDate      string         `json:"start_date"`
	KeepaliveDate  string         `json:"keepalive_date"`
	ExpirationDate string         `json:"expiration_date"`
	Server         KasmServerInfo `json:"server"`
}
type KasmServerInfo struct {
	ServerID string `json:"server_id"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

type UserResponse struct {
	UserID           string        `json:"user_id"`
	Username         string        `json:"username"`
	FirstName        *string       `json:"first_name"`
	LastName         *string       `json:"last_name"`
	Phone            *string       `json:"phone"`
	Organization     *string       `json:"organization"`
	Realm            string        `json:"realm"`
	LastSession      *string       `json:"last_session"`
	Groups           []UserGroup   `json:"groups"`
	Kasms            []KasmSession `json:"kasms"`
	Disabled         bool          `json:"disabled"`
	Locked           bool          `json:"locked"`
	Created          string        `json:"created"`
	Notes            *string       `json:"notes"`
	TwoFactorEnabled bool          `json:"two_factor"`
	ProgramId        *string       `json:"program_id"`
	Hash             string        `json:"hash"`
}
