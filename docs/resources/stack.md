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
resource "zenml_stack_component" "artifact_store" {
  name      = "my-artifact-store"
  type      = "artifact_store"
  flavor    = "gcp"
  user      = "user-uuid"
  workspace = "workspace-uuid"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
}

resource "zenml_stack" "my_stack" {
  name = "my-stack"
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
  }
  
  labels = {
    environment = "production"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the stack.
* `components` - (Required) A map of component types to component IDs that make up this stack.
* `labels` - (Optional) A map of labels to associate with the stack.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack.

## Import

Stacks can be imported using the `id`, e.g.

```
$ terraform import zenml_stack.example 12345678-1234-1234-1234-123456789012
```
