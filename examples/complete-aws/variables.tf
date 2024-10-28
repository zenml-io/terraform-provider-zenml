# examples/complete-aws/variables.tf
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
  default     = "us-west-2"
  description = "AWS region to deploy resources in"
}

variable "environment" {
  type        = string
  default     = "dev"
  description = "Environment name (e.g. dev, staging, prod)"
}

variable "name_prefix" {
  type        = string
  description = "Prefix for resource names"
}

variable "workspace" {
  type        = string
  default     = "default"
  description = "Name of the ZenML workspace"
}