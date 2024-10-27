package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	ServerURL string
	APIKey    string
	HTTPClient *http.Client
}

func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

type Stack struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Components  map[string]Component  `json:"components"`
}

type Component struct {
	ID                 string                 `json:"id,omitempty"`
	Name               string                 `json:"name"`
	Type               string                 `json:"type"`
	Flavor             string                 `json:"flavor"`
	Configuration      map[string]interface{} `json:"configuration"`
	ServiceConnectorID string                 `json:"service_connector_id,omitempty"`
}

func (c *Client) CreateStack(stack Stack) (string, error) {
	body, err := json.Marshal(stack)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/stacks", c.ServerURL), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create stack: %s", resp.Status)
	}

	var result Stack
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}