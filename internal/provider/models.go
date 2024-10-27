// models.go
package provider

import "fmt"
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

// ComponentBody represents the common fields for component requests/responses
type ComponentBody struct {
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	Flavor              string                 `json:"flavor"`
	Configuration       map[string]interface{} `json:"configuration"`
	Workspace           string                 `json:"workspace"`
	User                string                 `json:"user,omitempty"`
	ConnectorResourceID string                 `json:"connector_resource_id,omitempty"`
	Labels              map[string]string      `json:"labels,omitempty"`
}

// ComponentResponse represents the API response for a component
type ComponentResponse struct {
	ID   string        `json:"id"`
	Body *ComponentBody `json:"body"`
}

type ComponentUpdate struct {
	Name              string                 `json:"name,omitempty"`
	Configuration     map[string]interface{} `json:"configuration,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty"`
	ComponentSpecPath *string                `json:"component_spec_path,omitempty"`
	Connector         *string                `json:"connector,omitempty"`
}

type ComponentCreate struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Flavor        string                 `json:"flavor"`
	Configuration map[string]interface{} `json:"configuration"`
	Workspace     string                 `json:"workspace"`
	User          string                 `json:"user"`
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
	Detail  string `json:"detail"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (code: %d): %s - %s", e.Code, e.Message, e.Detail)
}
