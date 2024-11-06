---
page_title: "ZenML Provider"
subcategory: ""
description: |-
  The ZenML provider allows you to manage ZenML resources using Terraform.
---

# ZenML Provider

The ZenML provider allows you to manage [ZenML](https://zenml.io) resources using Terraform. It enables Infrastructure as Code management of:
- ZenML Stacks
- Stack Components
- Service Connectors

## Example Usage

```hcl
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
  }
}

# Configure the provider
provider "zenml" {
  # Configuration options will be loaded from environment variables:
  # ZENML_SERVER_URL
  # ZENML_API_KEY
}

# Create a service connector
resource "zenml_service_connector" "gcp" {
  name        = "gcp-connector"
  type        = "gcp"
  auth_method = "service-account"
  
  resource_types = ["artifact-store", "container-registry"]
  
  configuration = {
    project_id = "my-project"
  }
  
  secrets = {
    service_account_json = file("service-account.json")
  }
}

# Create a stack component
resource "zenml_stack_component" "artifact_store" {
  name      = "gcs-store"
  type      = "artifact_store"
  flavor    = "gcp"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
  
  connector_id = zenml_service_connector.gcp.id
}

# Create a stack
resource "zenml_stack" "production" {
  name = "production-stack"
  
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
  }
  
  labels = {
    environment = "production"
  }
}
```

## Authentication

The provider needs to be configured with a ZenML server URL and API key. These can be provided in several ways:

1. Using environment variables:
```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

2. Using provider configuration:
```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

### Generating an API Key

To generate a ZENML_API_KEY, follow these steps:

1. Install ZenML:
```bash
pip install zenml
```

2. Connect to your ZenML server:
```bash
zenml connect --url <API_URL>
```

3. Create a service account and get the API key:
```bash
zenml service-account create <MYSERVICEACCOUNTNAME>
```

This command will print out the ZENML_API_KEY that you can use with this provider.

## Provider Arguments

* `server_url` - (Optional) The URL of your ZenML server. Can be set with the `ZENML_SERVER_URL` environment variable.
* `api_key` - (Optional) Your ZenML API key. Can be set with the `ZENML_API_KEY` environment variable.
* `api_token` - (Optional) Your ZenML API token. Can be set with the `ZENML_API_TOKEN` environment variable.

## Resources

* [zenml_service_connector](resources/service_connector.md) - Manages service connectors for external services
* [zenml_stack_component](resources/stack_component.md) - Manages stack components
* [zenml_stack](resources/stack.md) - Manages stacks

## Data Sources

* [zenml_server](data-sources/server.md) - Retrieve information about the ZenML server
* [zenml_service_connector](data-sources/service_connector.md) - Retrieve information about a service connector
* [zenml_stack_component](data-sources/stack_component.md) - Retrieve information about a stack component
* [zenml_stack](data-sources/stack.md) - Retrieve information about a stack
