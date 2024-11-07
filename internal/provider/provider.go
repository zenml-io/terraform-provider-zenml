// provider.go
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_TOKEN", nil),
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

	// Should be handled by the schema
	if serverURL == "" {
		return nil, diag.Errorf("server_url must be configured")
	}
	if apiKey == "" && apiToken == "" {
		return nil, diag.Errorf(`an API key or an API token must be configured for the ZenML Terraform provider to be able to authenticate with your ZenML server.

		It is recommended to use an API key for long-term Terraform management operations, as API tokens expire after a short period of time.
		
		More information on how to configure a service account and an API key can be found at https://docs.zenml.io/how-to/connecting-to-zenml/connect-with-a-service-account.
		
		To configure the ZenML Terraform provider with an API key, add the following block to your Terraform configuration:
		
		provider "zenml" {
			server_url = "https://example.zenml.io"
			api_key   = "your api key"
		}
		
		or use the ZENML_API_KEY environment variable to set the API key.
		`)

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
