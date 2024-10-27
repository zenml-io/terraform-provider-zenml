package provider

type StackComponentType string

const (
	AlerterType           StackComponentType = "alerter"
	AnnotatorType         StackComponentType = "annotator"
	ArtifactStoreType    StackComponentType = "artifact_store"
	ContainerRegistryType StackComponentType = "container_registry"
	DataValidatorType     StackComponentType = "data_validator"
	ExperimentTrackerType StackComponentType = "experiment_tracker"
	FeatureStoreType      StackComponentType = "feature_store"
	ImageBuilderType      StackComponentType = "image_builder"
	ModelDeployerType     StackComponentType = "model_deployer"
	OrchestratorType      StackComponentType = "orchestrator"
	StepOperatorType      StackComponentType = "step_operator"
	ModelRegistryType     StackComponentType = "model_registry"
)

// Component models
type ComponentRequest struct {
	User               string                 `json:"user"`
	Workspace          string                 `json:"workspace"`
	Name               string                 `json:"name"`
	Type               StackComponentType     `json:"type"`
	Flavor            string                 `json:"flavor"`
	Configuration     map[string]interface{} `json:"configuration"`
	ConnectorResourceID *string               `json:"connector_resource_id,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty"`
	ComponentSpecPath *string                `json:"component_spec_path,omitempty"`
	Connector        *string                `json:"connector,omitempty"`
}

type ComponentResponse struct {
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Body     *ComponentResponseBody  `json:"body,omitempty"`
	Metadata *ComponentResponseMetadata `json:"metadata,omitempty"`
}