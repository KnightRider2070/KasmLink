package models

// Metadata represents a key-value pair for group metadata.
type Metadata struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value"`
}

// Group represents the basic details of a group.
type Group struct {
	Name        string `json:"name" validate:"required"`
	Priority    int    `json:"priority" validate:"oneof=high medium low"`
	Description string `json:"description,omitempty"`
}

// GroupStruct represents a detailed group including additional attributes.
type GroupStruct struct {
	Group
	GroupID        string     `json:"group_id" validate:"required"`
	IsSystem       bool       `json:"is_system"`
	GroupMetadata  []Metadata `json:"group_metadata,omitempty"`
	GroupMappings  []string   `json:"group_mappings,omitempty"`
	WorkspaceNames []string   `json:"workspace_names,omitempty"`
}

// GroupsResponse represents the response containing a list of groups.
type GroupsResponse struct {
	Groups []GroupStruct `json:"groups"`
}

// GroupRequest represents the overall request structure.
type GroupRequest struct {
	TargetGroup Group `json:"target_group" validate:"required"`
}
