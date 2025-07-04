// Package provider contains data models for the ZenML API
package provider

import (
	"encoding/json"
)

// Page represents a paginated response from the API
type Page[T any] struct {
	Index      int `json:"index"`
	MaxSize    int `json:"max_size"`
	TotalPages int `json:"total_pages"`
	Total      int `json:"total"`
	Items      []T `json:"items"`
}

// APIError represents an error response from the API
type APIError struct {
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return e.Detail
}

// ControlPlaneInfo represents the control plane information response
type ControlPlaneInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Version string            `json:"version"`
	URL     string            `json:"url"`
	Status  string            `json:"status"`
	Metadata map[string]string `json:"metadata"`
}

// ServerInfo represents the server information response from the API
type ServerInfo struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Version             string            `json:"version"`
	DeploymentType      string            `json:"deployment_type"`
	AuthScheme          string            `json:"auth_scheme"`
	ServerURL           string            `json:"server_url"`
	DashboardURL        string            `json:"dashboard_url"`
	ProDashboardURL     *string           `json:"pro_dashboard_url"`
	ProAPIURL           *string           `json:"pro_api_url"`
	ProOrganizationID   *string           `json:"pro_organization_id"`
	ProOrganizationName *string           `json:"pro_organization_name"`
	ProWorkspaceID      *string           `json:"pro_workspace_id"`
	ProWorkspaceName    *string           `json:"pro_workspace_name"`
	Metadata            map[string]string `json:"metadata"`
}

// Workspace models
type WorkspaceRequest struct {
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type WorkspaceResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Body        *WorkspaceResponseBody     `json:"body,omitempty"`
	Metadata    *WorkspaceResponseMetadata `json:"metadata,omitempty"`
	URL         string                     `json:"url"`
	Status      string                     `json:"status"`
}

type WorkspaceResponseBody struct {
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	Description string `json:"description"`
}

type WorkspaceResponseMetadata struct {
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type WorkspaceUpdate struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Team models
type TeamRequest struct {
	ControlPlaneID string   `json:"control_plane_id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	Members        []string `json:"members,omitempty"`
}

type TeamResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Body        *TeamResponseBody     `json:"body,omitempty"`
	Metadata    *TeamResponseMetadata `json:"metadata,omitempty"`
}

type TeamResponseBody struct {
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Description     string `json:"description"`
	ControlPlaneID  string `json:"control_plane_id"`
}

type TeamResponseMetadata struct {
	Members []TeamMember `json:"members,omitempty"`
}

type TeamMember struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TeamUpdate struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Members     []string `json:"members,omitempty"`
}

// Project models
type ProjectRequest struct {
	WorkspaceID string            `json:"workspace_id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ProjectResponse struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Body        *ProjectResponseBody     `json:"body,omitempty"`
	Metadata    *ProjectResponseMetadata `json:"metadata,omitempty"`
}

type ProjectResponseBody struct {
	Created     string `json:"created"`
	Updated     string `json:"updated"`
	Description string `json:"description"`
	WorkspaceID string `json:"workspace_id"`
}

type ProjectResponseMetadata struct {
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ProjectUpdate struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Role assignment models
type RoleAssignmentRequest struct {
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	UserID       *string `json:"user_id,omitempty"`
	TeamID       *string `json:"team_id,omitempty"`
	Role         string `json:"role"`
}

type RoleAssignmentResponse struct {
	ID           string                       `json:"id"`
	ResourceID   string                       `json:"resource_id"`
	ResourceType string                       `json:"resource_type"`
	Body         *RoleAssignmentResponseBody  `json:"body,omitempty"`
	Metadata     *RoleAssignmentResponseMetadata `json:"metadata,omitempty"`
}

type RoleAssignmentResponseBody struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
	UserID  *string `json:"user_id,omitempty"`
	TeamID  *string `json:"team_id,omitempty"`
	Role    string `json:"role"`
}

type RoleAssignmentResponseMetadata struct {
	User *UserResponse `json:"user,omitempty"`
	Team *TeamResponse `json:"team,omitempty"`
}

type RoleAssignmentUpdate struct {
	Role string `json:"role"`
}

// StackRequest represents a request to create a new stack
type StackRequest struct {
	Name       string              `json:"name"`
	Components map[string][]string `json:"components"` // Change to UUID strings
	Labels     map[string]string   `json:"labels"`
}

// StackResponse represents a stack response from the API
type StackResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Body     *StackResponseBody     `json:"body,omitempty"`
	Metadata *StackResponseMetadata `json:"metadata,omitempty"`
}

