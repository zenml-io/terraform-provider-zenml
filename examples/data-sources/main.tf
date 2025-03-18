# examples/data-sources/main.tf
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
  }
}

provider "zenml" {
  server_url = var.zenml_server_url
  api_key    = var.zenml_api_key
}

# Look up an existing stack by name
data "zenml_stack" "existing" {
  name      = "production-stack"
}

# Look up a stack component
data "zenml_stack_component" "artifact_store" {
  name      = "production-artifact-store"
  type      = "artifact_store"
}

# Look up a service connector
data "zenml_service_connector" "gcp" {
  name      = "production-gcp"
}

# Use the existing resources to create a new stack
resource "zenml_stack" "new_stack" {
  name = "new-stack"

  components = {
    artifact_store = data.zenml_stack_component.artifact_store.id
  }

  labels = {
    copied_from = data.zenml_stack.existing.name
  }
}