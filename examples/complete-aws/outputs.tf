output "stack_id" {
  description = "ID of the created ZenML stack"
  value       = zenml_stack.aws_stack.id
}

output "stack_name" {
  description = "Name of the created ZenML stack"
  value       = zenml_stack.aws_stack.name
}

output "artifact_store_path" {
  description = "S3 path for the artifact store"
  value       = "s3://${aws_s3_bucket.artifacts.bucket}/artifacts"
}

output "artifact_store_id" {
  description = "ID of the artifact store component"
  value       = zenml_stack_component.artifact_store.id
}

output "container_registry_uri" {
  description = "URI of the ECR repository"
  value       = aws_ecr_repository.containers.repository_url
}

output "container_registry_id" {
  description = "ID of the container registry component"
  value       = zenml_stack_component.container_registry.id
}

output "orchestrator_id" {
  description = "ID of the orchestrator component"
  value       = zenml_stack_component.orchestrator.id
}

output "role_arn" {
  description = "ARN of the IAM role created for ZenML"
  value       = aws_iam_role.zenml.arn
}

output "service_connector_id" {
  description = "ID of the AWS service connector"
  value       = zenml_service_connector.aws.id
}