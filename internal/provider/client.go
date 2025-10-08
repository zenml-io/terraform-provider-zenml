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
	ServerURL       string
	APIKey          string
	APIToken        string
	APITokenExpires *time.Time
	HTTPClient      *http.Client
}

func NewClient(serverURL, apiKey string, apiToken string) *Client {
	return &Client{
		ServerURL:       serverURL,
		APIKey:          apiKey,
		APIToken:        apiToken,
		APITokenExpires: nil,
		HTTPClient:      &http.Client{},
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

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, int, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("error marshaling request body: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.ServerURL, path), bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %v", err)
	}

	accessToken, err := c.getAPIToken(ctx)

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

// Project operations...
func (c *Client) CreateProject(ctx context.Context, project ProjectRequest) (*ProjectResponse, error) {
	endpoint := "/api/v1/projects"
	resp, _, err := c.doRequest(ctx, "POST", endpoint, project)
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

func (c *Client) GetProject(ctx context.Context, nameOrID string) (*ProjectResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/projects/%s", nameOrID)
	resp, status, err := c.doRequest(ctx, "GET", endpoint, nil)
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

func (c *Client) UpdateProject(ctx context.Context, nameOrID string, project ProjectUpdate) (*ProjectResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/projects/%s", nameOrID)
	resp, _, err := c.doRequest(ctx, "PUT", endpoint, project)
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

func (c *Client) DeleteProject(ctx context.Context, nameOrID string) error {
	endpoint := fmt.Sprintf("/api/v1/projects/%s", nameOrID)
	resp, status, err := c.doRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		if status == 404 {
			return nil
		}
		return err
	}
	defer resp.Body.Close()
	return nil
}
