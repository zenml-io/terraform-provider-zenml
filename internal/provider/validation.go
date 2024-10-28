// validation.go
package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

// All validation constants and variables
var (
	validConnectorTypes = []string{
		"aws", "gcp", "azure", "kubernetes",
		"github", "gitlab", "bitbucket", "docker",
		"mysql", "postgres", "snowflake", "databricks",
	}

	validServiceTypes = map[string][]string{
		"aws": {
			"artifact-store",
			"container-registry",
			"secrets-manager",
			"step-operator",
		},
		"gcp": {
			"artifact-store", 
			"container-registry",
			"secrets-manager",
			"step-operator",
		},
		"azure": {
			"artifact-store",
			"container-registry",
			"step-operator",
		},
		"kubernetes": {
			"orchestrator",
			"step-operator",
		},
	}

	validAuthMethods = map[string][]string{
		"aws": {
			"iam-role",
			"aws-access-keys",
			"web-identity",
		},
		"gcp": {
			"service-account",
			"oauth2",
			"workload-identity",
		},
		"azure": {
			"service-principal",
			"managed-identity",
		},
		"kubernetes": {
			"kubeconfig",
			"service-account",
		},
	}

	// Required configuration fields per connector type
	requiredConfigFields = map[string][]string{
		"aws":        {"region"},
		"gcp":        {"project_id"},
		"azure":      {"subscription_id", "tenant_id"},
		"kubernetes": {"context"},
	}

	// Optional configuration fields per connector type
	optionalConfigFields = map[string][]string{
		"aws":        {"role_arn", "external_id", "session_name"},
		"gcp":        {"zone", "location"},
		"azure":      {"resource_group"},
		"kubernetes": {"namespace", "cluster_name"},
	}

	// Required secrets per auth method
	requiredSecrets = map[string]map[string][]string{
		"aws": {
			"aws-access-keys": {"aws_access_key_id", "aws_secret_access_key"},
		},
		"gcp": {
			"service-account": {"service_account_json"},
		},
		"azure": {
			"service-principal": {"client_id", "client_secret"},
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

	// Validate required configuration fields
	if fields, ok := requiredConfigFields[connectorType]; ok {
		config := d.Get("configuration").(map[string]interface{})
		for _, field := range fields {
			if _, ok := config[field]; !ok {
				return fmt.Errorf("required configuration field %q missing for connector type %q",
					field, connectorType)
			}
		}
	}

	// Validate required secrets
	if secretFields, ok := requiredSecrets[connectorType][authMethod]; ok {
		secrets := d.Get("secrets").(map[string]interface{})
		for _, field := range secretFields {
			if _, ok := secrets[field]; !ok {
				return fmt.Errorf("required secret %q missing for auth method %q",
					field, authMethod)
			}
		}
	}

	// Validate resource types
	if v, ok := d.GetOk("resource_types"); ok {
		resourceTypes := v.(*schema.Set).List()
		validTypes := getValidResourceTypes(connectorType)
		for _, rt := range resourceTypes {
			found := false
			for _, vt := range validTypes {
				if rt.(string) == vt {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid resource type %q for connector type %q. Valid types are: %s",
					rt.(string), connectorType, strings.Join(validTypes, ", "))
			}
		}
	}

	return nil
}

func getValidResourceTypes(connectorType string) []string {
	switch connectorType {
	case "aws":
		return []string{
			"artifact-store",
			"container-registry",
			"step-operator",
			"orchestrator",
		}
	case "gcp":
		return []string{
			"artifact-store",
			"container-registry",
			"step-operator",
			"orchestrator",
		}
	case "azure":
		return []string{
			"artifact-store",
			"container-registry",
			"step-operator",
		}
	case "kubernetes":
		return []string{
			"orchestrator",
			"step-operator",
		}
	default:
		return []string{}
	}
}
