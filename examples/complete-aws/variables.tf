variable "zenml_server_url" {
  type        = string
  description = "URL of the ZenML server"
}

variable "zenml_api_key" {
  type        = string
  sensitive   = true
  description = "API key for authenticating with the ZenML server"
}

variable "region" {
  type        = string
  default     = "eu-central-1"
  description = "AWS region to deploy resources in"
}

variable "environment" {
  type        = string
  default     = "dev"
  description = "Environment name (e.g. dev, staging, prod)"
}
