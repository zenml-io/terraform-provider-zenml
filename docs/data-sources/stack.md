---
page_title: "zenml_stack Data Source - terraform-provider-zenml"
subcategory: ""
description: |-
  Data source for retrieving information about a ZenML stack.
---

# zenml_stack (Data Source)

Use this data source to retrieve information about a specific ZenML stack.

## Example Usage

```hcl
data "zenml_stack" "example" {
  name = "my-stack"
}

output "stack_id" {
  value = data.zenml_stack.example.id
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the stack to retrieve. Either `id` or `name` must be provided.
* `name` - (Optional) The name of the stack to retrieve. Either `id` or `name` must be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the stack.
* `name` - The name of the stack.
* `components` - A map of component types to component IDs for this stack.
* `labels` - A map of labels associated with this stack.

## Import

Stacks can be imported using the `id`, e.g.

```
$ terraform import zenml_stack.example 12345678-1234-1234-1234-123456789012
```