package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ListParams struct {
	Page     int
	PageSize int
	Filter   map[string]string
}

type Client struct {
	ServerURL         string
	ControlPlaneURL   string
	APIKey            string
	APIToken          string
	ClientID          string
	ClientSecret      string
	AccessToken       string
	AccessTokenExpires *time.Time
	APITokenExpires   *time.Time
	HTTPClient        *http.Client
	workspaceURLs     map[string]string // Cache for workspace URLs
}

func NewClient(serverURL, controlPlaneURL, apiKey, apiToken, clientID, clientSecret string) *Client {
	return &Client{
		ServerURL:         serverURL,
		ControlPlaneURL:   controlPlaneURL,
		APIKey:            apiKey,
		APIToken:          apiToken,
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		APITokenExpires:   nil,
		HTTPClient:        &http.Client{},
		workspaceURLs:     make(map[string]string),
	}
}

func (c *Client) getAPIToken(ctx context.Context) (string, error) {
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
		if c.APIKey == "" {
			// Token has expired and we can't refresh it
			return "", fmt.Errorf(`The API token configured for the ZenML Terraform provider has expired.

Please reconfigure the provider with a new API token or an API key.
It is recommended to use an API key for long-term Terraform management operations, as API tokens expire after a short period of time.

More information on how to configure a service account and an API key can be found at https://docs.zenml.io/how-to/connecting-to-zenml/connect-with-a-service-account.

To configure the ZenML Terraform provider, add the following block to your Terraform configuration:

provider "zenml" {
	server_url = "https://example.zenml.io"
	api_key   = "your api key"
}

or use the ZENML_API_KEY environment variable to set the API key.
`)
		}
	} else if c.APIKey == "" {
		// Shouldn't happen, as the provider should have already validated this.
		return "", fmt.Errorf("an API key or an API token must be configured for the ZenML Terraform provider to be able to authenticate with your ZenML server")
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
		time.Duration(tokenResp.ExpiresIn-300) * time.Second,
	)
	c.APITokenExpires = &expiresAt

	return c.APIToken, nil
}

func (c *Client) getOAuth2Token(ctx context.Context) (string, error) {
	if c.ClientID == "" || c.ClientSecret == "" {
		return "", fmt.Errorf("OAuth2 client ID and secret are required for control plane operations")
	}

	// Check if we have a valid access token
	if c.AccessToken != "" && c.AccessTokenExpires != nil {
		if time.Now().Before(*c.AccessTokenExpires) {
			return c.AccessToken, nil
		}
	}

	// Get new access token via OAuth2 client credentials flow
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("audience", "https://cloudapi.zenml.io")

	tokenReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://zenmlcloud.eu.auth0.com/oauth/token",
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("error creating OAuth2 token request: %v", err)
	}
	
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	tokenResp, err := c.HTTPClient.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("error making OAuth2 token request: %v", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("OAuth2 token request failed with status %d: %s", tokenResp.StatusCode, string(bodyBytes))
	}

	var tokenData struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		return "", fmt.Errorf("error decoding OAuth2 token response: %v", err)
	}

	c.AccessToken = tokenData.AccessToken
	// Set expiry time to 5 minutes before actual expiry to handle clock skew
	expiresAt := time.Now().Add(time.Duration(tokenData.ExpiresIn-300) * time.Second)
	c.AccessTokenExpires = &expiresAt

	return c.AccessToken, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, int, error) {
	return c.doRequestWithBaseURL(ctx, method, path, body, c.ServerURL)
}

func (c *Client) doControlPlaneRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, int, error) {
	return c.doRequestWithBaseURL(ctx, method, path, body, c.ControlPlaneURL)
}

func (c *Client) doWorkspaceRequest(ctx context.Context, workspaceID, method, path string, body interface{}) (*http.Response, int, error) {
	// Get or retrieve workspace URL
	workspaceURL, exists := c.workspaceURLs[workspaceID]
	if !exists {
		// Retrieve workspace URL from control plane
		workspace, err := c.GetWorkspace(ctx, workspaceID)
		if err != nil {
			return nil, 0, fmt.Errorf("error retrieving workspace URL: %v", err)
		}
		// In the real API, workspace URL needs to be constructed from ZenML service
		if workspace.ZenMLService.Status != nil && workspace.ZenMLService.Status.ServerURL != nil {
			workspaceURL = *workspace.ZenMLService.Status.ServerURL
		} else {
			return nil, 0, fmt.Errorf("workspace does not have a valid server URL")
		}
		c.workspaceURLs[workspaceID] = workspaceURL
	}

	return c.doRequestWithBaseURL(ctx, method, path, body, workspaceURL)
}

