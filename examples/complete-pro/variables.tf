variable "zenml_control_plane_url" {
  description = "URL of the ZenML control plane"
  type        = string
  default     = "https://cloudapi.zenml.io"
}

variable "zenml_service_account_key" {
  description = "Service account key for ZenML Pro authentication"
  type        = string
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

variable "aws_role_arn" {
  description = "ARN of the AWS IAM role for ZenML"
  type        = string
}

variable "aws_access_key_id" {
  description = "AWS access key ID"
  type        = string
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS secret access key"
  type        = string
  sensitive   = true
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for artifacts"
  type        = string
}

variable "ecr_repository_url" {
  description = "URL of the ECR repository"
  type        = string
}

variable "mlflow_username" {
  description = "Username for MLflow tracking server"
  type        = string
  sensitive   = true
}

variable "mlflow_password" {
  description = "Password for MLflow tracking server"
  type        = string
  sensitive   = true
}