package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
