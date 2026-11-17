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
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
}

resource "zenml_stack_component" "orchestrator" {
  name      = "my-orchestrator"
  type      = "orchestrator"
  flavor    = "kubernetes"
  
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

### Multiple Step Operators

A stack can have more than one step operator. Use the `step_operators` argument to attach several of them:

```hcl
resource "zenml_stack_component" "gpu_step_operator" {
  name   = "gpu-step-operator"
  type   = "step_operator"
  flavor = "kubernetes"
}

resource "zenml_stack_component" "cpu_step_operator" {
  name   = "cpu-step-operator"
  type   = "step_operator"
  flavor = "kubernetes"
}

resource "zenml_stack" "multi_step_operator_stack" {
  name = "multi-step-operator-stack"

  components = {
    artifact_store = zenml_stack_component.artifact_store.id
    orchestrator   = zenml_stack_component.orchestrator.id
  }

  step_operators = [
    zenml_stack_component.gpu_step_operator.id,
    zenml_stack_component.cpu_step_operator.id,
  ]
}
```

## Argument Reference

* `name` - (Required) The name of the stack.
* `components` - (Required) A map where keys are component types and values are component IDs. Each component type can only have one component here; use `step_operators` to attach multiple step operators. Valid component types include:
  * `artifact_store`
  * `container_registry`
  * `orchestrator`
  * `step_operator`
  * `model_deployer`
  * `experiment_tracker`
  * `deployer`
  * `log_store`
  * `alerter`
  * `annotator`
  * `data_validator`
  * `feature_store`
  * `image_builder`
  * `model_registry`
  * `sandbox`
* `step_operators` - (Optional) A set of step operator component IDs. Use this to attach multiple step operators to the stack. Cannot be combined with `components.step_operator`; a single step operator can be configured either way.
* `labels` - (Optional) A map of labels to associate with the stack.

## Update Behavior

Updating stack labels, name, or component references is performed in place whenever supported by the ZenML API. This preserves the stack ID and avoids transient outages caused by deleting and recreating the stack during `terraform apply`.

## Step Operator Representation

When reading a stack (for example after `terraform import`), the provider reports its step operators in `components.step_operator` if the stack has exactly one step operator and `step_operators` is not configured. If the stack has more than one step operator, or `step_operators` is configured, all of them are reported in `step_operators`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack.

## Import

Stacks can be imported using the `id`, e.g.

```shell
$ terraform import zenml_stack.example 12345678-1234-1234-1234-123456789012
```
