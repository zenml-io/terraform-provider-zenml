package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandStackComponentsFromTF_IgnoresEmptyAndUnknownValues(t *testing.T) {
	ctx := context.Background()

	input, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"artifact_store": types.StringValue("  component-1  "),
		"orchestrator":   types.StringNull(),
		"step_operator":  types.StringValue(""),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating input map: %v", diags)
	}

	var testDiags diag.Diagnostics
	got := expandStackComponentsFromTF(ctx, input, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 component, got %d (%v)", len(got), got)
	}
	if len(got["artifact_store"]) != 1 || got["artifact_store"][0] != "component-1" {
		t.Fatalf("unexpected artifact_store expansion: %#v", got["artifact_store"])
	}
}

func TestFlattenStackComponentsToTFMap_PreservesNullAndIgnoresEmptyAPIEntries(t *testing.T) {
	ctx := context.Background()
	resource := &StackResource{}

	existing, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"experiment_tracker": types.StringNull(),
		"orchestrator":       types.StringValue("old-orchestrator"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating existing map: %v", diags)
	}

	var testDiags diag.Diagnostics
	gotPtr := resource.flattenStackComponentsToTFMap(ctx, map[string][]ComponentResponse{
		"artifact_store":     {{ID: "artifact-1"}},
		"orchestrator":       {{ID: "new-orchestrator"}},
		"container_registry": {},
		"image_builder":      {{ID: "  "}},
	}, existing, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}
	if gotPtr == nil {
		t.Fatal("expected flattened map, got nil")
	}

	gotElems := make(map[string]types.String)
	testDiags = nil
	testDiags.Append(gotPtr.ElementsAs(ctx, &gotElems, false)...)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics decoding output map: %v", testDiags)
	}

	if _, ok := gotElems["container_registry"]; ok {
		t.Fatalf("expected empty API list to be ignored, got container_registry=%v", gotElems["container_registry"])
	}
	if _, ok := gotElems["image_builder"]; ok {
		t.Fatalf("expected blank API ID to be ignored, got image_builder=%v", gotElems["image_builder"])
	}
	if gotElems["artifact_store"].ValueString() != "artifact-1" {
		t.Fatalf("unexpected artifact_store value: %q", gotElems["artifact_store"].ValueString())
	}
	if gotElems["orchestrator"].ValueString() != "new-orchestrator" {
		t.Fatalf("unexpected orchestrator value: %q", gotElems["orchestrator"].ValueString())
	}
	if val, ok := gotElems["experiment_tracker"]; !ok || !val.IsNull() {
		t.Fatalf("expected null experiment_tracker placeholder to be preserved, got %v (present=%v)", val, ok)
	}
}

func TestFlattenStringMapToTFMap_EmptiesExistingLabels(t *testing.T) {
	existing, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"environment": types.StringValue("prod"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating existing labels map: %v", diags)
	}

	var testDiags diag.Diagnostics
	gotPtr := flattenStringMapToTFMap(map[string]string{}, existing, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}
	if gotPtr == nil {
		t.Fatal("expected null labels map result, got nil")
	}
	if !gotPtr.IsNull() {
		t.Fatalf("expected null labels map, got %#v", *gotPtr)
	}
}

func TestRequiredComponentTypes_ContainsExpectedTypes(t *testing.T) {
	if !requiredComponentTypes["orchestrator"] {
		t.Error("orchestrator should be a required component type")
	}
	if !requiredComponentTypes["artifact_store"] {
		t.Error("artifact_store should be a required component type")
	}
	if requiredComponentTypes["container_registry"] {
		t.Error("container_registry should NOT be a required component type")
	}
	if requiredComponentTypes["image_builder"] {
		t.Error("image_builder should NOT be a required component type")
	}
}

func TestRequiresReplaceIfRequiredComponentChanges_OrchestratorChange(t *testing.T) {
	ctx := context.Background()
	modifier := requiresReplaceIfRequiredComponentChanges{}

	stateValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("old-orch-id"),
		"artifact_store": types.StringValue("store-id"),
	})
	planValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("new-orch-id"),
		"artifact_store": types.StringValue("store-id"),
	})

	req := planmodifier.MapRequest{
		StateValue: stateValue,
		PlanValue:  planValue,
	}
	resp := &planmodifier.MapResponse{}

	modifier.PlanModifyMap(ctx, req, resp)

	if !resp.RequiresReplace {
		t.Error("expected RequiresReplace=true when orchestrator changes")
	}
}

