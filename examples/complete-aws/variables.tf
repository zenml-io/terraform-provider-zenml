# examples/complete-aws/variables.tf
variable "zenml_server_url" {
  type = string
}

variable "zenml_api_key" {
  type      = string
  sensitive = true
}

variable "region" {
  type    = string
  default = "us-west-2"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "name_prefix" {
  type        = string
  description = "Prefix for resource names"
}

variable "user_id" {
  type = string
}

variable "workspace_id" {
  type = string
}