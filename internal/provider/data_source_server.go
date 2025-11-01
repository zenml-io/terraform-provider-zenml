package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

type ServerDataSource struct {
	client *Client
}

type ServerDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Version             types.String `tfsdk:"version"`
	DeploymentType      types.String `tfsdk:"deployment_type"`
	AuthScheme          types.String `tfsdk:"auth_scheme"`
	ServerURL           types.String `tfsdk:"server_url"`
	DashboardURL        types.String `tfsdk:"dashboard_url"`
	ProDashboardURL     types.String `tfsdk:"pro_dashboard_url"`
	ProAPIURL           types.String `tfsdk:"pro_api_url"`
	ProOrganizationID   types.String `tfsdk:"pro_organization_id"`
	ProOrganizationName types.String `tfsdk:"pro_organization_name"`
	ProWorkspaceID      types.String `tfsdk:"pro_workspace_id"`
	ProWorkspaceName    types.String `tfsdk:"pro_workspace_name"`
	Metadata            types.Map    `tfsdk:"metadata"`
}

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for global ZenML server information",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Server version",
				Computed:            true,
			},
			"deployment_type": schema.StringAttribute{
				MarkdownDescription: "Server deployment type",
				Computed:            true,
			},
			"auth_scheme": schema.StringAttribute{
				MarkdownDescription: "Server authentication scheme",
				Computed:            true,
			},
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Server API URL",
				Computed:            true,
			},
			"dashboard_url": schema.StringAttribute{
				MarkdownDescription: "Server dashboard URL",
				Computed:            true,
			},
			"pro_dashboard_url": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro dashboard URL",
				Computed:            true,
			},
			"pro_api_url": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro API URL",
				Computed:            true,
			},
			"pro_organization_id": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro organization ID",
				Computed:            true,
			},
			"pro_organization_name": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro organization name",
				Computed:            true,
			},
			"pro_workspace_id": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro workspace ID",
				Computed:            true,
			},
			"pro_workspace_name": schema.StringAttribute{
				MarkdownDescription: "ZenML Pro workspace name",
				Computed:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: "Server metadata",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading server information")

	serverInfo, err := d.client.GetServerInfo(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server info, got error: %s", err))
		return
	}

	data.ID = types.StringValue(serverInfo.ID)
	data.Name = types.StringValue(serverInfo.Name)
	data.Version = types.StringValue(serverInfo.Version)
	data.DeploymentType = types.StringValue(serverInfo.DeploymentType)
	data.AuthScheme = types.StringValue(serverInfo.AuthScheme)
	data.ServerURL = types.StringValue(serverInfo.ServerURL)
	data.DashboardURL = types.StringValue(serverInfo.DashboardURL)

	if serverInfo.ProDashboardURL != nil {
		data.ProDashboardURL = types.StringValue(*serverInfo.ProDashboardURL)
	} else {
		data.ProDashboardURL = types.StringNull()
	}

	if serverInfo.ProAPIURL != nil {
		data.ProAPIURL = types.StringValue(*serverInfo.ProAPIURL)
	} else {
		data.ProAPIURL = types.StringNull()
	}

	if serverInfo.ProOrganizationID != nil {
		data.ProOrganizationID = types.StringValue(*serverInfo.ProOrganizationID)
	} else {
		data.ProOrganizationID = types.StringNull()
	}

	if serverInfo.ProOrganizationName != nil {
		data.ProOrganizationName = types.StringValue(*serverInfo.ProOrganizationName)
	} else {
		data.ProOrganizationName = types.StringNull()
	}

	if serverInfo.ProWorkspaceID != nil {
		data.ProWorkspaceID = types.StringValue(*serverInfo.ProWorkspaceID)
	} else {
		data.ProWorkspaceID = types.StringNull()
	}

	if serverInfo.ProWorkspaceName != nil {
		data.ProWorkspaceName = types.StringValue(*serverInfo.ProWorkspaceName)
	} else {
		data.ProWorkspaceName = types.StringNull()
	}

	if serverInfo.Metadata != nil {
		metadataMap := make(map[string]attr.Value)
		for k, v := range serverInfo.Metadata {
			metadataMap[k] = types.StringValue(v)
		}
		metadataValue, diags := types.MapValue(types.StringType, metadataMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Metadata = metadataValue
		}
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
