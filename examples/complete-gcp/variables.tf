# examples/complete-gcp/variables.tf
variable "zenml_server_url" {
  type        = string
  description = "URL of the ZenML server"
}

variable "zenml_api_key" {
  type        = string
  description = "API key for ZenML server"
  sensitive   = true
}

variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region"
  default     = "us-central1"
}

variable "environment" {
  type        = string
  description = "Environment name (e.g. dev, staging, prod)"
  default     = "dev"
}

variable "user_id" {
  type        = string
  description = "ZenML user ID"
}

variable "workspace_id" {
  type        = string
  description = "ZenML workspace ID"
}

variable "gcp_service_account_key" {
  type        = string
  description = "GCP service account key JSON"
  sensitive   = true
}
