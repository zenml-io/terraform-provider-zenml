package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &StackComponentDataSource{}

func NewStackComponentDataSource() datasource.DataSource {
	return &StackComponentDataSource{}
}

type StackComponentDataSource struct {
	client *Client
}

type StackComponentDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Flavor              types.String `tfsdk:"flavor"`
	Configuration       types.Map    `tfsdk:"configuration"`
	ConnectorID         types.String `tfsdk:"connector_id"`
	ConnectorResourceID types.String `tfsdk:"connector_resource_id"`
	Labels              types.Map    `tfsdk:"labels"`
	Created             types.String `tfsdk:"created"`
	Updated             types.String `tfsdk:"updated"`
}

func (d *StackComponentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_component"
}

func (d *StackComponentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for ZenML stack components",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the stack component",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stack component",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the stack component",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(validComponentTypes...),
				},
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Flavor of the stack component",
				Computed:            true,
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "Configuration for the stack component",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"connector_id": schema.StringAttribute{
				MarkdownDescription: "ID of the service connector used by this component",
				Computed:            true,
			},
			"connector_resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID used from the service connector",
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the stack component",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the stack component was created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the stack component was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *StackComponentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StackComponentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StackComponentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading stack component information")

	var component *ComponentResponse
	var err error

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		component, err = d.client.GetComponent(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() && data.Name.ValueString() != "" {
		params := &ListParams{
			Filter: map[string]string{
				"name": data.Name.ValueString(),
			},
		}
		if !data.Type.IsNull() && data.Type.ValueString() != "" {
			params.Filter["type"] = data.Type.ValueString()
		}

		components, err := d.client.ListStackComponents(ctx, params)
		if err == nil && len(components.Items) > 0 {
			component = &components.Items[0]
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to identify the stack component",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack component, got error: %s", err))
		return
	}

	if component == nil {
		resp.Diagnostics.AddError(
			"Stack Component Not Found",
			"No stack component found with the specified criteria",
		)
		return
	}

	data.ID = types.StringValue(component.ID)
	data.Name = types.StringValue(component.Name)

	if component.Body != nil {
		data.Type = types.StringValue(component.Body.Type)
		data.Flavor = types.StringValue(component.Body.Flavor)
		data.Created = types.StringValue(component.Body.Created)
		data.Updated = types.StringValue(component.Body.Updated)
	}

	if component.Metadata != nil && component.Metadata.Configuration != nil {
		configMap := make(map[string]attr.Value)
		for k, v := range component.Metadata.Configuration {
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

	if component.Metadata != nil && component.Metadata.Labels != nil {
		labelMap := make(map[string]attr.Value)
		for k, v := range component.Metadata.Labels {
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

	if component.Metadata != nil {
		if component.Metadata.Connector != nil {
			data.ConnectorID = types.StringValue(component.Metadata.Connector.ID)
		} else {
			data.ConnectorID = types.StringNull()
		}
		if component.Metadata.ConnectorResourceID != nil {
			data.ConnectorResourceID = types.StringValue(*component.Metadata.ConnectorResourceID)
		} else {
			data.ConnectorResourceID = types.StringNull()
		}
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
