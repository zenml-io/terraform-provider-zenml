---
page_title: "zenml_server Data Source - terraform-provider-zenml"
subcategory: ""
description: |-
  Data source for retrieving information about the ZenML server.
---

# zenml_server (Data Source)

Use this data source to retrieve information about the ZenML server.

## Example Usage

```hcl
data "zenml_server" "server_info" {
}

output "version" {
  value = data.zenml_server.server_info.version
}

output "dashboard_url" {
  value = data.zenml_server.server_info.dashboard_url
}
```

## Argument Reference

The `zenml_server` data source does not take any arguments.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the server.
* `name` - The name of the server.
* `version` - The version of the server.
* `deployment_type` - The deployment type of the server.
* `auth_scheme` - The authentication scheme of the server.
* `server_url` - The URL where the server API is hosted.
* `dashboard_url` - The URL where the server's dashboard is hosted.
* `pro_dashboard_url` - The URL where the server's Pro dashboard is hosted. Only set if the server is a ZenML Pro server.
* `pro_api_url` - The URL where the server's Pro API is hosted. Only set if the server is a ZenML Pro server.
* `pro_organization_id` - The ID of the ZenML Pro organization to which the server is connected. Only set if the server is a ZenML Pro server.
* `pro_organization_name` - The name of the ZenML Pro organization to which the server is connected. Only set if the server is a ZenML Pro server.
* `pro_workspace_id` - The ID of the ZenML Pro workspace to which the server is connected. Only set if the server is a ZenML Pro server.
* `pro_workspace_name` - The name of the ZenML Pro workspace to which the server is connected. Only set if the server is a ZenML Pro server.
* `metadata` - A map of metadata associated with the server.

## Import

The server info can be imported using the `id`, e.g.

```
$ terraform import zenml_server.server_info 12345678-1234-1234-1234-123456789012
```