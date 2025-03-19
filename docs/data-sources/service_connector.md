
---
page_title: "zenml_service_connector Data Source - terraform-provider-zenml"
subcategory: ""
description: |-
  Data source for retrieving information about a ZenML service connector.
---

# zenml_service_connector (Data Source)

Use this data source to retrieve information about a specific ZenML service connector.

## Example Usage

```hcl
data "zenml_service_connector" "example" {
  name = "my-gcp-connector"
}

output "connector_id" {
  value = data.zenml_service_connector.example.id
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the service connector to retrieve. Either `id` or `name` must be provided.
* `name` - (Optional) The name of the service connector to retrieve. Either `id` or `name` must be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service connector.
* `name` - The name of the service connector.
* `type` - The type of the service connector (e.g., "gcp", "aws", "azure", etc.).
* `auth_method` - The authentication method used by the service connector.
* `resource_type` - The type of resource the service connector is connected to (e.g., "s3-bucket", "docker-registry", etc.).
* `resource_id` - The ID of the resource the service connector is connected to.
* `configuration` - A map of configuration key-value pairs for the service connector. Sensitive values are not included.
* `labels` - A map of labels associated with this service connector.

## Import

Service connectors can be imported using the `id`, e.g.

```
$ terraform import zenml_service_connector.example 12345678-1234-1234-1234-123456789012
```