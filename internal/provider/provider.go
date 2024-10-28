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
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_KEY", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenml_stack":             resourceStack(),
			"zenml_stack_component":   resourceStackComponent(),
			"zenml_service_connector": resourceServiceConnector(),
		},
		DataSourcesMap: map[string]*schema.Resource{
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

	if serverURL == "" {
		return nil, diag.Errorf("server_url cannot be empty")
	}
	if apiKey == "" {
		return nil, diag.Errorf("api_key cannot be empty")
	}

	client := NewClient(serverURL, apiKey)
	if client == nil {
		return nil, diag.Errorf("failed to create client")
	}

	// Test the client connection
	// You might want to add a simple API call here to verify the connection
	client.GetServerInfo()

	return client, diags
}
