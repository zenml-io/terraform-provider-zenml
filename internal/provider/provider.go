// provider.go
package provider

import (
	"context"

	"github.com/hashicorp/go-version"
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
			"skip_version_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip the ZenML server version compatibility check. Use with caution as it may lead to unexpected behavior.",
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

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	serverURL := d.Get("server_url").(string)
	apiKey := d.Get("api_key").(string)
	apiToken := d.Get("api_token").(string)
	skipVersionCheck := d.Get("skip_version_check").(bool)

	// Should be handled by the schema
	if serverURL == "" {
		return nil, diag.Errorf("server_url must be configured")
	}
	if apiKey == "" && apiToken == "" {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "An API key or an API token must be configured for the ZenML Terraform provider to be able to authenticate with your ZenML server.",
				Detail: `
It is recommended to use an API key for long-term Terraform management operations, as API tokens expire after a short period of time.

More information on how to configure a service account and an API key can be found at https://docs.zenml.io/how-to/connecting-to-zenml/connect-with-a-service-account.

To configure the ZenML Terraform provider with an API key, add the following block to your Terraform configuration:

provider "zenml" {
	server_url = "https://example.zenml.io"
	api_key   = "your api key"
}

or use the ZENML_API_KEY environment variable to set the API key.`,
			},
		}
	}

	client := NewClient(serverURL, apiKey, apiToken)
	if client == nil {
		return nil, diag.Errorf("failed to create client")
	}

	// Test the client connection
	serverInfo, err := client.GetServerInfo(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get server info: %v", err)
	}

	if !skipVersionCheck {
		serverVersion, err := version.NewVersion(serverInfo.Version)
		if err != nil {
			return nil, diag.Errorf("Failed to parse server version: %s", err)
		}

		constraintStr := ">= 0.80.0"
		constraint, err := version.NewConstraint(constraintStr)
		if err != nil {
			return nil, diag.Errorf("Failed to parse version constraint: %s", err)
		}

		if !constraint.Check(serverVersion) {
			return nil, diag.Errorf(
				"ZenML server version must be at least 0.80.0 to use the current Terraform provider version (current version: %s)\n\n"+
					"To resolve this:\n\n"+
					"1. Upgrade your ZenML server to version 0.80.0 or higher, or\n"+
					"2. Add a provider version constraint to your Terraform configuration:\n\n"+
					"    terraform {\n"+
					"      required_providers {\n"+
					"        zenml = {\n"+
					"          source  = \"zenml/zenml\"\n"+
					"          version = \"< 2.0.0\"\n"+
					"        }\n"+
					"      }\n"+
					"    }\n\n"+
					"3. Use the skip_version_check attribute to skip this version check:\n\n"+
					"    provider \"zenml\" {\n"+
					"      skip_version_check = true\n"+
					"    }",
				serverInfo.Version,
			)
		}
	}

	return client, diags
}
