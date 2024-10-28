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
	APIToken   string
	HTTPClient *http.Client
}

type ServerInfo struct {
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
}

func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

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

	// Get authorization token from login if APIKey is set, otherwise use APIToken
	var authToken string
	if c.APIKey != "" {
		data := url.Values{}
		data.Set("password", c.APIKey)
		loginReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/login", c.ServerURL), bytes.NewBufferString(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("error creating login request: %v", err)
		}
		loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		loginResp, err := c.HTTPClient.Do(loginReq)
		if err != nil {
			return nil, fmt.Errorf("error making login request: %v", err)
		}
		defer loginResp.Body.Close()
		
		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(loginResp.Body).Decode(&tokenResp); err != nil {
			return nil, fmt.Errorf("error decoding login response: %v", err)
		}
		authToken = tokenResp.AccessToken
	} else {
		authToken = c.APIToken
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
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
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, &apiError
	}

	return resp, nil
}

// GetServerInfo fetches server info to determine version and capabilities
func (c *Client) GetServerInfo() (*ServerInfo, error) {
	resp, err := c.doRequest("GET", "/api/v1/info", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding server info: %v", err)
	}
	return &result, nil
}

// Stack operations
func (c *Client) CreateStack(workspace string, stack StackRequest) (*StackResponse, error) {
	// Check server version to determine which endpoint to use
	info, err := c.GetServerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %v", err)
	}

	log.Printf("Server info: %+v", info)

	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/stacks", workspace)

	// Set workspace in request if not already set
	if stack.Workspace == nil {
		stack.Workspace = &workspace
	}

	resp, err := c.doRequest("POST", endpoint, stack)
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

// Remaining methods from the original client...
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
	
	for k, v := range params.Filter {
		query.Add(k, v)
	}
	
	path := fmt.Sprintf("/api/v1/stacks?%s", query.Encode())
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page[StackResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Component operations...
func (c *Client) CreateComponent(workspace string, component ComponentRequest) (*ComponentResponse, error) {
	// Ensure workspace and user are set in the request
	component.Workspace = workspace
	
	resp, err := c.doRequest("POST", fmt.Sprintf("/api/v1/workspaces/%s/components", workspace), component)
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

	var result Page[ComponentResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Service Connector operations...
func (c *Client) CreateServiceConnector(connector ServiceConnectorRequest) (*ServiceConnectorResponse, error) {
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

	var result Page[ServiceConnectorResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Add this new method to the Client
func (c *Client) GetServiceConnectorByName(workspace, name string) (*ServiceConnectorResponse, error) {
	params := &ListParams{
		Filter: map[string]string{
			"name": name,
			"workspace": workspace,
		},
	}
	
	connectors, err := c.ListServiceConnectors(params)
	if err != nil {
		return nil, err
	}
	
	if len(connectors.Items) == 0 {
		return nil, fmt.Errorf("no service connector found with name %s", name)
	}
	
	return &connectors.Items[0], nil
}
