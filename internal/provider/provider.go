// provider.go
package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_SERVER_URL", nil),
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_KEY", nil),
				ExactlyOneOf: []string{"api_key", "api_token"},
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_TOKEN", nil),
				ExactlyOneOf: []string{"api_key", "api_token"},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenml_stack":             resourceStack(),
			"zenml_stack_component":   resourceStackComponent(),
			"zenml_service_connector": resourceServiceConnector(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zenml_server":            dataSourceServer(),
			"zenml_stack":             dataSourceStack(),
			"zenml_stack_component":   dataSourceStackComponent(),
			"zenml_service_connector": dataSourceServiceConnector(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	serverURL := d.Get("server_url").(string)
	apiKey := d.Get("api_key").(string)
	apiToken := d.Get("api_token").(string)

	if serverURL == "" {
		return nil, diag.Errorf("server_url cannot be empty")
	}
	if apiKey == "" && apiToken == "" {
		return nil, diag.Errorf("api_key and api_token cannot both be empty")
	}

	client := NewClient(serverURL, apiKey, apiToken)
	if client == nil {
		return nil, diag.Errorf("failed to create client")
	}

	// Test the client connection
	// You might want to add a simple API call here to verify the connection
	client.GetServerInfo()

	return client, diags
}
