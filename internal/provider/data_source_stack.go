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

var _ datasource.DataSource = &StackDataSource{}

func NewStackDataSource() datasource.DataSource {
	return &StackDataSource{}
}

type StackDataSource struct {
	client *Client
}

type StackDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Components types.List   `tfsdk:"components"`
	Labels     types.Map    `tfsdk:"labels"`
	Created    types.String `tfsdk:"created"`
	Updated    types.String `tfsdk:"updated"`
}

type StackComponentModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Flavor types.String `tfsdk:"flavor"`
}

func (d *StackDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (d *StackDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for ZenML stacks",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the stack",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stack",
				Optional:            true,
			},
			"components": schema.ListNestedAttribute{
				MarkdownDescription: "Components configured in the stack",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Component ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Component name",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Component type",
							Computed:            true,
						},
						"flavor": schema.StringAttribute{
							MarkdownDescription: "Component flavor",
							Computed:            true,
						},
					},
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the stack",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the stack was created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the stack was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *StackDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StackDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading stack information")

	var stack *StackResponse
	var err error

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		stack, err = d.client.GetStack(ctx, data.ID.ValueString())
	} else if !data.Name.IsNull() && data.Name.ValueString() != "" {
		params := &ListParams{
			Filter: map[string]string{
				"name": data.Name.ValueString(),
			},
		}
		stacks, err := d.client.ListStacks(ctx, params)
		if err == nil && len(stacks.Items) > 0 {
			stack = &stacks.Items[0]
		}
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to identify the stack",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack, got error: %s", err))
		return
	}

	if stack == nil {
		resp.Diagnostics.AddError(
			"Stack Not Found",
			"No stack found with the specified criteria",
		)
		return
	}

	data.ID = types.StringValue(stack.ID)
	data.Name = types.StringValue(stack.Name)

	if stack.Body != nil {
		data.Created = types.StringValue(stack.Body.Created)
		data.Updated = types.StringValue(stack.Body.Updated)
	}

	var components []StackComponentModel
	if stack.Metadata != nil && stack.Metadata.Components != nil {
		for _, compList := range stack.Metadata.Components {
			for _, comp := range compList {
				component := StackComponentModel{
					ID:   types.StringValue(comp.ID),
					Name: types.StringValue(comp.Name),
				}
				if comp.Body != nil {
					component.Type = types.StringValue(comp.Body.Type)
					component.Flavor = types.StringValue(comp.Body.Flavor)
				}
				components = append(components, component)
			}
		}
	}

	componentsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":     types.StringType,
			"name":   types.StringType,
			"type":   types.StringType,
			"flavor": types.StringType,
		},
	}, components)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Components = componentsList
	}

	if stack.Metadata != nil && stack.Metadata.Labels != nil {
		labelMap := make(map[string]attr.Value)
		for k, v := range stack.Metadata.Labels {
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
