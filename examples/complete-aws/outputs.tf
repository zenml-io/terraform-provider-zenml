# examples/complete-aws/outputs.tf
output "stack_id" {
  value = zenml_stack.aws_stack.id
}

output "artifact_store_path" {
  value = "s3://${aws_s3_bucket.artifacts.bucket}/artifacts"
}

output "container_registry_uri" {
  value = aws_ecr_repository.containers.repository_url
}

output "role_arn" {
  value = aws_iam_role.zenml.arn
}