func TestRequiresReplaceIfRequiredComponentChanges_ArtifactStoreChange(t *testing.T) {
	ctx := context.Background()
	modifier := requiresReplaceIfRequiredComponentChanges{}

	stateValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("orch-id"),
		"artifact_store": types.StringValue("old-store-id"),
	})
	planValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("orch-id"),
		"artifact_store": types.StringValue("new-store-id"),
	})

	req := planmodifier.MapRequest{
		StateValue: stateValue,
		PlanValue:  planValue,
	}
	resp := &planmodifier.MapResponse{}

	modifier.PlanModifyMap(ctx, req, resp)

	if !resp.RequiresReplace {
		t.Error("expected RequiresReplace=true when artifact_store changes")
	}
}

func TestRequiresReplaceIfRequiredComponentChanges_OptionalComponentChange(t *testing.T) {
	ctx := context.Background()
	modifier := requiresReplaceIfRequiredComponentChanges{}

	stateValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":       types.StringValue("orch-id"),
		"artifact_store":     types.StringValue("store-id"),
		"container_registry": types.StringValue("old-registry-id"),
	})
	planValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":       types.StringValue("orch-id"),
		"artifact_store":     types.StringValue("store-id"),
		"container_registry": types.StringValue("new-registry-id"),
	})

	req := planmodifier.MapRequest{
		StateValue: stateValue,
		PlanValue:  planValue,
	}
	resp := &planmodifier.MapResponse{}

	modifier.PlanModifyMap(ctx, req, resp)

	if resp.RequiresReplace {
		t.Error("expected RequiresReplace=false when only optional component changes")
	}
}

func TestRequiresReplaceIfRequiredComponentChanges_NoChange(t *testing.T) {
	ctx := context.Background()
	modifier := requiresReplaceIfRequiredComponentChanges{}

	stateValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("orch-id"),
		"artifact_store": types.StringValue("store-id"),
	})
	planValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"orchestrator":   types.StringValue("orch-id"),
		"artifact_store": types.StringValue("store-id"),
	})

	req := planmodifier.MapRequest{
		StateValue: stateValue,
		PlanValue:  planValue,
	}
	resp := &planmodifier.MapResponse{}

	modifier.PlanModifyMap(ctx, req, resp)

	if resp.RequiresReplace {
		t.Error("expected RequiresReplace=false when nothing changes")
	}
}

func TestValidComponentTypes_IncludesSandbox(t *testing.T) {
	ctx := context.Background()

	found := false
	for _, componentType := range validComponentTypes {
		if componentType == "sandbox" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q to be a valid component type, got %v", "sandbox", validComponentTypes)
	}

	input, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"sandbox": types.StringValue("sandbox-1"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating input map: %v", diags)
	}

	var testDiags diag.Diagnostics
	got := expandStackComponentsFromTF(ctx, input, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}
	if len(got["sandbox"]) != 1 || got["sandbox"][0] != "sandbox-1" {
		t.Fatalf("unexpected sandbox expansion: %#v", got["sandbox"])
	}
}

func TestExpandStackFromTF_MergesStepOperators(t *testing.T) {
	ctx := context.Background()

	components, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"artifact_store": types.StringValue("store-id"),
		"orchestrator":   types.StringValue("orch-id"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating components map: %v", diags)
	}
	stepOperators, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("step-op-1"),
		types.StringValue("step-op-2"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating step operators set: %v", diags)
	}

	data := &StackResourceModel{Components: components, StepOperators: stepOperators}

	var testDiags diag.Diagnostics
	got := expandStackFromTF(ctx, data, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}

	if len(got["artifact_store"]) != 1 || len(got["orchestrator"]) != 1 {
		t.Fatalf("expected required components to be kept, got %#v", got)
	}
	gotStepOps := map[string]bool{}
	for _, id := range got["step_operator"] {
		gotStepOps[id] = true
	}
	if len(gotStepOps) != 2 || !gotStepOps["step-op-1"] || !gotStepOps["step-op-2"] {
		t.Fatalf("unexpected step operator expansion: %#v", got["step_operator"])
	}
}

func TestExpandStackFromTF_NullStepOperatorsKeepsSingleStepOperator(t *testing.T) {
	ctx := context.Background()

	components, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"artifact_store": types.StringValue("store-id"),
		"orchestrator":   types.StringValue("orch-id"),
		"step_operator":  types.StringValue("step-op-1"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating components map: %v", diags)
	}

	data := &StackResourceModel{
		Components:    components,
		StepOperators: types.SetNull(types.StringType),
	}

	var testDiags diag.Diagnostics
	got := expandStackFromTF(ctx, data, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}
	if len(got["step_operator"]) != 1 || got["step_operator"][0] != "step-op-1" {
		t.Fatalf("unexpected step operator expansion: %#v", got["step_operator"])
	}
}

func stackResponseWithStepOperators(ids ...string) *StackResponse {
	stepOps := make([]ComponentResponse, 0, len(ids))
	for _, id := range ids {
		stepOps = append(stepOps, ComponentResponse{ID: id})
	}
	return &StackResponse{
		ID:   "stack-id",
		Name: "stack",
		Metadata: &StackResponseMetadata{
			Components: map[string][]ComponentResponse{
				"artifact_store": {{ID: "store-id"}},
				"orchestrator":   {{ID: "orch-id"}},
				"step_operator":  stepOps,
			},
		},
	}
}

