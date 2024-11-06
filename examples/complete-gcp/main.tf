# examples/complete-gcp/main.tf
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "zenml" {
  server_url = var.zenml_server_url
  api_key    = var.zenml_api_key
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Create GCP resources if needed

resource "google_service_account" "zenml_sa" {
  account_id   = "zenml-${var.environment}"
  display_name = "ZenML Service Account"
}


resource "google_service_account_key" "zenml_sa_key" {
  service_account_id = google_service_account.zenml_sa.name
}

resource "google_storage_bucket" "artifacts" {
  name     = "${var.project_id}-zenml-artifacts-${var.environment}"
  location = var.region
}

resource "google_artifact_registry_repository" "containers" {
  location      = var.region
  repository_id = "zenml-containers-${var.environment}"
  format        = "DOCKER"
}

# ZenML Service Connector for GCP
resource "zenml_service_connector" "gcp" {
  name        = "gcp-${var.environment}"
  type        = "gcp"
  auth_method = "service-account"

  configuration = {
    project_id = var.project_id
    region     = var.region
    service_account_json = "${google_service_account_key.zenml_sa_key.0.private_key}"
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# Artifact Store Component
resource "zenml_stack_component" "artifact_store" {
  name   = "gcs-${var.environment}"
  type   = "artifact_store"
  flavor = "gcp"

  configuration = {
    path = "gs://${google_storage_bucket.artifacts.name}/artifacts"
  }

  connector_id = zenml_service_connector.gcp.id

  labels = {
    environment = var.environment
  }
}

# Container Registry Component
resource "zenml_stack_component" "container_registry" {
  name   = "gcr-${var.environment}"
  type   = "container_registry"
  flavor = "gcp"

  configuration = {
    uri = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.containers.repository_id}"
  }

  connector_id = zenml_service_connector.gcp.id

  labels = {
    environment = var.environment
  }
}

# Vertex AI Orchestrator
resource "zenml_stack_component" "orchestrator" {
  name   = "vertex-${var.environment}"
  type   = "orchestrator"
  flavor = "vertex"

  configuration = {
    location    = var.region
    synchronous = true
  }

  connector_id = zenml_service_connector.gcp.id

  labels = {
    environment = var.environment
  }
}

# Complete Stack
resource "zenml_stack" "gcp_stack" {
  name = "gcp-${var.environment}"

  components = {
    artifact_store     = zenml_stack_component.artifact_store.id
    container_registry = zenml_stack_component.container_registry.id
    orchestrator      = zenml_stack_component.orchestrator.id
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}