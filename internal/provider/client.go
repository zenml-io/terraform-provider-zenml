package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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
	APITokenExpires *time.Time
	HTTPClient *http.Client
}

func NewClient(serverURL, apiKey string, apiToken string) *Client {
	return &Client{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		APIToken:   apiToken,
		APITokenExpires: nil,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) getAPIToken() (string, error) {
	if c.APIToken != "" {
		if c.APITokenExpires == nil {
			// No expiry, so just return the token
			return c.APIToken, nil
		}
		// Check if the token has expired
		if time.Now().Before(*c.APITokenExpires) {
			// Token is still valid
			return c.APIToken, nil
		}
	}

	if c.APIKey == "" {
		return "", fmt.Errorf("API key is required to get an API token")
	}

	// Get a new token from the API key using the password flow
	data := url.Values{}
	data.Set("password", c.APIKey)
	loginReq, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/v1/login", c.ServerURL),
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("error creating login request: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginResp, err := c.HTTPClient.Do(loginReq)
	if err != nil {
		return "", fmt.Errorf("error making login request: %v", err)
	}
	defer loginResp.Body.Close()
	
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("error decoding login response: %v", err)
	}
	
	c.APIToken = tokenResp.AccessToken
	// Set the expiry time to 5 minutes before the actual expiry, to account for
	// clock skew and to avoid using an expired token when making requests
	expiresAt := time.Now().Add(
		time.Duration(tokenResp.ExpiresIn - 300) * time.Second,
	)
	c.APITokenExpires = &expiresAt

	return c.APIToken, nil
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

	accessToken, err := c.getAPIToken()

	if err != nil {
		return nil, fmt.Errorf("error getting API token: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	log.Printf("[ZENML] Making request: %s %s", method, req.URL.String())
	log.Printf("[ZENML] Request body: %s", bodyReader)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	log.Printf("[ZENML] Response status: %d", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
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
	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/stacks", workspace)
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
	} else {
		if params.Page <= 0 {
			params.Page = 1
		}
		if params.PageSize <= 0 {
			params.PageSize = 100
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
	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/components", workspace)
	resp, err := c.doRequest("POST", endpoint, component)
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

func (c *Client) ListStackComponents(workspace string, params *ListParams) (*Page[ComponentResponse], error) {
	if params == nil {
		params = &ListParams{
			Page:     1,
			PageSize: 100,
		}
	} else {
		if params.Page <= 0 {
			params.Page = 1
		}
		if params.PageSize <= 0 {
			params.PageSize = 100
		}
	}
	
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", params.Page))
	query.Add("size", fmt.Sprintf("%d", params.PageSize))
	for k, v := range params.Filter {
		query.Add(k, v)
	}
	
	path := fmt.Sprintf("/api/v1/workspaces/%s/components?%s", workspace, query.Encode())
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
func (c *Client) VerifyServiceConnector(connector ServiceConnectorRequest) (*ServiceConnectorResources, error) {
	resp, err := c.doRequest("POST", "/api/v1/service_connectors/verify", connector)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResources
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) CreateServiceConnector(workspace string, connector ServiceConnectorRequest) (*ServiceConnectorResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/workspaces/%s/service_connectors", workspace)
	resp, err := c.doRequest("POST", endpoint, connector)
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

	log.Printf("[ZENML] Response body: %v", result)

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
	} else {
		if params.Page <= 0 {
			params.Page = 1
		}
		if params.PageSize <= 0 {
			params.PageSize = 100
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

// Add this new method to the Client
func (c *Client) GetWorkspaceByName(name string) (*WorkspaceResponse, error) {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/v1/workspaces/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result WorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Add this method to get the current user
func (c *Client) GetCurrentUser() (*UserResponse, error) {
	resp, err := c.doRequest("GET", "/api/v1/current-user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding user response: %v", err)
	}
	return &result, nil
}

