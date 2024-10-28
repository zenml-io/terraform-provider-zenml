---
page_title: "zenml_stack Resource - terraform-provider-zenml"
subcategory: ""
description: |-
  Manages a ZenML stack.
---

# zenml_stack (Resource)

Manages a ZenML stack, which is a collection of components that define the infrastructure for your ML pipelines.

## Example Usage

```hcl
# First, create the required stack components
resource "zenml_stack_component" "artifact_store" {
  name      = "my-artifact-store"
  type      = "artifact_store"
  flavor    = "gcp"
  workspace = "default"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
}

resource "zenml_stack_component" "orchestrator" {
  name      = "my-orchestrator"
  type      = "orchestrator"
  flavor    = "kubernetes"
  workspace = "default"
  
  configuration = {
    kubernetes_context = "my-k8s-cluster"
  }
}

# Then create the stack using the component IDs
resource "zenml_stack" "my_stack" {
  name = "my-production-stack"
  
  # Map component types to their IDs
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
    orchestrator   = zenml_stack_component.orchestrator.id
  }
  
  labels = {
    environment = "production"
    team        = "ml-ops"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the stack.
* `components` - (Required) A map where keys are component types and values are component IDs. Each component type can only have one component. Valid component types include:
  * `artifact_store`
  * `container_registry`
  * `orchestrator`
  * `step_operator`
  * `model_deployer`
  * `experiment_tracker`
  * `alerter`
  * `annotator`
  * `data_validator`
  * `feature_store`
  * `image_builder`
* `labels` - (Optional) A map of labels to associate with the stack.

-> **Note** The stack will be created in the default workspace. Future versions may allow workspace configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack.

## Import

Stacks can be imported using the `id`, e.g.

```shell
$ terraform import zenml_stack.example 12345678-1234-1234-1234-123456789012
```
