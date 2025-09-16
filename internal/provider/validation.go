// validation.go
package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func validateServiceConnector(d *schema.ResourceDiff) error {
	connectorType := d.Get("type").(string)

	// Validate connector type first
	validType := false
	for _, t := range validConnectorTypes {
		if t == connectorType {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid connector type %q. Valid types are: %s",
			connectorType, strings.Join(validConnectorTypes, ", "))
	}

	authMethod := d.Get("auth_method").(string)

	// Validate auth method for connector type
	if methods, ok := validAuthMethods[connectorType]; ok {
		validMethod := false
		for _, m := range methods {
			if m == authMethod {
				validMethod = true
				break
			}
		}
		if !validMethod {
			return fmt.Errorf("invalid auth_method %q for connector type %q. Valid methods are: %s",
				authMethod, connectorType, strings.Join(validAuthMethods[connectorType], ", "))
		}
	}

	// Validate resource type
	if v, ok := d.GetOk("resource_type"); ok {
		validTypes := validResourceTypes[connectorType]
		resourceType := v.(string)
		valid := false
		for _, t := range validTypes {
			if t == resourceType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid resource type %q for connector type %q. Valid types are: %s",
				resourceType, connectorType, strings.Join(validTypes, ", "))
		}
	}

	// NOTE: we specifically don't validate the configuration here
	// for two reasons:
	// 1. The configuration can be derived from resources and data
	//    sources that are not available during plan time.
	// 2. The configuration are validated by the ZenML server
	//    when the connector is validated / created and we don't want to
	//    duplicate that logic here.

	return nil
}
