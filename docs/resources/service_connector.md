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
      "private_key_id": "key-id",
      "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
      "client_email": "service-account@project.iam.gserviceaccount.com",
      "client_id": "client-id",
      "auth_uri": "https://accounts.google.com/o/oauth2/auth",
      "token_uri": "https://oauth2.googleapis.com/token"
    })
  }
  
  labels = {
    environment = "production"
    team        = "ml-ops"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the service connector.
* `type` - (Required, Forces new resource) The type of the service connector. Valid values include: `aws`, `gcp`, `azure`, and others depending on your ZenML version.
* `auth_method` - (Required, Forces new resource) The authentication method used by the connector. Valid values include:
  * AWS: `iam-role`, `aws-access-keys`, `web-identity`
  * GCP: `service-account`, `oauth2`, `workload-identity`
  * Azure: `service-principal`, `managed-identity`
  * Kubernetes: `kubeconfig`, `service-account`
* `user` - (Required, Forces new resource) The ID of the user who owns this connector.
* `workspace` - (Required, Forces new resource) The ID of the workspace this connector belongs to.
* `resource_types` - (Optional) A list of resource types this connector can be used for (e.g., `artifact-store`, `container-registry`, `orchestrator`).
* `configuration` - (Required, Sensitive) A map of configuration key-value pairs for the connector.
* `secrets` - (Optional, Sensitive) A map of secret key-value pairs for the connector.
* `labels` - (Optional) A map of labels to associate with the connector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service connector.

## Import

Service connectors can be imported using the `id`, e.g.

```
$ terraform import zenml_service_connector.example 12345678-1234-1234-1234-123456789012
```