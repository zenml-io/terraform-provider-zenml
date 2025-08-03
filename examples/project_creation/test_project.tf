terraform {
  required_providers {
    zenml = {
      source = "zenml/zenml"
    }
  }
}

provider "zenml" {
  server_url = var.zenml_server_url
  api_key    = var.zenml_api_key
  api_token  = var.zenml_api_token
}

resource "zenml_project" "test_project" {
  name         = var.project_name
  display_name = var.project_display_name
  description  = var.project_description
}

output "project_id" {
  value = zenml_project.test_project.id
}

output "project_created" {
  value = zenml_project.test_project.created
}