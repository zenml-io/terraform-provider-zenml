// models.go
package provider

// Common response format
type Page[T any] struct {
	Total  int64 `json:"total"`
	Items  []T   `json:"items"`
	Cursor any   `json:"cursor,omitempty"`
}

// Stack models
type StackResponse struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Components map[string]Component `json:"components"`
	Labels     map[string]string    `json:"labels,omitempty"`
	Metadata   *ResponseMetadata    `json:"metadata,omitempty"`
}

type StackUpdate struct {
	Name       string               `json:"name,omitempty"`
	Components map[string]Component `json:"components,omitempty"`
	Labels     map[string]string    `json:"labels,omitempty"`
}

// Component models
type Component struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name,omitempty"`
	Type          string                 `json:"type,omitempty"`
	Flavor        string                 `json:"flavor,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	ConnectorID   string                 `json:"connector_id,omitempty"`
}

type ComponentResponse struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Flavor   string            `json:"flavor"`
	Body     *ComponentBody    `json:"body,omitempty"`
	Metadata *ResponseMetadata `json:"metadata,omitempty"`
}

type ComponentBody struct {
	User                string                 `json:"user"`
	Workspace           string                 `json:"workspace"`
	Configuration       map[string]interface{} `json:"configuration"`
	ConnectorResourceID *string                `json:"connector_resource_id,omitempty"`
	Labels              map[string]string      `json:"labels,omitempty"`
	ComponentSpecPath   *string                `json:"component_spec_path,omitempty"`
	Connector           *string                `json:"connector,omitempty"`
}

type ComponentUpdate struct {
	Name              string                 `json:"name,omitempty"`
	Configuration     map[string]interface{} `json:"configuration,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty"`
	ComponentSpecPath *string                `json:"component_spec_path,omitempty"`
	Connector         *string                `json:"connector,omitempty"`
}

// Service Connector models
type ServiceConnectorResponse struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Type       string                `json:"type"`
	AuthMethod string                `json:"auth_method"`
	Body       *ServiceConnectorBody `json:"body,omitempty"`
	Metadata   *ResponseMetadata     `json:"metadata,omitempty"`
}

type ServiceConnectorBody struct {
	User          string                 `json:"user"`
	Workspace     string                 `json:"workspace"`
	Configuration map[string]interface{} `json:"configuration"`
	Secrets       map[string]interface{} `json:"secrets,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	ResourceTypes []string               `json:"resource_types,omitempty"`
}

type ServiceConnectorUpdate struct {
	Name          string                 `json:"name,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Secrets       map[string]interface{} `json:"secrets,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	ResourceTypes []string               `json:"resource_types,omitempty"`
}

type ResponseMetadata struct {
	Created   string `json:"created"`
	Updated   string `json:"updated"`
	User      string `json:"user"`
	Workspace string `json:"workspace"`
}

// APIError represents an error response from the API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (code: %d): %s - %s", e.Code, e.Message, e.Details)
}
