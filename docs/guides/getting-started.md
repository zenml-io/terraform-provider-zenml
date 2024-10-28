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
- A ZenML server
- An API key for authentication

## Provider Configuration

Configure the ZenML provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    zenml = {
      source = "zenml/zenml"
    }
  }
}

provider "zenml" {
  # Configuration options will be loaded from environment variables:
  # ZENML_SERVER_URL
  # ZENML_API_KEY
}
```

Set your environment variables:

```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

## Basic Example

Here's a simple example creating a stack with a component:

```hcl
# Create a stack component
resource "zenml_stack_component" "artifact_store" {
  workspace = "default"
  name      = "my-artifact-store"
  type      = "artifact_store"
  flavor    = "local"
  
  configuration = {
    path = "/path/to/artifacts"
  }
}

# Create a stack using the component
resource "zenml_stack" "example" {
  name = "my-stack"
  
  components = {
    "artifact_store" = zenml_stack_component.artifact_store.id
  }
  
  labels = {
    environment = "dev"
  }
}
```

## Next Steps

- Explore the [Stack Component](/docs/resources/stack_component) resource
- Learn about [Stack](/docs/resources/stack) configuration
- Check out [Service Connectors](/docs/resources/service_connector) for external services