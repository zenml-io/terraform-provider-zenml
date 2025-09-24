variable "zenml_server_url" {
  description = "The URL of your ZenML server"
  type        = string
  default     = "https://your-zenml-server.com"
}

variable "zenml_api_key" {
  description = "Your ZenML API key (for service accounts)"
  type        = string
  sensitive   = true
  default     = null
}

variable "zenml_api_token" {
  description = "Your ZenML API token (JWT token)"
  type        = string
  sensitive   = true
  default     = null
}

variable "project_name" {
  description = "The name of the project to create"
  type        = string
  default     = "my-terraform-project"
}

variable "project_display_name" {
  description = "The display name of the project"
  type        = string
  default     = "My Terraform Project"
}

variable "project_description" {
  description = "The description of the project"
  type        = string
  default     = "A project managed by Terraform"
}