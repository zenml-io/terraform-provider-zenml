package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ListParams struct {
	Page     int
	PageSize int
	Filter   map[string]string
}

func normalizeListParams(params *ListParams) *ListParams {
	if params == nil {
		return &ListParams{Page: 1, PageSize: 100}
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 100
	}
	return params
}

type Client struct {
	ServerURL       string
	APIKey          string
	APIToken        string
	APITokenExpires *time.Time
	HTTPClient      *http.Client
	tokenMu         sync.Mutex
}

func NewClient(serverURL, apiKey string, apiToken string) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 4
	retryClient.Logger = nil

	httpClient := retryClient.StandardClient()
	// No hard client timeout — context deadlines from Terraform resource
	// timeouts handle cancellation. A fixed timeout here would conflict
	// with retryablehttp backoff and risk killing long-running operations
	// (e.g. service connector verification) prematurely.

	return &Client{
		ServerURL:       strings.TrimRight(serverURL, "/"),
		APIKey:          apiKey,
		APIToken:        apiToken,
		APITokenExpires: nil,
		HTTPClient:      httpClient,
	}
}

// invalidateToken atomically clears the cached token so the next call to
// getAPIToken will re-authenticate using the API key.
func (c *Client) invalidateToken() {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.APIToken = ""
	c.APITokenExpires = nil
}

func (c *Client) getAPIToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

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
	loginReq, err := http.NewRequestWithContext(
		ctx,
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

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		return "", fmt.Errorf("authentication failed: login request returned status %d: %s", loginResp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("error decoding login response: %v", err)
	}

	c.APIToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn <= 0 {
		// Server returned no expiry or an invalid value — treat the token
		// as non-expiring so we don't force a refresh on every request.
		c.APITokenExpires = nil
	} else {
		// Set the expiry time with a buffer before the actual expiry, to account
		// for clock skew and to avoid using an expired token when making requests.
		// Clamp the buffer so short-lived tokens don't get a negative expiry.
		buffer := 300
		if tokenResp.ExpiresIn < buffer {
			buffer = tokenResp.ExpiresIn / 2
		}
		expiresAt := time.Now().Add(
			time.Duration(tokenResp.ExpiresIn-buffer) * time.Second,
		)
		c.APITokenExpires = &expiresAt
	}

	return c.APIToken, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, int, error) {
	// Marshal body once — reused if we need to retry on 401.
	var jsonBody []byte
	if body != nil {
		var err error
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("error marshaling request body: %v", err)
		}
	}

	// Generate idempotency key once for POST requests — reused across retries
	// so the server deduplicates transport-level and auth-refresh retries alike.
	//
	// NOTE: This relies on the ZenML server having request_deduplication enabled
	// (the default). If a server admin disables it, retried POST requests could
	// create duplicate resources. See the ZenML RequestManager for details.
	var idempotencyKey string
	if method == "POST" {
		idempotencyKey = uuid.NewString()
	}

	url := fmt.Sprintf("%s%s", c.ServerURL, path)
	tflog.Info(ctx, fmt.Sprintf("[ZENML] Making request: %s %s", method, url))
	if jsonBody != nil {
		var indented bytes.Buffer
		if json.Indent(&indented, jsonBody, "", "  ") == nil {
			tflog.Debug(ctx, fmt.Sprintf("[ZENML] Request body (JSON):\n%s", indented.String()))
		}
	}

	// Allow one retry on 401 when an API key is available for re-authentication.
	// Note: retryablehttp handles transport-level retries (connection errors, 5xx,
	// 429) transparently. This loop only retries on 401 (auth expiry), which
	// retryablehttp does not retry. The two layers are complementary, not multiplicative.
	maxAttempts := 1
	if c.APIKey != "" {
		maxAttempts = 2
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var bodyReader io.Reader
		if jsonBody != nil {
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, 0, fmt.Errorf("error creating request: %v", err)
		}

		accessToken, err := c.getAPIToken(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("error getting API token: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		if jsonBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("error making request: %v", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, 0, fmt.Errorf("error reading response body: %v", err)
		}

		if len(respBody) > 0 {
			var prettyBody map[string]interface{}
			if err := json.Unmarshal(respBody, &prettyBody); err == nil {
				prettyJSON, _ := json.MarshalIndent(prettyBody, "", "  ")
				tflog.Debug(ctx, fmt.Sprintf("[ZENML] Response body (JSON):\n%s", prettyJSON))
			} else {
				tflog.Debug(ctx, fmt.Sprintf("[ZENML] Response body:\n%s", string(respBody)))
			}
		}

		tflog.Info(ctx, fmt.Sprintf("[ZENML] Response status: %d", resp.StatusCode))

		// On 401, invalidate the cached token and retry with fresh credentials.
		if resp.StatusCode == http.StatusUnauthorized && attempt < maxAttempts-1 {
			tflog.Info(ctx, "[ZENML] Got 401, refreshing token and retrying")
			c.invalidateToken()
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, resp.StatusCode, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		// Re-wrap the body so that the caller can still read it
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		return resp, resp.StatusCode, nil
	}

	// Unreachable in practice, but satisfies the compiler.
	return nil, 0, fmt.Errorf("exhausted request attempts")
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
	params = normalizeListParams(params)
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

// ListStacksByComponent returns all stacks that use the given component.
func (c *Client) ListStacksByComponent(
	ctx context.Context,
	componentID string,
) ([]StackResponse, error) {
	params := &ListParams{
		Filter: map[string]string{
			"component_id": componentID,
		},
	}
	page, err := c.ListStacks(ctx, params)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// RemoveComponentFromStack removes a specific component by ID from a stack.
// Since stacks can have multiple components of the same type, this filters
// out only the component with the matching ID.
func (c *Client) RemoveComponentFromStack(
	ctx context.Context,
	stackID string,
	componentID string,
) error {
	stack, err := c.GetStack(ctx, stackID)
	if err != nil {
		return err
	}
	if stack == nil {
		return fmt.Errorf("stack %s not found", stackID)
	}

	newComponents := make(map[string][]string)
	for compType, compList := range stack.Metadata.Components {
		var filteredIDs []string
		for _, comp := range compList {
			if comp.ID != componentID {
				filteredIDs = append(filteredIDs, comp.ID)
			}
		}
		if len(filteredIDs) > 0 {
			newComponents[compType] = filteredIDs
		}
	}

	labels := make(map[string]string)
	if stack.Metadata != nil && stack.Metadata.Labels != nil {
		labels = stack.Metadata.Labels
	}

	update := StackUpdate{
		Name:       stack.Name,
		Components: newComponents,
		Labels:     labels,
	}

	_, err = c.UpdateStack(ctx, stackID, update)
	return err
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
	params = normalizeListParams(params)
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
	params = normalizeListParams(params)
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
