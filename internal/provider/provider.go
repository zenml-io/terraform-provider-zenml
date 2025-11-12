// provider.go
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &ZenMLProvider{}
var _ provider.ProviderWithFunctions = &ZenMLProvider{}

type ZenMLProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ZenMLProviderModel struct {
	ServerURL        types.String `tfsdk:"server_url"`
	APIKey           types.String `tfsdk:"api_key"`
	APIToken         types.String `tfsdk:"api_token"`
	SkipVersionCheck types.Bool   `tfsdk:"skip_version_check"`
}

func (p *ZenMLProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zenml"
	resp.Version = p.version
}

func (p *ZenMLProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "The URL of the ZenML server",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for authentication with the ZenML server",
				Optional:            true,
				Sensitive:           true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The API token for authentication with the ZenML server",
				Optional:            true,
				Sensitive:           true,
			},
			"skip_version_check": schema.BoolAttribute{
				MarkdownDescription: "Skip the ZenML server version compatibility check. Use with caution as it may lead to unexpected behavior.",
				Optional:            true,
			},
		},
	}
}

func (p *ZenMLProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ZenMLProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serverURL := data.ServerURL.ValueString()
	apiKey := data.APIKey.ValueString()
	apiToken := data.APIToken.ValueString()
	skipVersionCheck := data.SkipVersionCheck.ValueBool()

	if data.ServerURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Unknown ZenML Server URL",
			"The provider cannot create the ZenML API client as there is an "+
				"unknown configuration value for the ZenML server URL. Either "+
				"target apply the source of the value first, set the value "+
				"statically in the configuration, or use the ZENML_SERVER_URL "+
				"environment variable.",
		)
	}

	if data.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown ZenML API Key",
			"The provider cannot create the ZenML API client as there is an "+
				"unknown configuration value for the ZenML API key. Either "+
				"target apply the source of the value first, set the value "+
				"statically in the configuration, or use the ZENML_API_KEY "+
				"environment variable.",
		)
	}

	if data.APIToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown ZenML API Token",
			"The provider cannot create the ZenML API client as "+
				"there is an unknown configuration value for the "+
				"ZenML API token. Either target apply the source "+
				"of the value first, set the value statically in "+
				"the configuration, or use the ZENML_API_TOKEN "+
				"environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	if serverURL == "" {
		serverURL = os.Getenv("ZENML_SERVER_URL")
	}

	if apiKey == "" {
		apiKey = os.Getenv("ZENML_API_KEY")
	}

	if apiToken == "" {
		apiToken = os.Getenv("ZENML_API_TOKEN")
	}

	if serverURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Missing ZenML Server URL",
			"The provider requires a ZenML server URL. Set the "+
				"server_url value in the configuration or use the "+
				"ZENML_SERVER_URL environment variable. If either is already "+
				"set, ensure the value is not empty.",
		)
	}

	if apiKey == "" && apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing ZenML API Credentials",
			"An API key or an API token must be configured for the ZenML "+
				"Terraform provider to be able to authenticate with your "+
				"ZenML server.\n\n"+
				"It is recommended to use an API key for long-term Terraform "+
				"management operations, as API tokens expire after a short "+
				"period of time.\n\n"+
				"More information on how to configure a service account and "+
				"an API key can be found at "+
				"https://docs.zenml.io/how-to/connecting-to-zenml/"+
				"connect-with-a-service-account.\n\n"+
				"To configure the ZenML Terraform provider with an API key, "+
				"add the following block to your Terraform configuration:\n\n"+
				"provider \"zenml\" {\n"+
				"  server_url = \"https://example.zenml.io\"\n"+
				"  api_key   = \"your api key\"\n"+
				"}\n\n"+
				"or use the ZENML_API_KEY environment variable to set the "+
				"API key.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "zenml_server_url", serverURL)
	ctx = tflog.SetField(ctx, "zenml_api_key", apiKey)
	ctx = tflog.SetField(ctx, "zenml_api_token", apiToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "zenml_api_key")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "zenml_api_token")

	tflog.Debug(ctx, "Creating ZenML client")

	client := NewClient(serverURL, apiKey, apiToken)
	if client == nil {
		resp.Diagnostics.AddError(
			"Unable to Create ZenML API Client",
			"An unexpected error occurred when creating the ZenML API "+
				"client. If the error is not clear, please contact the "+
				"provider developers.\n\n"+
				"ZenML Client Error: failed to create client",
		)
		return
	}

	// Test the client connection
	serverInfo, err := client.GetServerInfo(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to ZenML Server",
			"An unexpected error occurred when connecting to the ZenML "+
				"server. Please verify that the server URL and credentials "+
				"are correct.\n\n"+
				"ZenML Client Error: "+err.Error(),
		)
		return
	}

	if !skipVersionCheck {
		serverVersion, err := version.NewVersion(serverInfo.Version)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid ZenML Server Version",
				"Failed to parse server version: "+err.Error(),
			)
			return
		}

		constraintStr := ">= 0.80.0"
		constraint, err := version.NewConstraint(constraintStr)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Version Constraint",
				"Failed to parse version constraint: "+err.Error(),
			)
			return
		}

		if !constraint.Check(serverVersion) {
			resp.Diagnostics.AddError(
				"Incompatible ZenML Server Version",
				"ZenML server version must be at least 0.80.0 to use the "+
					"current Terraform provider version (current version: "+
					serverInfo.Version+")\n\n"+
					"To resolve this:\n\n"+
					"1. Upgrade your ZenML server to version 0.80.0 or "+
					"higher, or\n"+
					"2. Add a provider version constraint to your Terraform "+
					"configuration:\n\n"+
					"    terraform {\n"+
					"      required_providers {\n"+
					"        zenml = {\n"+
					"          source  = \"zenml/zenml\"\n"+
					"          version = \"< 2.0.0\"\n"+
					"        }\n"+
					"      }\n"+
					"    }\n\n"+
					"3. Use the skip_version_check attribute to skip this "+
					"version check:\n\n"+
					"    provider \"zenml\" {\n"+
					"      skip_version_check = true\n"+
					"    }",
			)
			return
		}
	}

	// Make the ZenML client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured ZenML client", map[string]any{"success": true})
}

func (p *ZenMLProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStackResource,
		NewStackComponentResource,
		NewServiceConnectorResource,
		NewProjectResource,
	}
}

func (p *ZenMLProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
		NewStackDataSource,
		NewStackComponentDataSource,
		NewServiceConnectorDataSource,
	}
}

func (p *ZenMLProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// No functions are implemented yet
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZenMLProvider{
			version: version,
		}
	}
}
