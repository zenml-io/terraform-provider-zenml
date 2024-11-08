variable "zenml_server_url" {
  type        = string
  description = "URL of the ZenML server"
}

variable "zenml_api_key" {
  type        = string
  sensitive   = true
  description = "API key for authenticating with the ZenML server"
}

variable "location" {
  description = "The Azure region where resources will be created"
  # Make a choice from the list of Azure regions
  type        = string
  default     = "westus"
}

variable "environment" {
  type        = string
  default     = "dev"
  description = "Environment name (e.g. dev, staging, prod)"
}
