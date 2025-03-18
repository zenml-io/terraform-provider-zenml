# examples/data-sources/variables.tf
variable "zenml_server_url" {
  type = string
}

variable "zenml_api_key" {
  type      = string
  sensitive = true
}
