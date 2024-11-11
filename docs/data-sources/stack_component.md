---
page_title: "zenml_stack_component Data Source - terraform-provider-zenml"
subcategory: ""
description: |-
  Data source for retrieving information about a ZenML stack component.
---

# zenml_stack_component (Data Source)

Use this data source to retrieve information about a specific ZenML stack component.

## Example Usage

```hcl
data "zenml_stack_component" "example" {
  name = "my-artifact-store"
}

output "component_id" {
  value = data.zenml_stack_component.example.id
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the stack component to retrieve. Either `id` or `name` must be provided.
* `name` - (Optional) The name of the stack component to retrieve. Either `id` or `name` must be provided.
* `workspace` - (Optional) The workspace ID to filter the component search. If not provided, the default workspace will be used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack component.
* `name` - The name of the stack component.
* `type` - The type of the stack component (e.g., "artifact_store", "orchestrator", etc.).
* `flavor` - The flavor of the stack component (e.g., "local", "gcp", "aws", etc.).
* `configuration` - A map of configuration key-value pairs for the stack component.
* `workspace` - The workspace ID this stack component belongs to.
* `labels` - A map of labels associated with this stack component.
* `connector` - The ID of the service connector associated with this stack component.
* `connector_resource_id` - The ID of the resource the service connector is connected to.

## Import

Stack components can be imported using the `id`, e.g.

```
$ terraform import zenml_stack_component.example 12345678-1234-1234-1234-123456789012
```