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
  workspace   = "default"
  
  configuration = {
    project_id = "my-gcp-project"
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
* `workspace` - (Optional) The workspace this connector belongs to. Defaults to "default". Forces new resource if changed.
* `resource_type` - (Optional) A resource type this connector can be used for (e.g., `s3-bucket`, `kubernetes-cluster`, `docker-registry`).
* `configuration` - (Required, Sensitive) A map of configuration key-value pairs for the connector.
* `labels` - (Optional) A map of labels to associate with the connector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service connector.

## Import

Service connectors can be imported using the `id`, e.g.

```
$ terraform import zenml_service_connector.example 12345678-1234-1234-1234-123456789012
```