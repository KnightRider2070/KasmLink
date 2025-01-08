package models

// Group represents an individual group
type Group struct {
	GroupID        string                 `json:"group_id"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description"`
	Priority       string                 `json:"priority"`
	IsSystem       bool                   `json:"is_system"`
	GroupMetadata  map[string]interface{} `json:"group_metadata"`
	GroupMappings  []interface{}          `json:"group_mappings"`
	WorkspaceNames []string               `json:"workspace_names"`
}

// GroupsResponse represents the response containing a list of groups.
type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

// TargetGroup represents the group details in the request.
type TargetGroup struct {
	Name        string `json:"name"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
}

// GroupRequest represents the overall request structure.
type GroupRequest struct {
	TargetGroup TargetGroup `json:"target_group"`
}
