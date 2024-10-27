# examples/data-sources/outputs.tf
output "existing_stack_components" {
  value = data.zenml_stack.existing.components
}

output "artifact_store_configuration" {
  value     = data.zenml_stack_component.artifact_store.configuration
  sensitive = true
}

output "gcp_connector_resource_types" {
  value = data.zenml_service_connector.gcp.resource_types
}