// resource_stack.go
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &StackResource{}
var _ resource.ResourceWithImportState = &StackResource{}
var _ resource.ResourceWithConfigValidators = &StackResource{}

func NewStackResource() resource.Resource {
	return &StackResource{}
}

type StackResource struct {
	client *Client
}

type StackResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Components types.Map    `tfsdk:"components"`
	Labels     types.Map    `tfsdk:"labels"`
}

func (r *StackResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (r *StackResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stack resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Stack identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the stack",
				Required:            true,
			},
			"components": schema.MapAttribute{
				MarkdownDescription: "Map of component types to component IDs",
				ElementType:         types.StringType,
				Required:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the stack",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *StackResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&stackConfigValidator{},
	}
}

type stackConfigValidator struct{}

func (v stackConfigValidator) Description(ctx context.Context) string {
	return "Validates stack component types and ensures component IDs are non-null and non-empty"
}

func (v stackConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates stack component types and ensures component IDs are non-null and non-empty"
}

func (v stackConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data StackResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Components.IsNull() && !data.Components.IsUnknown() {
		componentElements := make(map[string]types.String, len(data.Components.Elements()))
		resp.Diagnostics.Append(data.Components.ElementsAs(ctx, &componentElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for compType, compID := range componentElements {
			valid := false
			for _, validType := range validComponentTypes {
				if compType == validType {
					valid = true
					break
				}
			}
			if !valid {
				resp.Diagnostics.AddAttributeError(
					path.Root("components").AtMapKey(compType),
					"Invalid component type",
					fmt.Sprintf("Invalid component type %q. Valid types are: %s", compType, strings.Join(validComponentTypes, ", ")),
				)
				continue
			}

			if compID.IsNull() || compID.IsUnknown() || strings.TrimSpace(compID.ValueString()) == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("components").AtMapKey(compType),
					"Invalid component ID",
					"Component IDs in the stack components map must be known, non-null, and non-empty.",
				)
			}
		}
	}
}

func (r *StackResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StackResource) populateStackModel(
	ctx context.Context,
	stack *StackResponse,
	data *StackResourceModel,
	diags *diag.Diagnostics,
) {
	data.ID = types.StringValue(stack.ID)
	data.Name = types.StringValue(stack.Name)

	if stack.Metadata != nil {
		componentValue := r.flattenStackComponentsToTFMap(ctx, stack.Metadata.Components, data.Components, diags)
		if diags.HasError() {
			return
		}
		if componentValue != nil {
			data.Components = *componentValue
		}

		labelValue := flattenStringMapToTFMap(stack.Metadata.Labels, data.Labels, diags)
		if diags.HasError() {
			return
		}
		if labelValue != nil {
			data.Labels = *labelValue
		}
	}
}

func expandStackComponentsFromTF(
	ctx context.Context,
	components types.Map,
	diags *diag.Diagnostics,
) map[string][]string {
	result := make(map[string][]string)
	if components.IsNull() || components.IsUnknown() {
		return result
	}

	componentElements := make(map[string]types.String, len(components.Elements()))
	diags.Append(components.ElementsAs(ctx, &componentElements, false)...)
	if diags.HasError() {
		return nil
	}

	for compType, compID := range componentElements {
		if compID.IsNull() || compID.IsUnknown() {
			continue
		}

		id := strings.TrimSpace(compID.ValueString())
		if id == "" {
			continue
		}
		result[compType] = []string{id}
	}

	return result
}

func expandStringMapFromTF(
	ctx context.Context,
	values types.Map,
	diags *diag.Diagnostics,
) map[string]string {
	result := make(map[string]string)
	if values.IsNull() || values.IsUnknown() {
		return result
	}

	elements := make(map[string]types.String, len(values.Elements()))
	diags.Append(values.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return nil
	}

	for k, v := range elements {
		if v.IsNull() || v.IsUnknown() {
			continue
		}
		result[k] = v.ValueString()
	}

	return result
}

func flattenStringMapToTFMap(
	values map[string]string,
	existing types.Map,
	diags *diag.Diagnostics,
) *types.Map {
	if len(values) == 0 {
		if !existing.IsNull() && len(existing.Elements()) > 0 {
			nullMap := types.MapNull(types.StringType)
			return &nullMap
		}
		return nil
	}

	valueMap := make(map[string]attr.Value, len(values))
	for k, v := range values {
		valueMap[k] = types.StringValue(v)
	}

	tfMap, tfDiags := types.MapValue(types.StringType, valueMap)
	diags.Append(tfDiags...)
	if diags.HasError() {
		return nil
	}

	return &tfMap
}

func (r *StackResource) flattenStackComponentsToTFMap(
	ctx context.Context,
	apiComponents map[string][]ComponentResponse,
	existing types.Map,
	diags *diag.Diagnostics,
) *types.Map {
	if apiComponents == nil {
		return nil
	}

	componentMap := make(map[string]attr.Value)

	// Preserve explicit null placeholders from prior state without inventing new ones.
	if !existing.IsNull() && !existing.IsUnknown() {
		existingComponents := make(map[string]types.String)
		diags.Append(existing.ElementsAs(ctx, &existingComponents, false)...)
		if diags.HasError() {
			return nil
		}

		for compType, compValue := range existingComponents {
			if compValue.IsNull() {
				componentMap[compType] = types.StringNull()
			}
		}
	}

	for compType, compList := range apiComponents {
		if len(compList) == 0 {
			continue
		}

		componentID := strings.TrimSpace(compList[0].ID)
		if componentID == "" {
			continue
		}

		componentMap[compType] = types.StringValue(componentID)
	}

	tfMap, tfDiags := types.MapValue(types.StringType, componentMap)
	diags.Append(tfDiags...)
	if diags.HasError() {
		return nil
	}

	return &tfMap
}

func (r *StackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	components := expandStackComponentsFromTF(ctx, data.Components, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := expandStringMapFromTF(ctx, data.Labels, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	stackReq := StackRequest{
		Name:       data.Name.ValueString(),
		Components: components,
		Labels:     labels,
	}

	tflog.Trace(ctx, "creating stack")

	stack, err := r.client.CreateStack(ctx, stackReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create stack, got error: %s", err))
		return
	}

	r.populateStackModel(ctx, stack, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a stack")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stack, err := r.client.GetStack(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read stack, got error: %s", err))
		return
	}

	if stack == nil {
		// Stack was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateStackModel(ctx, stack, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	components := expandStackComponentsFromTF(ctx, data.Components, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := expandStringMapFromTF(ctx, data.Labels, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := StackUpdate{
		Name:       data.Name.ValueString(),
		Components: components,
		Labels:     labels,
	}

	tflog.Trace(ctx, "updating stack")

	stack, err := r.client.UpdateStack(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update stack, got error: %s", err))
		return
	}

	r.populateStackModel(ctx, stack, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "deleting stack")

	err := r.client.DeleteStack(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete stack, got error: %s", err))
		return
	}
}

func (r *StackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
