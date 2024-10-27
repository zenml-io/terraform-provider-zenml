---
page_title: "Getting Started - ZenML Provider"
subcategory: ""
description: |-
  Getting started with the ZenML Provider
---

# Getting Started with the ZenML Provider

This guide will help you get started with the ZenML provider for Terraform.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) 0.13+
- A ZenML server and API key

## Provider Configuration

First, configure the ZenML provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    zenml = {
      source = "zenml/zenml"
    }
  }
}

provider "zenml" {
  # Configuration options
}
```

Set your ZenML server URL and API key using environment variables:

```sh
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

## Creating Resources

Here's an example of how to create a ZenML stack:

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
}

resource "zenml_stack" "my_stack" {
  name = "my-stack"
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
  }
}
```

## Next Steps

- Explore the [resources](/docs/resources) and [data sources](/docs/data-sources) provided by the ZenML provider.
- Check out the [examples](https://github.com/zenml-io/terraform-provider-zenml/tree/main/examples) in the provider repository.
- Read the [ZenML documentation](https://docs.zenml.io/) to learn more about ZenML concepts and features.