func decodeStringSet(t *testing.T, set types.Set) map[string]bool {
	t.Helper()
	var ids []string
	if diags := set.ElementsAs(context.Background(), &ids, false); diags.HasError() {
		t.Fatalf("unexpected diagnostics decoding set: %v", diags)
	}
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result
}

func TestPopulateStackModel_SingleStepOperatorStaysInComponents(t *testing.T) {
	ctx := context.Background()
	r := &StackResource{}
	data := &StackResourceModel{
		Components:    types.MapNull(types.StringType),
		StepOperators: types.SetNull(types.StringType),
		Labels:        types.MapNull(types.StringType),
	}

	var diags diag.Diagnostics
	r.populateStackModel(ctx, stackResponseWithStepOperators("step-op-1"), data, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.StepOperators.IsNull() {
		t.Fatalf("expected step_operators to stay null, got %v", data.StepOperators)
	}
	got := data.Components.Elements()["step_operator"]
	if got == nil || got.(types.String).ValueString() != "step-op-1" {
		t.Fatalf("expected components.step_operator=step-op-1, got %v", got)
	}
}

func TestPopulateStackModel_MultipleStepOperatorsUseSet(t *testing.T) {
	ctx := context.Background()
	r := &StackResource{}
	data := &StackResourceModel{
		Components:    types.MapNull(types.StringType),
		StepOperators: types.SetNull(types.StringType),
		Labels:        types.MapNull(types.StringType),
	}

	var diags diag.Diagnostics
	r.populateStackModel(ctx, stackResponseWithStepOperators("step-op-1", "step-op-2"), data, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if _, ok := data.Components.Elements()["step_operator"]; ok {
		t.Fatalf("expected step_operator to be absent from components, got %v", data.Components)
	}
	got := decodeStringSet(t, data.StepOperators)
	if len(got) != 2 || !got["step-op-1"] || !got["step-op-2"] {
		t.Fatalf("unexpected step_operators: %v", got)
	}
}

func TestPopulateStackModel_ConfiguredStepOperatorSetIsKept(t *testing.T) {
	ctx := context.Background()
	r := &StackResource{}
	existingSet, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("step-op-1"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating set: %v", diags)
	}
	data := &StackResourceModel{
		Components:    types.MapNull(types.StringType),
		StepOperators: existingSet,
		Labels:        types.MapNull(types.StringType),
	}

	r.populateStackModel(ctx, stackResponseWithStepOperators("step-op-1"), data, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if _, ok := data.Components.Elements()["step_operator"]; ok {
		t.Fatalf("expected step_operator to be absent from components, got %v", data.Components)
	}
	got := decodeStringSet(t, data.StepOperators)
	if len(got) != 1 || !got["step-op-1"] {
		t.Fatalf("unexpected step_operators: %v", got)
	}
}

func TestExpandStepOperatorsFromTF_PassesIDsThroughUnchanged(t *testing.T) {
	ctx := context.Background()

	// Whitespace-padded IDs are rejected by the config validator. Expansion
	// must not silently rewrite them, otherwise the API result would not match
	// the plan.
	stepOperators, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue(" step-op-1 "),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating step operators set: %v", diags)
	}

	var testDiags diag.Diagnostics
	got := expandStepOperatorsFromTF(ctx, stepOperators, &testDiags)
	if testDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", testDiags)
	}
	if len(got) != 1 || got[0] != " step-op-1 " {
		t.Fatalf("expected ID to be passed through unchanged, got %#v", got)
	}
}

func TestCheckServerSupportsStepOperators(t *testing.T) {
	two, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("step-op-1"),
		types.StringValue("step-op-2"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating set: %v", diags)
	}
	one, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("step-op-1"),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating set: %v", diags)
	}

	tests := []struct {
		name          string
		serverVersion string
		stepOperators types.Set
		wantError     bool
	}{
		{"multiple on old server", "0.94.2", two, true},
		{"multiple on much older server", "0.80.0", two, true},
		{"multiple on minimum version", "0.94.3", two, false},
		{"multiple on newer server", "0.97.0", two, false},
		{"single on old server", "0.94.2", one, false},
		{"null set on old server", "0.94.2", types.SetNull(types.StringType), false},
		{"unknown set on old server", "0.94.2", types.SetUnknown(types.StringType), false},
		{"unparseable server version", "not-a-version", two, false},
		{"unknown server version", "", two, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkServerSupportsStepOperators(tt.serverVersion, tt.stepOperators)
			if got.HasError() != tt.wantError {
				t.Fatalf("expected error=%v, got diagnostics %v", tt.wantError, got)
			}
		})
	}
}
