---
page_title: "zenml_secret Resource - terraform-provider-zenml"
subcategory: ""
description: |-
  Manages a ZenML secret.
---

# zenml_secret (Resource)

Manages a ZenML secret whose values can be referenced by stack components and other ZenML resources.

> Secret values are stored in Terraform state. The `sensitive` designation hides them from normal CLI output but does not encrypt them. Use an encrypted, access-controlled remote state backend.

## Example Usage

```hcl
variable "databricks_client_id" {
  type      = string
  sensitive = true
}

variable "databricks_client_secret" {
  type      = string
  sensitive = true
}

resource "zenml_secret" "databricks" {
  name    = "databricks-oauth"
  private = false

  values = {
    client_id     = var.databricks_client_id
    client_secret = var.databricks_client_secret
  }
}

resource "zenml_stack_component" "databricks" {
  name   = "databricks-orchestrator"
  type   = "orchestrator"
  flavor = "databricks"

  configuration = {
    host          = "https://example.cloud.databricks.com"
    client_id     = format("{{%s.client_id}}", zenml_secret.databricks.name)
    client_secret = format("{{%s.client_secret}}", zenml_secret.databricks.name)
  }
}
```

## Argument Reference

* `name` - (Required) The unique name of the secret within its scope.
* `values` - (Required, Sensitive) A map of values stored in the secret. Removing a key from this map removes it from the ZenML secret.
* `private` - (Optional) Whether only the user that created the secret can access it. Defaults to `false`. Public secrets are generally preferable for IaC because private secrets are owned by the service account that runs Terraform.

The provider principal must have permission to create, update, and delete secrets, including `READ_SECRET_VALUE`. Terraform cannot refresh or import a secret when its values are redacted.

## Attributes Reference

* `id` - The secret ID.
* `user_id` - The ID of the user that owns the secret.
* `created` - The timestamp when the secret was created.
* `updated` - The timestamp when the secret was last updated.

## Import

Secrets can be imported by UUID:

```shell
terraform import zenml_secret.example 12345678-1234-1234-1234-123456789012
```

Import requires `READ_SECRET_VALUE` permission because all values must be loaded into Terraform state.
