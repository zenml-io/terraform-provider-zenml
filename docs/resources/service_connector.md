---
page_title: "zenml_service_connector Resource - terraform-provider-zenml"
subcategory: ""
description: |-
  Manages a ZenML service connector.
---

# zenml_service_connector (Resource)

Manages a ZenML service connector, which provides a way to connect to external services and resources.

## Example Usage

```hcl
resource "zenml_service_connector" "gcp_connector" {
  name        = "my-gcp-connector"
  type        = "gcp"
  auth_method = "service-account"
  user        = "user-uuid"
  workspace   = "workspace-uuid"
  
  resource_types = ["artifact-store", "container-registry"]
  
  configuration = {
    project_id = "my-gcp-project"
  }
  
  secrets = {
    service_account_json = jsonencode({
      "type": "service_account",
      "project_id": "my-gcp-project",
      # ... other service account details
    })
  }
}
```

## Argument Reference

* `name` - (Required) The name of the service connector.
* `type` - (Required) The type of the service connector (e.g., "gcp", "aws", "azure").
* `auth_method` - (Required) The authentication method used by the service connector.
* `user` - (Required) The ID of the user who owns this connector.
* `workspace` - (Required) The ID of the workspace this connector belongs to.
* `resource_types` - (Optional) A list of resource types this connector can be used for.
* `configuration` - (Required) A map of configuration key-value pairs for the connector.
* `secrets` - (Optional) A map of secret key-value pairs for the connector. These are sensitive and will not be output.
* `labels` - (Optional) A map of labels to associate with the connector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service connector.

## Import

Service connectors can be imported using the `id`, e.g.

```
$ terraform import zenml_service_connector.example 12345678-1234-1234-1234-123456789012
```