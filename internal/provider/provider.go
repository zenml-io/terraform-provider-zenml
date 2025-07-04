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
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_SERVER_URL", nil),
				Description: "The URL of the ZenML server (workspace URL). Only required if control_plane_url is not set.",
			},
			"control_plane_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_CONTROL_PLANE_URL", "https://cloudapi.zenml.io"),
				Description: "The URL of the ZenML control plane. Required for Pro features.",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_KEY", nil),
				Description: "API key for authentication. Recommended for long-term operations.",
			},
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_API_TOKEN", nil),
				Description: "API token for authentication. Expires after a short period.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_CLIENT_ID", nil),
				Description: "OAuth2 client ID for ZenML Pro authentication.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZENML_CLIENT_SECRET", nil),
				Description: "OAuth2 client secret for ZenML Pro authentication.",
			},
			"skip_version_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip the ZenML server version compatibility check. Use with caution as it may lead to unexpected behavior.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenml_stack":                     resourceStack(),
			"zenml_stack_component":           resourceStackComponent(),
			"zenml_service_connector":         resourceServiceConnector(),
			"zenml_workspace":                 resourceWorkspace(),
			"zenml_team":                      resourceTeam(),
			"zenml_project":                   resourceProject(),
			"zenml_workspace_role_assignment": resourceWorkspaceRoleAssignment(),
			"zenml_project_role_assignment":   resourceProjectRoleAssignment(),
			"zenml_stack_role_assignment":     resourceStackRoleAssignment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zenml_server":            dataSourceServer(),
			"zenml_stack":             dataSourceStack(),
			"zenml_stack_component":   dataSourceStackComponent(),
			"zenml_service_connector": dataSourceServiceConnector(),
			"zenml_workspace":         dataSourceWorkspace(),
			"zenml_team":              dataSourceTeam(),
			"zenml_project":           dataSourceProject(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	serverURL := d.Get("server_url").(string)
	controlPlaneURL := d.Get("control_plane_url").(string)
	apiKey := d.Get("api_key").(string)
	apiToken := d.Get("api_token").(string)
	clientID := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	skipVersionCheck := d.Get("skip_version_check").(bool)

	// Validate configuration
	if serverURL == "" && controlPlaneURL == "" {
		return nil, diag.Errorf("either server_url or control_plane_url must be configured")
	}

	if apiKey == "" && apiToken == "" && (clientID == "" || clientSecret == "") {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Authentication must be configured for the ZenML Terraform provider",
				Detail: `
For workspace operations, use api_key or api_token.
For control plane operations (Pro features), use client_id and client_secret.

More information on authentication can be found at https://docs.zenml.io/how-to/connecting-to-zenml/connect-with-a-service-account.

Example configuration:

provider "zenml" {
  control_plane_url = "https://cloudapi.zenml.io"
  client_id = "your oauth2 client id"
  client_secret = "your oauth2 client secret"
}
`,
			},
		}
	}

	client := NewClient(serverURL, controlPlaneURL, apiKey, apiToken, clientID, clientSecret)
	if client == nil {
		return nil, diag.Errorf("failed to create client")
	}

	// Test the client connection
	if serverURL != "" {
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
	}

	// Test control plane connection if configured
	if controlPlaneURL != "" {
		_, err := client.GetControlPlaneInfo(ctx)
		if err != nil {
			return nil, diag.Errorf("failed to connect to control plane: %v", err)
		}
	}

	return client, diags
}
