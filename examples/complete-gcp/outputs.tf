
# examples/complete-gcp/outputs.tf
output "stack_id" {
  value       = zenml_stack.gcp_stack.id
  description = "The ID of the created ZenML stack"
}

output "artifact_store_path" {
  value       = "gs://${google_storage_bucket.artifacts.name}/artifacts"
  description = "GCS path for artifacts"
}

output "container_registry_uri" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
  description = "GCR URI for container images"
}