type StackResponseBody struct {
	Created string        `json:"created"`
	Updated string        `json:"updated"`
	User    *UserResponse `json:"user,omitempty"`
}

type StackResponseMetadata struct {
	Components map[string][]ComponentResponse `json:"components"`
	Labels     map[string]string              `json:"labels,omitempty"`
}

// StackUpdate represents an update to an existing stack
type StackUpdate struct {
	Name       *string             `json:"name,omitempty"`
	Components map[string][]string `json:"components,omitempty"` // Only UUIDs for updates
	Labels     map[string]string   `json:"labels,omitempty"`
}

// ComponentRequest represents a request to create a new component
type ComponentRequest struct {
	User                string                 `json:"user"`
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	Flavor              string                 `json:"flavor"`
	Configuration       map[string]interface{} `json:"configuration"`
	ConnectorID         *string                `json:"connector,omitempty"`
	ConnectorResourceID *string                `json:"connector_resource_id,omitempty"`
	Labels              map[string]string      `json:"labels,omitempty"`
}

// ComponentResponse represents a stack component response from the API
type ComponentResponse struct {
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Body     *ComponentResponseBody     `json:"body,omitempty"`
	Metadata *ComponentResponseMetadata `json:"metadata,omitempty"`
}

type ComponentResponseBody struct {
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
	User        *UserResponse `json:"user,omitempty"`
	Type        string        `json:"type"`
	Flavor      string        `json:"flavor_name"`
	Integration *string       `json:"integration,omitempty"`
}

type ComponentResponseMetadata struct {
	Configuration       map[string]interface{}    `json:"configuration"`
	Labels              map[string]string         `json:"labels,omitempty"`
	ConnectorResourceID *string                   `json:"connector_resource_id,omitempty"`
	Connector           *ServiceConnectorResponse `json:"connector,omitempty"`
}

// ComponentUpdate represents an update to an existing component
type ComponentUpdate struct {
	Name                *string                `json:"name,omitempty"`
	Type                *string                `json:"type,omitempty"`
	Flavor              *string                `json:"flavor,omitempty"`
	Configuration       map[string]interface{} `json:"configuration,omitempty"`
	ConnectorID         *string                `json:"connector,omitempty"`
	ConnectorResourceID *string                `json:"connector_resource_id,omitempty"`
	Labels              map[string]string      `json:"labels,omitempty"`
}

