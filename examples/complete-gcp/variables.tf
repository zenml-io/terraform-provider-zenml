# examples/complete-gcp/variables.tf
variable "zenml_server_url" {
  type        = string
  description = "URL of the ZenML server"
  default     = "http://127.0.0.1:8237"
}

variable "zenml_api_key" {
  type        = string
  description = "API key for ZenML server"
  sensitive   = true
  default     = "ZENKEY_eyJpZCI6ImNiZDYxZDE1LWM3ZGEtNDdjMS1hNjg2LWRjYWZjYWMxYmNiMSIsImtleSI6IjExZTI4ZGY5M2RiNWJjMmVjNDI4ZmVjZTBjNTNlYTRkNjY5NGViOGU4ZGJhNTVkNmI2ZTFkMTc4NDU4MzY2OGUifQ=="
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

variable "gcp_service_account_key" {
  type        = string
  description = "GCP service account key JSON"
  sensitive   = true
}
