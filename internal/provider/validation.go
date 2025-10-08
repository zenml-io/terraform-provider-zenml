// validation.go
package provider

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
