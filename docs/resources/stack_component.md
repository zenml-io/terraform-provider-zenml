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
  user      = "user-uuid"
  workspace = "workspace-uuid"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
  
  labels = {
    environment = "production"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the stack component.
* `type` - (Required) The type of the stack component (e.g., "artifact_store", "orchestrator").
* `flavor` - (Required) The flavor of the stack component (e.g., "local", "gcp", "aws").
* `user` - (Required) The ID of the user who owns this component.
* `workspace` - (Required) The ID of the workspace this component belongs to.
* `configuration` - (Required) A map of configuration key-value pairs for the component.
* `connector_resource_id` - (Optional) The ID of the connector resource to use with this component.
* `labels` - (Optional) A map of labels to associate with the component.
* `component_spec_path` - (Optional) The path to the component specification file.
* `connector` - (Optional) The ID of the service connector to use with this component.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack component.

## Import

Stack components can be imported using the `id`, e.g.

```
$ terraform import zenml_stack_component.example 12345678-1234-1234-1234-123456789012
```
