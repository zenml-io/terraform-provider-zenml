
output "stack_id" {
  value       = zenml_stack.gcp_stack.id
  description = "The ID of the created ZenML stack"
}

output "stack_name" {
  description = "Name of the created ZenML stack"
  value       = zenml_stack.gcp_stack.name
}

output "artifact_store_path" {
  value       = "gs://${google_storage_bucket.artifacts.name}/artifacts"
  description = "GCS path for artifacts"
}

output "artifact_store_id" {
  description = "ID of the artifact store component"
  value       = zenml_stack_component.artifact_store.id
}

output "container_registry_uri" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
  description = "GCR URI for container images"
}

output "container_registry_id" {
  description = "ID of the container registry component"
  value       = zenml_stack_component.container_registry.id
}

output "orchestrator_id" {
  description = "ID of the orchestrator component"
  value       = zenml_stack_component.orchestrator.id
}

output "service_connector_id" {
  description = "ID of the AWS service connector"
  value       = zenml_service_connector.gcp.id
}