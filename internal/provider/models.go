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
	ID             *string `json:"id,omitempty"`
	Name           *string `json:"name,omitempty"`
	DisplayName    *string `json:"display_name,omitempty"`
	Description    *string `json:"description,omitempty"`
	LogoURL        *string `json:"logo_url,omitempty"`
	OwnerID        *string `json:"owner_id,omitempty"`
	OrganizationID *string `json:"organization_id,omitempty"`
	IsManaged      bool    `json:"is_managed"`
	EnrollmentKey  *string `json:"enrollment_key,omitempty"`
}

type WorkspaceResponse struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	DisplayName    string                  `json:"display_name"`
	Description    *string                 `json:"description,omitempty"`
	LogoURL        *string                 `json:"logo_url,omitempty"`
	Organization   OrganizationResponse    `json:"organization"`
	Owner          UserResponse            `json:"owner"`
	IsManaged      bool                    `json:"is_managed"`
	EnrollmentKey  *string                 `json:"enrollment_key,omitempty"`
	ZenMLService   ZenMLServiceResponse    `json:"zenml_service"`
	MLflowService  *MLflowServiceResponse  `json:"mlflow_service,omitempty"`
	UsageCounts    map[string]int          `json:"usage_counts"`
	DesiredState   string                  `json:"desired_state"`
	StateReason    string                  `json:"state_reason"`
	Status         string                  `json:"status"`
	Created        string                  `json:"created"`
	Updated        string                  `json:"updated"`
	StatusUpdated  *string                 `json:"status_updated,omitempty"`
}

type OrganizationResponse struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Description           *string `json:"description,omitempty"`
	LogoURL              *string `json:"logo_url,omitempty"`
	Created              string  `json:"created"`
	Updated              string  `json:"updated"`
	Owner                UserResponse `json:"owner"`
	HasActiveSubscription *bool   `json:"has_active_subscription,omitempty"`
	TrialExpiry          *int    `json:"trial_expiry,omitempty"`
}

type ZenMLServiceResponse struct {
	Configuration *ZenMLServiceConfiguration `json:"configuration,omitempty"`
	Status        *ZenMLServiceStatus         `json:"status,omitempty"`
}

type ZenMLServiceConfiguration struct {
	Version          string `json:"version"`
	AnalyticsEnabled bool   `json:"analytics_enabled"`
}

type ZenMLServiceStatus struct {
	ServerURL    *string `json:"server_url,omitempty"`
	Version      *string `json:"version,omitempty"`
	StorageSize  *int    `json:"storage_size,omitempty"`
}

type MLflowServiceResponse struct {
	Configuration MLflowServiceConfiguration `json:"configuration"`
	Status        *MLflowServiceStatus        `json:"status,omitempty"`
}

type MLflowServiceConfiguration struct {
	Version string `json:"version"`
}

type MLflowServiceStatus struct {
	ServerURL string `json:"server_url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type WorkspaceUpdate struct {
	OwnerID      *string `json:"owner_id,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	Description  *string `json:"description,omitempty"`
	LogoURL      *string `json:"logo_url,omitempty"`
	DesiredState *string `json:"desired_state,omitempty"`
	StateReason  *string `json:"state_reason,omitempty"`
}

// Team models
type TeamRequest struct {
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	OrganizationID string  `json:"organization_id"`
}

type TeamResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Created     string  `json:"created"`
	Updated     string  `json:"updated"`
	MemberCount int     `json:"member_count"`
}

type TeamMemberResponse struct {
	User *UserResponse `json:"user,omitempty"`
	Team *TeamResponse `json:"team,omitempty"`
	Roles []MemberRoleAssignment `json:"roles"`
}

type MemberRoleAssignment struct {
	RoleID  string `json:"role_id"`
	Level   string `json:"level"`
	ViaTeam bool   `json:"via_team"`
}

type TeamUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
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
	RoleID       string  `json:"role_id"`
	UserID       *string `json:"user_id,omitempty"`
	TeamID       *string `json:"team_id,omitempty"`
	WorkspaceID  *string `json:"workspace_id,omitempty"`
	ProjectID    *string `json:"project_id,omitempty"`
}

type RoleAssignmentResponse struct {
	User           *UserResponse `json:"user,omitempty"`
	Team           *TeamResponse `json:"team,omitempty"`
	Role           RoleResponse  `json:"role"`
	OrganizationID string        `json:"organization_id"`
	WorkspaceID    *string       `json:"workspace_id,omitempty"`
	ProjectID      *string       `json:"project_id,omitempty"`
}

type RoleResponse struct {
	ID             string  `json:"id"`
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	OrganizationID string  `json:"organization_id"`
	Level          string  `json:"level"`
	SystemManaged  bool    `json:"system_managed"`
	Type           string  `json:"type"`
}

type RoleAssignmentUpdate struct {
	UserID      *string `json:"user_id,omitempty"`
	TeamID      *string `json:"team_id,omitempty"`
	WorkspaceID *string `json:"workspace_id,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
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
