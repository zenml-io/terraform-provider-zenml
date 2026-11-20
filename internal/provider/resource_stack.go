// resource_stack.go
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
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
var _ resource.ResourceWithModifyPlan = &StackResource{}

// requiresReplaceIfRequiredComponentChanges is a plan modifier that triggers
// resource replacement when required component types (orchestrator,
// artifact_store) change. These components cannot be removed from a stack, so
// changing them requires replacing the entire stack.
type requiresReplaceIfRequiredComponentChanges struct{}

var _ planmodifier.Map = requiresReplaceIfRequiredComponentChanges{}

func (m requiresReplaceIfRequiredComponentChanges) Description(
	ctx context.Context,
) string {
	return "Requires replacement when orchestrator or artifact_store changes"
}

func (m requiresReplaceIfRequiredComponentChanges) MarkdownDescription(
	ctx context.Context,
) string {
	return "Requires replacement when orchestrator or artifact_store changes"
}

func (m requiresReplaceIfRequiredComponentChanges) PlanModifyMap(
	ctx context.Context,
	req planmodifier.MapRequest,
	resp *planmodifier.MapResponse,
) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	var oldComponents, newComponents map[string]types.String
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &oldComponents, false)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &newComponents, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for requiredType := range requiredComponentTypes {
		oldVal := oldComponents[requiredType]
		newVal := newComponents[requiredType]

		oldID := ""
		if !oldVal.IsNull() && !oldVal.IsUnknown() {
			oldID = oldVal.ValueString()
		}
		newID := ""
		if !newVal.IsNull() && !newVal.IsUnknown() {
			newID = newVal.ValueString()
		}

		if oldID != newID {
			resp.RequiresReplace = true
			return
		}
	}
}

func NewStackResource() resource.Resource {
	return &StackResource{}
}

type StackResource struct {
	client *Client
}

type StackResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Components    types.Map    `tfsdk:"components"`
	StepOperators types.Set    `tfsdk:"step_operators"`
	Labels        types.Map    `tfsdk:"labels"`
}

// stepOperatorComponentType is the component type that ZenML allows to be
// attached to a stack multiple times.
const stepOperatorComponentType = "step_operator"

// minServerVersionMultipleComponentsPerType is the first ZenML server version
// that supports attaching multiple components of the same type to a stack
// (zenml-io/zenml#4671).
const minServerVersionMultipleComponentsPerType = "0.94.3"

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
				MarkdownDescription: "Map of component types to component IDs. " +
					"Changing the orchestrator or artifact_store will force " +
					"stack replacement since these are required components.",
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Map{
					requiresReplaceIfRequiredComponentChanges{},
				},
			},
			"step_operators": schema.SetAttribute{
				MarkdownDescription: "Set of step operator component IDs to " +
					"attach to the stack. Use this instead of " +
					"`components.step_operator` when the stack needs more " +
					"than one step operator. Cannot be combined with " +
					"`components.step_operator`.",
				ElementType: types.StringType,
				Optional:    true,
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

			if compID.IsUnknown() || compID.IsNull() {
				continue
			}

			if strings.TrimSpace(compID.ValueString()) == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("components").AtMapKey(compType),
					"Invalid component ID",
					"Component IDs must be non-empty strings when provided; null and unknown values are allowed during planning.",
				)
			}
		}

		// Unknown values may still resolve to null, so only report a conflict
		// once both sides are known. Terraform validates again at apply time.
		if stepOperator, ok := componentElements[stepOperatorComponentType]; ok &&
			!stepOperator.IsNull() && !stepOperator.IsUnknown() &&
			!data.StepOperators.IsNull() && !data.StepOperators.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("step_operators"),
				"Conflicting step operator configuration",
				"Step operators can be configured either via `components.step_operator` "+
					"or via `step_operators`, but not both. Use `step_operators` to "+
					"attach multiple step operators to a stack.",
			)
		}
	}

	if !data.StepOperators.IsNull() && !data.StepOperators.IsUnknown() {
		stepOperatorElements := make([]types.String, 0, len(data.StepOperators.Elements()))
		resp.Diagnostics.Append(data.StepOperators.ElementsAs(ctx, &stepOperatorElements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, id := range stepOperatorElements {
			if id.IsUnknown() {
				continue
			}
			if id.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("step_operators"),
					"Invalid step operator ID",
					"Step operator IDs must not be null.",
				)
				continue
			}
			value := id.ValueString()
			if value == "" || strings.TrimSpace(value) != value {
				resp.Diagnostics.AddAttributeError(
					path.Root("step_operators"),
					"Invalid step operator ID",
					fmt.Sprintf("Step operator IDs must be non-empty and must not contain "+
						"leading or trailing whitespace, got %q.", value),
				)
			}
		}
	}
}

