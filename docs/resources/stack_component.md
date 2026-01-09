---
page_title: "zenml_stack_component Resource - terraform-provider-zenml"
subcategory: ""
description: |-
  Manages a ZenML stack component.
---

# zenml_stack_component (Resource)

Manages a ZenML stack component, which represents a specific piece of infrastructure or service used in a ZenML stack.

## Example Usage

```hcl
resource "zenml_stack_component" "artifact_store" {
  name      = "my-artifact-store"
  type      = "artifact_store"
  flavor    = "gcp"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
  
  # Optional: Connect to a service connector
  connector_id           = "connector-uuid"
  connector_resource_id  = "resource-id"
  
  labels = {
    environment = "production"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the stack component.
* `type` - (Required) The type of the stack component. Must be one of the valid component types supported by ZenML:
  * `alerter` - Alerter
  * `annotator` - Annotator
  * `artifact_store` - Artifact store
  * `container_registry` - Container registry
  * `data_validator` - Data validator
  * `experiment_tracker` - Experiment tracker
  * `feature_store` - Feature store
  * `image_builder` - Image builder
  * `model_deployer` - Model deployer
  * `orchestrator` - Orchestrator
  * `step_operator` - Step operator
  * `model_registry` - Model registry
  * `deployer` - Deployer
  * `log_store` - Log store
* `flavor` - (Required) The flavor of the stack component (e.g., "local", "gcp", "aws"). To find out which flavors are supported by a component type, run `zenml stack-component describe-type <component-type>` or visit the [Component Gallery section of the ZenML documentation](https://docs.zenml.io/stack-components/component-guide) for more information.
* `configuration` - (Optional, Sensitive) A map of configuration key-value pairs for the component.
* `connector_id` - (Optional) The ID of the service connector to use with this component. Must be specified together with `connector_resource_id`.
* `connector_resource_id` - (Optional) The ID of the connector resource to use with this component. Must be specified together with `connector_id`.
* `labels` - (Optional) A map of labels to associate with the component.

-> **Note** When using service connectors, both `connector_id` and `connector_resource_id` must be specified together. Specifying only one will result in an error.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack component.

## Import

Stack components can be imported using the `id`, e.g.

```shell
$ terraform import zenml_stack_component.example 12345678-1234-1234-1234-123456789012
```
