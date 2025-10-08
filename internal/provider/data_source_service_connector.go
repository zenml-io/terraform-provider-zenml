package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ServiceConnectorDataSource{}

func NewServiceConnectorDataSource() datasource.DataSource {
	return &ServiceConnectorDataSource{}
}

type ServiceConnectorDataSource struct {
	client *Client
}

type ServiceConnectorDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	AuthMethod    types.String `tfsdk:"auth_method"`
	ResourceType  types.String `tfsdk:"resource_type"`
	ResourceID    types.String `tfsdk:"resource_id"`
	Configuration types.Map    `tfsdk:"configuration"`
	Labels        types.Map    `tfsdk:"labels"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	Created       types.String `tfsdk:"created"`
	Updated       types.String `tfsdk:"updated"`
}

func (d *ServiceConnectorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_connector"
}

func (d *ServiceConnectorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for ZenML service connectors",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the service connector",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service connector",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the service connector",
				Computed:            true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method of the service connector",
				Computed:            true,
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Resource type associated with the service connector",
				Computed:            true,
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID for the service connector",
				Computed:            true,
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "Configuration for the service connector",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the service connector",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Expiration time for the service connector",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service connector was created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service connector was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *ServiceConnectorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceConnectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServiceConnectorDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service connector information")

	var connector *ServiceConnectorResponse
	var err error

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		connector, err = d.client.GetServiceConnector(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() && data.Name.ValueString() != "" {
		connector, err = d.client.GetServiceConnectorByName(ctx, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to identify the service connector",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service connector, got error: %s", err))
		return
	}

	if connector == nil {
		resp.Diagnostics.AddError(
			"Service Connector Not Found",
			"No service connector found with the specified criteria",
		)
		return
	}

	data.ID = types.StringValue(connector.ID)
	data.Name = types.StringValue(connector.Name)

	if connector.Body != nil {
		// Handle connector type (can be string or object)
		var connectorType string
		if err := json.Unmarshal(connector.Body.ConnectorType, &connectorType); err != nil {
			// If it's not a string, try to extract from object
			var connectorTypeObj struct {
				ConnectorType string `json:"connector_type"`
			}
			if err := json.Unmarshal(connector.Body.ConnectorType, &connectorTypeObj); err == nil {
				connectorType = connectorTypeObj.ConnectorType
			}
		}
		if connectorType != "" {
			data.Type = types.StringValue(connectorType)
		}

		data.AuthMethod = types.StringValue(connector.Body.AuthMethod)

		// If there are multiple resource types, leave the resource_type field empty
		if len(connector.Body.ResourceTypes) == 1 {
			data.ResourceType = types.StringValue(connector.Body.ResourceTypes[0])
		} else {
			data.ResourceType = types.StringNull()
		}

		if connector.Body.ResourceID != nil {
			data.ResourceID = types.StringValue(*connector.Body.ResourceID)
		} else {
			data.ResourceID = types.StringNull()
		}

		if connector.Body.ExpiresAt != nil {
			data.ExpiresAt = types.StringValue(*connector.Body.ExpiresAt)
		} else {
			data.ExpiresAt = types.StringNull()
		}

		data.Created = types.StringValue(connector.Body.Created)
		data.Updated = types.StringValue(connector.Body.Updated)
	}

	if connector.Metadata != nil && connector.Metadata.Configuration != nil {
		configMap := make(map[string]attr.Value)
		for k, v := range connector.Metadata.Configuration {
			// Always convert to string representation
			// If it's already a string, use it as-is; otherwise convert to string
			switch val := v.(type) {
			case string:
				configMap[k] = types.StringValue(val)
			default:
				configMap[k] = types.StringValue(fmt.Sprintf("%v", val))
			}
		}
		configValue, diags := types.MapValue(types.StringType, configMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Configuration = configValue
		}
	} else {
		data.Configuration = types.MapNull(types.StringType)
	}

	if connector.Metadata != nil && connector.Metadata.Labels != nil {
		labelMap := make(map[string]attr.Value)
		for k, v := range connector.Metadata.Labels {
			labelMap[k] = types.StringValue(v)
		}
		labelValue, diags := types.MapValue(types.StringType, labelMap)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Labels = labelValue
		}
	} else {
		data.Labels = types.MapNull(types.StringType)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
