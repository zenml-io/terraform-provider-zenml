output "stack_id" {
  description = "ID of the created ZenML stack"
  value       = zenml_stack.azure_stack.id
}

output "stack_name" {
  description = "Name of the created ZenML stack"
  value       = zenml_stack.azure_stack.name
}

output "artifact_store_path" {
  description = "Azure Blobg storage path for the artifact store"
  value       = "az://${azurerm_storage_container.artifact_container.name}"
}

output "artifact_store_id" {
  description = "ID of the artifact store component"
  value       = zenml_stack_component.artifact_store.id
}

output "container_registry_uri" {
  description = "URI of the Azure Container Registry repository"
  value       = azurerm_container_registry.containers.login_server
}

output "container_registry_id" {
  description = "ID of the container registry component"
  value       = zenml_stack_component.container_registry.id
}

output "orchestrator_id" {
  description = "ID of the orchestrator component"
  value       = zenml_stack_component.orchestrator.id
}

output "principal_id" {
  description = "Application client ID of the service principal"
  value       = azuread_application.service_principal_app.client_id
}

output "service_connector_id" {
  description = "ID of the Azure service connector"
  value       = zenml_service_connector.azure.id
}