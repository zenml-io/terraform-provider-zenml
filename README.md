# Terraform Provider for ZenML

[![Tests](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/test.yml/badge.svg)](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/test.yml)
[![Release](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/release.yml/badge.svg)](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/release.yml)

This Terraform provider allows you to manage ZenML resources using Infrastructure as Code. It provides the ability to manage:
- ZenML Stacks
- Stack Components
- Service Connectors

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20
- [ZenML Server](https://docs.zenml.io/) >= 0.40.0

## Building The Provider

1. Clone the repository
```bash
git clone git@github.com:zenml-io/terraform-provider-zenml.git
```

2. Build the provider
```bash
make build
```

## Installing The Provider

To use the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
  }
}
```

## Using the Provider

### Authentication

#### Service Account API Key

Configure the provider with your ZenML server URL and API key:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

You can also use environment variables:
```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"
```

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

#### API Token

Alternatively, you can use an API token for authentication:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_token  = "your-api-token"
}
```

You can also use environment variables:
```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_TOKEN="your-api-token"
```

### Example Usage

> **Hint:** The ZenML Terraform provider is being heavily used in all our Terraform modules. Their code is available on GitHub and can be used as a reference:
> - [zenml-stack/aws](https://github.com/zenml-io/terraform-aws-zenml-stack)
> - [zenml-stack/gcp](https://github.com/zenml-io/terraform-gcp-zenml-stack)
> - [zenml-stack/azure](https://github.com/zenml-io/terraform-azure-zenml-stack)

Here's a basic example of creating a stack with components:

```hcl
# Create a service connector for GCP
resource "zenml_service_connector" "gcp" {
  name        = "gcp-connector"
  type        = "gcp"
  auth_method = "service-account"
  # workspace defaults to "default" if not specified
  
  configuration = {
    project_id = "my-project"
    location   = "us-central1"
    service_account_json = file("service-account.json")
  }
  
  labels = {
    environment = "production"
  }
}

# Create an artifact store component
resource "zenml_stack_component" "artifact_store" {
  name   = "gcs-store"
  type   = "artifact_store"
  flavor = "gcp"
  # workspace defaults to "default" if not specified
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
  
  connector_id = zenml_service_connector.gcp.id
  
  labels = {
    environment = "production"
  }
}

# Create a stack using the components
resource "zenml_stack" "ml_stack" {
  name = "production-stack"
  # workspace defaults to "default" if not specified
  
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
  }
  
  labels = {
    environment = "production"
  }
}
```

> **Note:** All resources support an optional `workspace` parameter that defaults to "default" if not specified. You can override this by setting `workspace = "your-workspace-name"` in any resource.

See the [examples](./examples/) directory for more complete examples.

## Development

### Requirements

- [Go](https://golang.org/doc/install) >= 1.20
- [GNU Make](https://www.gnu.org/software/make/)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

### Building

1. Clone the repository
```bash
git clone git@github.com:zenml-io/terraform-provider-zenml.git
cd terraform-provider-zenml
```

2. Build the provider
```bash
make build
```

### Testing

Run unit tests:
```bash
make test
```

Run acceptance tests (requires a running ZenML server):
```bash
make testacc
```

### Local Development Installation

To install the provider locally for testing:

```bash
make install
```

This will build and install the provider to your local Terraform plugins directory.

### Documentation

Generate provider documentation:

```bash
make docs
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on contributing to this provider.

## Resource Documentation

### Stacks
- [Stack Resource](./docs/resources/stack.md)
- [Stack Data Source](./docs/data-sources/stack.md)

### Stack Components
- [Stack Component Resource](./docs/resources/stack_component.md)
- [Stack Component Data Source](./docs/data-sources/stack_component.md)

### Service Connectors
- [Service Connector Resource](./docs/resources/service_connector.md)
- [Service Connector Data Source](./docs/data-sources/service_connector.md)

## License

[Apache License 2.0](./LICENSE)