// ServiceConnectorRequest represents a request to create a new service connector
type ServiceConnectorRequest struct {
	User          string                 `json:"user"`
	Name          string                 `json:"name"`
	ConnectorType string                 `json:"connector_type"`
	AuthMethod    string                 `json:"auth_method"`
	ResourceTypes []string               `json:"resource_types"`
	Configuration map[string]interface{} `json:"configuration"`
	Secrets       map[string]string      `json:"secrets,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	ResourceID    *string                `json:"resource_id,omitempty"`
	ExpiresAt     *string                `json:"expires_at,omitempty"`
}

type ServiceConnectorResourceType struct {
	Name              string   `json:"name"`
	ResourceType      string   `json:"resource_type"`
	Description       string   `json:"description"`
	AuthMethods       []string `json:"auth_methods"`
	SupportsInstances bool     `json:"supports_instances"`
}

type ServiceConnectorAuthenticationMethod struct {
	Name        string `json:"name"`
	AuthMethod  string `json:"auth_method"`
	Description string `json:"description"`
}

type ServiceConnectorType struct {
	Name          string                                 `json:"name"`
	ConnectorType string                                 `json:"connector_type"`
	Description   string                                 `json:"description"`
	ResourceTypes []ServiceConnectorResourceType         `json:"resource_types"`
	AuthMethods   []ServiceConnectorAuthenticationMethod `json:"auth_methods"`
}

// ServiceConnectorResponse represents a service connector response from the API
type ServiceConnectorResponse struct {
	ID       string                            `json:"id"`
	Name     string                            `json:"name"`
	Body     *ServiceConnectorResponseBody     `json:"body,omitempty"`
	Metadata *ServiceConnectorResponseMetadata `json:"metadata,omitempty"`
}

type ServiceConnectorResponseBody struct {
	Created       string          `json:"created"`
	Updated       string          `json:"updated"`
	User          *UserResponse   `json:"user,omitempty"`
	ConnectorType json.RawMessage `json:"connector_type"` // Can be string or ServiceConnectorType
	AuthMethod    string          `json:"auth_method"`
	ResourceTypes []string        `json:"resource_types"`
	ResourceID    *string         `json:"resource_id,omitempty"`
	ExpiresAt     *string         `json:"expires_at,omitempty"`
}

type ServiceConnectorResponseMetadata struct {
	Configuration map[string]interface{} `json:"configuration"`
	SecretID      *string                `json:"secret_id,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
}

// ServiceConnectorResources represents a service connector resources response
// from the API
type ServiceConnectorResources struct {
	ID            string                           `json:"id"`
	Name          string                           `json:"name"`
	ConnectorType json.RawMessage                  `json:"connector_type"` // Can be string or ServiceConnectorType
	Resources     []ServiceConnectorTypedResources `json:"resources"`
	Error         *string                          `json:"error,omitempty"`
}

type ServiceConnectorTypedResources struct {
	ResourceType string   `json:"resource_type"`
	ResourceIDs  []string `json:"resource_ids"`
	Error        *string  `json:"error,omitempty"`
}

// ServiceConnectorUpdate represents an update to an existing service connector
type ServiceConnectorUpdate struct {
	Name          *string                `json:"name,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Secrets       map[string]string      `json:"secrets,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	ResourceTypes []string               `json:"resource_types"`
	ResourceID    *string                `json:"resource_id,omitempty"`
	ExpiresAt     *string                `json:"expires_at,omitempty"`
}

// UserResponse represents a user response from the API
type UserResponse struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Body             *UserResponseBody     `json:"body,omitempty"`
	Metadata         *UserResponseMetadata `json:"metadata,omitempty"`
	Resources        interface{}           `json:"resources"`
	PermissionDenied bool                  `json:"permission_denied"`
}

type UserResponseBody struct {
	Created          string  `json:"created"`
	Updated          string  `json:"updated"`
	Active           bool    `json:"active"`
	ActivationToken  *string `json:"activation_token"`
	FullName         string  `json:"full_name"`
	EmailOptedIn     bool    `json:"email_opted_in"`
	IsServiceAccount bool    `json:"is_service_account"`
	IsAdmin          bool    `json:"is_admin"`
}

type UserResponseMetadata struct {
	Email          *string      `json:"email"`
	ExternalUserID *string      `json:"external_user_id"`
	UserMetadata   UserMetadata `json:"user_metadata"`
}

type UserMetadata struct {
	PrimaryUse               string   `json:"primary_use"`
	UsageReason              string   `json:"usage_reason"`
	InfraProviders           []string `json:"infra_providers"`
	FinishedOnboardingSurvey bool     `json:"finished_onboarding_survey"`
	OverviewTourDone         bool     `json:"overview_tour_done"`
}