// ModifyPlan rejects plans that attach more than one step operator to a stack
// when the ZenML server is too old to support multiple components of the same
// type. Older servers silently keep only one of them, which would otherwise
// surface as an "inconsistent result after apply" error after the stack has
// already been modified.
func (r *StackResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.client == nil {
		return
	}

	var stepOperators types.Set
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("step_operators"), &stepOperators)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(checkServerSupportsStepOperators(r.client.ServerVersion, stepOperators)...)
}

// checkServerSupportsStepOperators returns an error diagnostic if more than
// one step operator is configured and the server version is known to be older
// than minServerVersionMultipleComponentsPerType. Unparseable or unknown
// server versions are not rejected.
func checkServerSupportsStepOperators(serverVersion string, stepOperators types.Set) diag.Diagnostics {
	var diags diag.Diagnostics

	if stepOperators.IsNull() || stepOperators.IsUnknown() || len(stepOperators.Elements()) <= 1 {
		return diags
	}

	current, err := version.NewVersion(serverVersion)
	if err != nil {
		return diags
	}
	minimum := version.Must(version.NewVersion(minServerVersionMultipleComponentsPerType))
	if !current.LessThan(minimum) {
		return diags
	}

	diags.AddAttributeError(
		path.Root("step_operators"),
		"Multiple step operators not supported by ZenML server",
		fmt.Sprintf("Attaching more than one step operator to a stack requires ZenML "+
			"server version %s or later, but the server reports version %s. "+
			"Upgrade the ZenML server or configure at most one step operator.",
			minServerVersionMultipleComponentsPerType, serverVersion),
	)
	return diags
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
		apiComponents := stack.Metadata.Components

		// Step operators are tracked in the dedicated `step_operators` set if
		// the user manages them there, or if the stack has more than one step
		// operator (which `components` cannot represent). Otherwise they stay
		// in `components.step_operator` for backwards compatibility.
		stepOperatorIDs := componentIDs(apiComponents[stepOperatorComponentType])
		useStepOperatorSet := !data.StepOperators.IsNull() || len(stepOperatorIDs) > 1
		if useStepOperatorSet {
			apiComponents = withoutComponentType(apiComponents, stepOperatorComponentType)
		}

		componentValue := r.flattenStackComponentsToTFMap(ctx, apiComponents, data.Components, diags)
		if diags.HasError() {
			return
		}
		if componentValue != nil {
			data.Components = *componentValue
		}

		if useStepOperatorSet {
			data.StepOperators = flattenStringSliceToTFSet(stepOperatorIDs, diags)
			if diags.HasError() {
				return
			}
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

// componentIDs returns the trimmed, non-empty IDs of the given components.
func componentIDs(components []ComponentResponse) []string {
	ids := make([]string, 0, len(components))
	for _, comp := range components {
		id := strings.TrimSpace(comp.ID)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// withoutComponentType returns a shallow copy of components without the given
// component type.
func withoutComponentType(
	components map[string][]ComponentResponse,
	componentType string,
) map[string][]ComponentResponse {
	if components == nil {
		return nil
	}
	result := make(map[string][]ComponentResponse, len(components))
	for compType, compList := range components {
		if compType != componentType {
			result[compType] = compList
		}
	}
	return result
}

// flattenStringSliceToTFSet converts a list of strings into a Terraform set.
// An empty list is represented as an empty (non-null) set.
func flattenStringSliceToTFSet(values []string, diags *diag.Diagnostics) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	set, setDiags := types.SetValue(types.StringType, elements)
	diags.Append(setDiags...)
	return set
}

// expandStepOperatorsFromTF converts the `step_operators` set into a list of
// component IDs. Values are passed through unchanged (the config validator
// rejects null, empty and whitespace-padded IDs) so that the stack returned by
// the API matches the plan exactly.
func expandStepOperatorsFromTF(
	ctx context.Context,
	stepOperators types.Set,
	diags *diag.Diagnostics,
) []string {
	if stepOperators.IsNull() || stepOperators.IsUnknown() {
		return nil
	}

	elements := make([]types.String, 0, len(stepOperators.Elements()))
	diags.Append(stepOperators.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return nil
	}

	ids := make([]string, 0, len(elements))
	for _, el := range elements {
		if el.IsNull() || el.IsUnknown() {
			continue
		}
		ids = append(ids, el.ValueString())
	}
	return ids
}

// expandStackFromTF builds the component map sent to the ZenML API from both
// the `components` map and the `step_operators` set.
func expandStackFromTF(
	ctx context.Context,
	data *StackResourceModel,
	diags *diag.Diagnostics,
) map[string][]string {
	components := expandStackComponentsFromTF(ctx, data.Components, diags)
	if diags.HasError() {
		return nil
	}

	stepOperators := expandStepOperatorsFromTF(ctx, data.StepOperators, diags)
	if diags.HasError() {
		return nil
	}
	if len(stepOperators) > 0 {
		components[stepOperatorComponentType] = append(
			components[stepOperatorComponentType], stepOperators...,
		)
	}

	return components
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

	components := expandStackFromTF(ctx, &data, &resp.Diagnostics)
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

	components := expandStackFromTF(ctx, &data, &resp.Diagnostics)
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
