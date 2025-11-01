// resource_stack_component.go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &StackComponentResource{}
var _ resource.ResourceWithImportState = &StackComponentResource{}
var _ resource.ResourceWithConfigValidators = &StackComponentResource{}

func NewStackComponentResource() resource.Resource {
	return &StackComponentResource{}
}

type StackComponentResource struct {
	client *Client
}

type StackComponentResourceModel struct {
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

func (r *StackComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_component"
}

func (r *StackComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stack component resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Stack component identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stack component",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the stack component",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(validComponentTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flavor": schema.StringAttribute{
				MarkdownDescription: "Flavor of the stack component",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "Configuration for the stack component",
				ElementType:         types.StringType,
				Optional:            true,
				Sensitive:           true,
			},
			"connector_id": schema.StringAttribute{
				MarkdownDescription: "ID of the service connector to use for this component",
				Optional:            true,
				// We cannot delete service connectors while they are still in
				// use by a component, so we need to force new components when
				// the connector is changed.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"connector_resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID to use from the service connector",
				Optional:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the stack component",
				ElementType:         types.StringType,
				Optional:            true,
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

func (r *StackComponentResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&stackComponentConfigValidator{},
	}
}

// stackComponentConfigValidator validates that if connector_resource_id is set, connector_id should also be set
type stackComponentConfigValidator struct{}

func (v stackComponentConfigValidator) Description(ctx context.Context) string {
	return "Validates that connector_id is set when connector_resource_id is specified"
}

func (v stackComponentConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that connector_id is set when connector_resource_id is specified"
}

func (v stackComponentConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackComponentResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if connector_resource_id is set but connector_id is not
	if !data.ConnectorResourceID.IsNull() && !data.ConnectorResourceID.IsUnknown() &&
		data.ConnectorResourceID.ValueString() != "" &&
		(data.ConnectorID.IsNull() || data.ConnectorID.IsUnknown() || data.ConnectorID.ValueString() == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("connector_id"),
			"Missing connector_id",
			"connector_id must be set when connector_resource_id is specified",
		)
	}
}

func (r *StackComponentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *StackComponentResource) populateStackComponentModel(
	ctx context.Context,
	component *ComponentResponse,
	data *StackComponentResourceModel,
	diags *diag.Diagnostics,
	updateConfiguration bool,
) {
	data.ID = types.StringValue(component.ID)
	data.Name = types.StringValue(component.Name)

	if component.Body != nil {
		data.Type = types.StringValue(component.Body.Type)
		data.Flavor = types.StringValue(component.Body.Flavor)
		data.Created = types.StringValue(component.Body.Created)
		data.Updated = types.StringValue(component.Body.Updated)
	}

	if component.Metadata != nil {
		if component.Metadata.Configuration != nil {
			if updateConfiguration {
				cfg, changed := MergeOrCompareConfiguration(
					ctx,
					data.Configuration,
					component.Metadata.Configuration,
					diags,
					true,
				)
				if !diags.HasError() && changed {
					data.Configuration = cfg
				}
			} else {
				_, _ = MergeOrCompareConfiguration(
					ctx,
					data.Configuration,
					component.Metadata.Configuration,
					diags,
					false,
				)
			}
		}

		if component.Metadata.Labels != nil {
			labelMap := make(map[string]attr.Value)
			for k, v := range component.Metadata.Labels {
				labelMap[k] = types.StringValue(v)
			}
			labelValue, labelDiags := types.MapValue(types.StringType, labelMap)
			diags.Append(labelDiags...)
			if !diags.HasError() {
				data.Labels = labelValue
			}
		}

		if component.Metadata.Connector != nil {
			data.ConnectorID = types.StringValue(component.Metadata.Connector.ID)
		}
		if component.Metadata.ConnectorResourceID != nil {
			data.ConnectorResourceID = types.StringValue(*component.Metadata.ConnectorResourceID)
		}
	}
}

func (r *StackComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StackComponentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get current user, got error: %s", err))
		return
	}

	configuration := make(map[string]interface{})
	if !data.Configuration.IsNull() {
		configElements := make(map[string]types.String, len(data.Configuration.Elements()))
		resp.Diagnostics.Append(data.Configuration.ElementsAs(ctx, &configElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range configElements {
			// Always treat configuration values as strings
			// The ZenML API expects string values, and any JSON encoding
			// should be handled by the user in their Terraform configuration
			configuration[k] = v.ValueString()
		}
	}

	labels := make(map[string]string)
	if !data.Labels.IsNull() {
		labelElements := make(map[string]types.String, len(data.Labels.Elements()))
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range labelElements {
			labels[k] = v.ValueString()
		}
	}

	componentReq := ComponentRequest{
		User:          user.ID,
		Name:          data.Name.ValueString(),
		Type:          data.Type.ValueString(),
		Flavor:        data.Flavor.ValueString(),
		Configuration: configuration,
		Labels:        labels,
	}

	if !data.ConnectorID.IsNull() && data.ConnectorID.ValueString() != "" {
		connectorID := data.ConnectorID.ValueString()
		componentReq.ConnectorID = &connectorID
	}

	if !data.ConnectorResourceID.IsNull() && data.ConnectorResourceID.ValueString() != "" {
		connectorResourceID := data.ConnectorResourceID.ValueString()
		componentReq.ConnectorResourceID = &connectorResourceID
	}

	tflog.Trace(ctx, "creating stack component")

	component, err := r.client.CreateComponent(ctx, componentReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create stack component, got error: %s", err))
		return
	}

	r.populateStackComponentModel(ctx, component, &data, &resp.Diagnostics, false)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a stack component")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StackComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	component, err := r.client.GetComponent(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack component, got error: %s", err))
		return
	}

	if component == nil {
		// Component was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateStackComponentModel(ctx, component, &data, &resp.Diagnostics, true)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StackComponentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configuration := make(map[string]interface{})
	if !data.Configuration.IsNull() {
		configElements := make(map[string]types.String, len(data.Configuration.Elements()))
		resp.Diagnostics.Append(data.Configuration.ElementsAs(ctx, &configElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range configElements {
			// Always treat configuration values as strings
			// The ZenML API expects string values, and any JSON encoding
			// should be handled by the user in their Terraform configuration
			configuration[k] = v.ValueString()
		}
	}

	labels := make(map[string]string)
	if !data.Labels.IsNull() {
		labelElements := make(map[string]types.String, len(data.Labels.Elements()))
		resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &labelElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range labelElements {
			labels[k] = v.ValueString()
		}
	}

	updateReq := ComponentUpdate{
		Configuration: configuration,
		Labels:        labels,
	}

	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		updateReq.Name = &name
	}

	if !data.ConnectorID.IsNull() && data.ConnectorID.ValueString() != "" {
		connectorID := data.ConnectorID.ValueString()
		updateReq.ConnectorID = &connectorID
	}

	if !data.ConnectorResourceID.IsNull() && data.ConnectorResourceID.ValueString() != "" {
		connectorResourceID := data.ConnectorResourceID.ValueString()
		updateReq.ConnectorResourceID = &connectorResourceID
	}

	tflog.Trace(ctx, "updating stack component")

	component, err := r.client.UpdateComponent(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update stack component, got error: %s", err))
		return
	}

	r.populateStackComponentModel(ctx, component, &data, &resp.Diagnostics, false)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StackComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "deleting stack component")

	err := r.client.DeleteComponent(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete stack component, got error: %s", err))
		return
	}
}

func (r *StackComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