func (c *Client) doRequestWithBaseURL(ctx context.Context, method, path string, body interface{}, baseURL string) (*http.Response, int, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("error marshaling request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, path), bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %v", err)
	}

	var accessToken string
	if baseURL == c.ControlPlaneURL {
		accessToken, err = c.getOAuth2Token(ctx)
	} else {
		accessToken, err = c.getAPIToken(ctx)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("error getting API token: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	tflog.Info(ctx, fmt.Sprintf("[ZENML] Making request: %s %s", method, req.URL.String()))
	if body != nil {
		prettyJSON, _ := json.MarshalIndent(body, "", "  ")
		tflog.Debug(ctx, fmt.Sprintf("[ZENML] Request body (JSON):\n%s", prettyJSON))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error making request: %v", err)
	}

	// Read the response body once and store it in a variable
	defer resp.Body.Close()
	resp_body, _ := io.ReadAll(resp.Body)

	// Print the response body as JSON if available
	if len(resp_body) > 0 {
		var prettyBody map[string]interface{}
		if err := json.Unmarshal(resp_body, &prettyBody); err == nil {
			prettyJSON, _ := json.MarshalIndent(prettyBody, "", "  ")
			tflog.Debug(ctx, fmt.Sprintf("[ZENML] Response body (JSON):\n%s", prettyJSON))
		} else {
			tflog.Debug(ctx, fmt.Sprintf("[ZENML] Response body:\n%s", string(resp_body)))
		}
	}

	tflog.Info(ctx, fmt.Sprintf("[ZENML] Response status: %d", resp.StatusCode))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(resp_body))
	}

	// Re-wrap the body so that the caller can still read it
	resp.Body = io.NopCloser(bytes.NewReader(resp_body))

	return resp, resp.StatusCode, nil
}

// GetControlPlaneInfo fetches control plane info
func (c *Client) GetControlPlaneInfo(ctx context.Context) (*ControlPlaneInfo, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "GET", "/server/info", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ControlPlaneInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding control plane info: %v", err)
	}
	return &result, nil
}

