package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var stringSetType = tftypes.Set{ElementType: tftypes.String}

// stepOperatorSet builds a known `step_operators` value from the given IDs.
func stepOperatorSet(ids ...tftypes.Value) tftypes.Value {
	return tftypes.NewValue(stringSetType, ids)
}

func validateStackConfig(
	t *testing.T,
	components map[string]tftypes.Value,
	stepOperators tftypes.Value,
) resource.ValidateConfigResponse {
	t.Helper()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	(&StackResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}

	stringMap := tftypes.Map{ElementType: tftypes.String}

	raw := tftypes.NewValue(
		schemaResp.Schema.Type().TerraformType(ctx),
		map[string]tftypes.Value{
			"id":             tftypes.NewValue(tftypes.String, nil),
			"name":           tftypes.NewValue(tftypes.String, "stack"),
			"components":     tftypes.NewValue(stringMap, components),
			"step_operators": stepOperators,
			"labels":         tftypes.NewValue(stringMap, nil),
		},
	)

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw},
	}
	resp := resource.ValidateConfigResponse{}
	stackConfigValidator{}.ValidateResource(ctx, req, &resp)
	return resp
}

func requiredStackComponents() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"artifact_store": tftypes.NewValue(tftypes.String, "store-id"),
		"orchestrator":   tftypes.NewValue(tftypes.String, "orch-id"),
	}
}

func TestStackConfigValidator_AllowsMultipleStepOperators(t *testing.T) {
	resp := validateStackConfig(t, requiredStackComponents(), stepOperatorSet(
		tftypes.NewValue(tftypes.String, "step-op-1"),
		tftypes.NewValue(tftypes.String, "step-op-2"),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestStackConfigValidator_RejectsStepOperatorInBothPlaces(t *testing.T) {
	components := requiredStackComponents()
	components["step_operator"] = tftypes.NewValue(tftypes.String, "step-op-1")

	resp := validateStackConfig(t, components, stepOperatorSet(
		tftypes.NewValue(tftypes.String, "step-op-2"),
	))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when step operators are set in both components and step_operators")
	}
}

func TestStackConfigValidator_AllowsNullStepOperatorPlaceholderWithSet(t *testing.T) {
	components := requiredStackComponents()
	components["step_operator"] = tftypes.NewValue(tftypes.String, nil)

	resp := validateStackConfig(t, components, stepOperatorSet(
		tftypes.NewValue(tftypes.String, "step-op-1"),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestStackConfigValidator_RejectsEmptyStepOperatorID(t *testing.T) {
	resp := validateStackConfig(t, requiredStackComponents(), stepOperatorSet(
		tftypes.NewValue(tftypes.String, "  "),
	))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for an empty step operator ID")
	}
}

func TestStackConfigValidator_RejectsInvalidStepOperatorEntries(t *testing.T) {
	tests := map[string]tftypes.Value{
		"null entry": stepOperatorSet(
			tftypes.NewValue(tftypes.String, nil),
		),
		"valid and null entry": stepOperatorSet(
			tftypes.NewValue(tftypes.String, "step-op-1"),
			tftypes.NewValue(tftypes.String, nil),
		),
		"surrounding whitespace": stepOperatorSet(
			tftypes.NewValue(tftypes.String, " step-op-1 "),
		),
		"empty string": stepOperatorSet(
			tftypes.NewValue(tftypes.String, ""),
		),
	}

	for name, stepOperators := range tests {
		t.Run(name, func(t *testing.T) {
			resp := validateStackConfig(t, requiredStackComponents(), stepOperators)
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected an error for an invalid step operator entry")
			}
		})
	}
}

func TestStackConfigValidator_AllowsUnknownStepOperatorEntry(t *testing.T) {
	resp := validateStackConfig(t, requiredStackComponents(), stepOperatorSet(
		tftypes.NewValue(tftypes.String, "step-op-1"),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestStackConfigValidator_SkipsConflictCheckForUnknownComponentStepOperator(t *testing.T) {
	components := requiredStackComponents()
	components["step_operator"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

	resp := validateStackConfig(t, components, stepOperatorSet(
		tftypes.NewValue(tftypes.String, "step-op-1"),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestStackConfigValidator_SkipsConflictCheckForUnknownStepOperatorSet(t *testing.T) {
	components := requiredStackComponents()
	components["step_operator"] = tftypes.NewValue(tftypes.String, "step-op-1")

	resp := validateStackConfig(t, components, tftypes.NewValue(stringSetType, tftypes.UnknownValue))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}
