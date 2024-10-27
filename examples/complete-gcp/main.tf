terraform {
  required_providers {
    zenml = {
      source = "zenml/zenml"
    }
  }
}

provider "zenml" {
  server_url = "https://your-zenml-server"
  api_key    = var.zenml_api_key
}