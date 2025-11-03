// validation.go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// All validation constants and variables
var (
	validConnectorTypes = []string{
		"aws", "gcp", "azure", "kubernetes",
		"docker", "hyperai",
	}

	validResourceTypes = map[string][]string{
		"aws": {
			"aws-generic",
			"s3-bucket",
			"kubernetes-cluster",
			"docker-registry",
		},
		"gcp": {
			"gcp-generic",
			"gcs-bucket",
			"kubernetes-cluster",
			"docker-registry",
		},
		"azure": {
			"azure-generic",
			"blob-container",
			"kubernetes-cluster",
			"docker-registry",
		},
		"kubernetes": {
			"kubernetes-cluster",
		},
		"docker": {
			"docker-registry",
		},
		"hyperai": {
			"hyperai-instance",
		},
	}

	validAuthMethods = map[string][]string{
		"aws": {
			"implicit",
			"secret-key",
			"iam-role",
			"sts-token",
			"session-token",
			"federation-token",
		},
		"gcp": {
			"implicit",
			"user-account",
			"service-account",
			"external-account",
			"oauth2-token",
			"impersonation",
		},
		"azure": {
			"implicit",
			"service-principal",
			"access-token",
		},
		"kubernetes": {
			"password",
			"token",
		},
		"docker": {
			"password",
		},
		"hyperai": {
			"rsa-key",
			"dsa-key",
			"ecdsa-key",
			"ed25519-key",
		},
	}

	validComponentTypes = []string{
		"alerter",
		"annotator",
		"artifact_store",
		"container_registry",
		"data_validator",
		"experiment_tracker",
		"feature_store",
		"image_builder",
		"model_deployer",
		"orchestrator",
		"step_operator",
		"model_registry",
		"deployer",
	}
)

func NormalizeServerConfig(raw map[string]interface{}) map[string]string {
	if raw == nil {
		return map[string]string{}
	}
	normalized := make(map[string]string, len(raw))
	for k, v := range raw {
		switch vv := v.(type) {
		case string:
			normalized[k] = vv
		default:
			normalized[k] = fmt.Sprintf("%v", vv)
		}
	}
	return normalized
}

func MergeOrCompareConfiguration(
	ctx context.Context,
	existing types.Map,
	serverRaw map[string]interface{},
	diags *diag.Diagnostics,
) (types.Map, bool) {

	// MergeOrCompareConfiguration centralizes the logic for reconciling a
	// Terraform configuration (types.Map of strings) with a server-provided
	// configuration (raw map). It returns a merged configuration with server
	// values overlayed on top of the existing data configuration, and warns
	// about any keys present only in data.
	//
	// This is required because the ZenML server may mutate the configuration
	// of stack components and service connectors on create/update, and we need
	// to ensure that the Terraform configuration provided by the user is never
	// overwritten by the server.

	serverConfig := NormalizeServerConfig(serverRaw)

	existingTyped := make(map[string]types.String)
	if !existing.IsNull() && !existing.IsUnknown() {
		diags.Append(existing.ElementsAs(ctx, &existingTyped, false)...)
		if diags.HasError() {
			return types.Map{}, false
		}
	}

	// Merge: start with existing, overlay server.
	merged := make(map[string]attr.Value, len(existingTyped)+len(serverConfig))
	ignoredKeys := make([]string, 0)

	for k, v := range existingTyped {
		if _, ok := serverConfig[k]; !ok {
			ignoredKeys = append(ignoredKeys, k)
		}
		merged[k] = types.StringValue(v.ValueString())
	}

	for k, v := range serverConfig {
		merged[k] = types.StringValue(v)
	}

	cfg, cfgDiags := types.MapValue(types.StringType, merged)
	diags.Append(cfgDiags...)
	if diags.HasError() {
		return types.Map{}, false
	}

	if len(ignoredKeys) > 0 {
		diags.AddWarning(
			"Configuration attributes ignored by ZenML server",
			fmt.Sprintf(
				"The following configuration attributes are present in Terraform "+
					"state but not recognized by the server and were ignored: %v.",
				ignoredKeys,
			),
		)
	}

	return cfg, true
}
