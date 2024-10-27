// client.go
package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"log"
)

type ListParams struct {
	Page     int
	PageSize int
	Filter   map[string]string
}

type Client struct {
	ServerURL  string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.ServerURL, path), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var apiError APIError
		if resp.Body == nil {
			return nil, fmt.Errorf("API request failed with status %d: no response body", resp.StatusCode)
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			// If we can't decode the error response, return a generic error
			body, _ := io.ReadAll(resp.Body)  // Ignoring error from ReadAll
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, &apiError
	}

	return resp, nil
}

// Stack operations
func (c *Client) CreateStack(stack StackUpdate) (*StackResponse, error) {
	resp, err := c.doRequest("POST", "/api/v1/stacks", stack)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetStack(id string) (*StackResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/stacks/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateStack(id string, stack StackUpdate) (*StackResponse, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/v1/stacks/%s", id), stack)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteStack(id string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/stacks/%s", id), nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Component operations
func (c *Client) CreateComponent(body ComponentBody) (*ComponentResponse, error) {
	// First get the workspace UUID
	workspaceID, err := c.GetWorkspaceByName(body.Workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace ID: %w", err)
	}

	// Update the body with workspace UUID
	body.Workspace = workspaceID

	url := fmt.Sprintf("%s/api/v1/workspaces/%s/components", c.ServerURL, workspaceID)
	
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[DEBUG] Making request to %s with body: %s", url, string(reqBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("[DEBUG] Response status: %d, body: %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error (code: %d): %s", resp.StatusCode, string(respBody))
	}

	var result ComponentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetComponent(id string) (*ComponentResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/components/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetComponentByName(name, workspace string) (*ComponentResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/components?name=%s&workspace=%s", name, workspace), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateComponent(id string, component ComponentUpdate) (*ComponentResponse, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/v1/components/%s", id), component)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteComponent(id string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/components/%s", id), nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// client.go (add these methods)

func (c *Client) CreateServiceConnector(connector ServiceConnectorBody) (*ServiceConnectorResponse, error) {
	resp, err := c.doRequest("POST", "/api/v1/service_connectors", connector)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetServiceConnector(id string) (*ServiceConnectorResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/service_connectors/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetServiceConnectorByName(name, workspace string) (*ServiceConnectorResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/service_connectors?name=%s&workspace=%s", name, workspace), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateServiceConnector(id string, connector ServiceConnectorUpdate) (*ServiceConnectorResponse, error) {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/v1/service_connectors/%s", id), connector)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteServiceConnector(id string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/service_connectors/%s", id), nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListStacks(params *ListParams) (*Page[StackResponse], error) {
	if params == nil {
		params = &ListParams{
			Page:     1,
			PageSize: 100,
		}
	}

	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", params.Page))
	query.Add("size", fmt.Sprintf("%d", params.PageSize))
	
	// Add filters if any
	for k, v := range params.Filter {
		query.Add(k, v)
	}
	
	path := fmt.Sprintf("/api/v1/stacks?%s", query.Encode())
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result Page[StackResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Add pagination support to all list methods
func (c *Client) ListStackComponents(params *ListParams) (*Page[ComponentResponse], error) {
	if params == nil {
		params = &ListParams{
			Page:     1,
			PageSize: 100,
		}
	}
	
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", params.Page))
	query.Add("size", fmt.Sprintf("%d", params.PageSize))
	for k, v := range params.Filter {
		query.Add(k, v)
	}
	
	path := fmt.Sprintf("/api/v1/components?%s", query.Encode())
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result Page[ComponentResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

func (c *Client) ListServiceConnectors(params *ListParams) (*Page[ServiceConnectorResponse], error) {
	if params == nil {
		params = &ListParams{
			Page:     1,
			PageSize: 100,
		}
	}
	
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", params.Page))
	query.Add("size", fmt.Sprintf("%d", params.PageSize))
	for k, v := range params.Filter {
		query.Add(k, v)
	}
	
	path := fmt.Sprintf("/api/v1/service_connectors?%s", query.Encode())
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result Page[ServiceConnectorResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Add the GetWorkspaceByName method
func (c *Client) GetWorkspaceByName(name string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/workspaces", c.ServerURL)
	
	// Add query parameter for filtering by name
	query := url + fmt.Sprintf("?name=%s", name)
	
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("workspace %s not found", name)
	}

	return result.Items[0].ID, nil
}