// GetServerInfo fetches server info to determine version and capabilities
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	resp, _, err := c.doRequest(ctx, "GET", "/api/v1/info", nil)
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
func (c *Client) CreateStack(ctx context.Context, stack StackRequest) (*StackResponse, error) {
	endpoint := "/api/v1/stacks"
	resp, _, err := c.doRequest(ctx, "POST", endpoint, stack)
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

func (c *Client) GetStack(ctx context.Context, id string) (*StackResponse, error) {
	resp, status, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/stacks/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the stack is not found
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result StackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateStack(ctx context.Context, id string, stack StackUpdate) (*StackResponse, error) {
	resp, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/stacks/%s", id), stack)
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

func (c *Client) DeleteStack(ctx context.Context, id string) error {
	resp, status, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/stacks/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the stack is not found
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListStacks(ctx context.Context, params *ListParams) (*Page[StackResponse], error) {
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
	resp, _, err := c.doRequest(ctx, "GET", path, nil)
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
func (c *Client) CreateComponent(ctx context.Context, component ComponentRequest) (*ComponentResponse, error) {
	endpoint := "/api/v1/components"
	resp, _, err := c.doRequest(ctx, "POST", endpoint, component)
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

func (c *Client) GetComponent(ctx context.Context, id string) (*ComponentResponse, error) {
	resp, status, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/components/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the component is not found
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result ComponentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateComponent(ctx context.Context, id string, component ComponentUpdate) (*ComponentResponse, error) {
	resp, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/components/%s", id), component)
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

func (c *Client) DeleteComponent(ctx context.Context, id string) error {
	resp, status, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/components/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the component is not found
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListStackComponents(ctx context.Context, params *ListParams) (*Page[ComponentResponse], error) {
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

	path := fmt.Sprintf("/api/v1/components?%s", query.Encode())
	resp, _, err := c.doRequest(ctx, "GET", path, nil)
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
func (c *Client) VerifyServiceConnector(ctx context.Context, connector ServiceConnectorRequest) (*ServiceConnectorResources, error) {
	resp, _, err := c.doRequest(ctx, "POST", "/api/v1/service_connectors/verify", connector)
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

func (c *Client) CreateServiceConnector(ctx context.Context, connector ServiceConnectorRequest) (*ServiceConnectorResponse, error) {
	endpoint := "/api/v1/service_connectors"
	resp, _, err := c.doRequest(ctx, "POST", endpoint, connector)
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

func (c *Client) GetServiceConnector(ctx context.Context, id string) (*ServiceConnectorResponse, error) {
	resp, status, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/service_connectors/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the service connector is not found
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result ServiceConnectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

func (c *Client) UpdateServiceConnector(ctx context.Context, id string, connector ServiceConnectorUpdate) (*ServiceConnectorResponse, error) {
	resp, _, err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/service_connectors/%s", id), connector)
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

func (c *Client) DeleteServiceConnector(ctx context.Context, id string) error {
	resp, status, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/service_connectors/%s", id), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the service connector is not found
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListServiceConnectors(ctx context.Context, params *ListParams) (*Page[ServiceConnectorResponse], error) {
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
	resp, _, err := c.doRequest(ctx, "GET", path, nil)
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
func (c *Client) GetServiceConnectorByName(ctx context.Context, name string) (*ServiceConnectorResponse, error) {
	params := &ListParams{
		Filter: map[string]string{
			"name": name,
		},
	}

	connectors, err := c.ListServiceConnectors(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(connectors.Items) == 0 {
		return nil, nil
	}

	return &connectors.Items[0], nil
}

// Add this new method to the Client
func (c *Client) GetProjectByName(ctx context.Context, name string) (*ProjectResponse, error) {
	resp, status, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/projects/%s", name), nil)
	if err != nil {
		if status == 404 {
			// Return nil if the project is not found
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &result, nil
}

// Add this method to get the current user
func (c *Client) GetCurrentUser(ctx context.Context) (*UserResponse, error) {
	resp, _, err := c.doRequest(ctx, "GET", "/api/v1/current-user", nil)
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

// Workspace operations
func (c *Client) CreateWorkspace(ctx context.Context, workspace WorkspaceRequest) (*WorkspaceResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "POST", "/workspaces", workspace)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result WorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding workspace response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetWorkspace(ctx context.Context, id string) (*WorkspaceResponse, error) {
	resp, status, err := c.doControlPlaneRequest(ctx, "GET", fmt.Sprintf("/workspaces/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result WorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding workspace response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, id string, workspace WorkspaceUpdate) (*WorkspaceResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "PATCH", fmt.Sprintf("/workspaces/%s", id), workspace)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result WorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding workspace response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string) error {
	resp, status, err := c.doControlPlaneRequest(ctx, "DELETE", fmt.Sprintf("/workspaces/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListWorkspaces(ctx context.Context, params *ListParams) (*Page[WorkspaceResponse], error) {
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

	path := fmt.Sprintf("/workspaces?%s", query.Encode())
	resp, _, err := c.doControlPlaneRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page[WorkspaceResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding workspaces response: %v", err)
	}
	return &result, nil
}

// Team operations
func (c *Client) CreateTeam(ctx context.Context, team TeamRequest) (*TeamResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "POST", "/teams", team)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding team response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetTeam(ctx context.Context, id string) (*TeamResponse, error) {
	resp, status, err := c.doControlPlaneRequest(ctx, "GET", fmt.Sprintf("/teams/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding team response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateTeam(ctx context.Context, id string, team TeamUpdate) (*TeamResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "PATCH", fmt.Sprintf("/teams/%s", id), team)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TeamResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding team response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	resp, status, err := c.doControlPlaneRequest(ctx, "DELETE", fmt.Sprintf("/teams/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListTeams(ctx context.Context, params *ListParams) (*Page[TeamResponse], error) {
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

	path := fmt.Sprintf("/teams?%s", query.Encode())
	resp, _, err := c.doControlPlaneRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page[TeamResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding teams response: %v", err)
	}
	return &result, nil
}

// Team member operations
func (c *Client) AddTeamMember(ctx context.Context, teamID, userID string) error {
	body := map[string]string{"user_id": userID}
	resp, _, err := c.doControlPlaneRequest(ctx, "POST", fmt.Sprintf("/teams/%s/members", teamID), body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	// For DELETE with body, we need to structure the request properly
	body := map[string]string{"user_id": userID}
	resp, _, err := c.doControlPlaneRequest(ctx, "DELETE", fmt.Sprintf("/teams/%s/members", teamID), body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListTeamMembers(ctx context.Context, teamID string) ([]TeamMemberResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "GET", fmt.Sprintf("/teams/%s/members", teamID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []TeamMemberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding team members response: %v", err)
	}
	return result, nil
}

// Project operations
func (c *Client) CreateProject(ctx context.Context, project ProjectRequest) (*ProjectResponse, error) {
	resp, _, err := c.doWorkspaceRequest(ctx, project.WorkspaceID, "POST", "/api/v1/projects", project)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding project response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetProject(ctx context.Context, workspaceID, id string) (*ProjectResponse, error) {
	resp, status, err := c.doWorkspaceRequest(ctx, workspaceID, "GET", fmt.Sprintf("/api/v1/projects/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding project response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateProject(ctx context.Context, workspaceID, id string, project ProjectUpdate) (*ProjectResponse, error) {
	resp, _, err := c.doWorkspaceRequest(ctx, workspaceID, "PUT", fmt.Sprintf("/api/v1/projects/%s", id), project)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding project response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteProject(ctx context.Context, workspaceID, id string) error {
	resp, status, err := c.doWorkspaceRequest(ctx, workspaceID, "DELETE", fmt.Sprintf("/api/v1/projects/%s", id), nil)
	if err != nil {
		if status == 404 {
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListProjects(ctx context.Context, workspaceID string, params *ListParams) (*Page[ProjectResponse], error) {
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

	path := fmt.Sprintf("/api/v1/projects?%s", query.Encode())
	resp, _, err := c.doWorkspaceRequest(ctx, workspaceID, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page[ProjectResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding projects response: %v", err)
	}
	return &result, nil
}

// Role assignment operations
func (c *Client) CreateRoleAssignment(ctx context.Context, assignment RoleAssignmentRequest) (*RoleAssignmentResponse, error) {
	// Real API uses /roles/{role_id}/assignments endpoint
	endpoint := fmt.Sprintf("/roles/%s/assignments", assignment.RoleID)
	resp, _, err := c.doControlPlaneRequest(ctx, "POST", endpoint, assignment)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RoleAssignmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding role assignment response: %v", err)
	}
	return &result, nil
}

func (c *Client) GetRoleAssignment(ctx context.Context, roleID, assignmentID string) (*RoleAssignmentResponse, error) {
	resp, status, err := c.doControlPlaneRequest(ctx, "GET", fmt.Sprintf("/roles/%s/assignments/%s", roleID, assignmentID), nil)
	if err != nil {
		if status == 404 {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var result RoleAssignmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding role assignment response: %v", err)
	}
	return &result, nil
}

func (c *Client) UpdateRoleAssignment(ctx context.Context, roleID, assignmentID string, assignment RoleAssignmentUpdate) (*RoleAssignmentResponse, error) {
	resp, _, err := c.doControlPlaneRequest(ctx, "PATCH", fmt.Sprintf("/roles/%s/assignments/%s", roleID, assignmentID), assignment)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RoleAssignmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding role assignment response: %v", err)
	}
	return &result, nil
}

func (c *Client) DeleteRoleAssignment(ctx context.Context, roleID, assignmentID string) error {
	resp, status, err := c.doControlPlaneRequest(ctx, "DELETE", fmt.Sprintf("/roles/%s/assignments/%s", roleID, assignmentID), nil)
	if err != nil {
		if status == 404 {
			return nil
		}
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) ListRoleAssignments(ctx context.Context, roleID string, params *ListParams) (*Page[RoleAssignmentResponse], error) {
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

	path := fmt.Sprintf("/roles/%s/assignments?%s", roleID, query.Encode())
	resp, _, err := c.doControlPlaneRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Page[RoleAssignmentResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding role assignments response: %v", err)
	}
	return &result, nil
}
