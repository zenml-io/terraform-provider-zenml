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
      "type": "service-account",
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
* `type` - (Required, Forces new resource) The type of the service connector. Valid values include: `aws`, `gcp`, `azure`, and others depending on your ZenML version. You can run `zenml service-connector list-types` to get the list of available types or take a look at the [ZenML documentation](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management/service-connectors-guide#explore-service-connector-types) for more information.
* `auth_method` - (Required, Forces new resource) The authentication method used by the connector. Valid values include:
  * AWS: `iam-role`, `secret-key`, `implicit`, `sts-token`, `session-token` or `federation-token`. Run `zenml service-connector describe-type aws` or visit the [AWS Service Connector ZenML documentation page](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management/aws-service-connector) for more information.
  * GCP: `service-account`, `external-account`, `user-account`, `implicit`, `oauth2-token` or `impersonation`. Run `zenml service-connector describe-type gcp` or visit the [GCP Service Connector ZenML documentation page](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management/gcp-service-connector) for more information.
  * Azure: `service-principal`, `access-token` or `implicit`. Run `zenml service-connector describe-type azure` or visit the [Azure Service Connector ZenML documentation page](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management/azure-service-connector) for more information.
  * Kubernetes: `password` or `token`. Run `zenml service-connector describe-type kubernetes` or visit the [Kubernetes Service Connector ZenML documentation page](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management/kubernetes-service-connector) for more information.
* `workspace` - (Optional) The workspace this connector belongs to. Defaults to "default". Forces new resource if changed.
* `resource_type` - (Optional) A resource type this connector can be used for (e.g., `s3-bucket`, `kubernetes-cluster`, `docker-registry`). To find out which resource types are supported by a connector, run `zenml service-connector describe-type <connector-type>`.
* `configuration` - (Required, Sensitive) A map of configuration key-value pairs for the connector. Every authentication method has its own set of required and optional configuration parameters. To find out which parameters are required and optional for a given authentication method, run `zenml service-connector describe-type <connector-type> -a <auth-method>` or visit the [Service Connector ZenML documentation page](https://docs.zenml.io/how-to/infrastructure-deployment/auth-management) for the connector type and authentication method for more information.
* `labels` - (Optional) A map of labels to associate with the connector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service connector.

## Import

Service connectors can be imported using the `id`, e.g.

```
$ terraform import zenml_service_connector.example 12345678-1234-1234-1234-123456789012
```