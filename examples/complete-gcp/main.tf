terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    zenml = {
        source = "zenml-io/zenml"
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

data "google_client_config" "current" {}
data "google_project" "project" {
  project_id = data.google_client_config.current.project
}

# Enable required APIs
resource "google_project_service" "common_services" {
  for_each = toset([
    "iam.googleapis.com",
    "artifactregistry.googleapis.com",
    "storage-api.googleapis.com",
    "aiplatform.googleapis.com",
  ])
  service = each.key
  disable_on_destroy = false
}

# Create GCS bucket for ZenML artifacts

resource "google_storage_bucket" "artifacts" {
  name     = "${var.project_id}-zenml-artifacts-${var.environment}"
  location = var.region
  depends_on = [google_project_service.common_services]
  force_destroy = true
}

# Create Artifact Registry repository for ZenML containers

resource "google_artifact_registry_repository" "containers" {
  location      = var.region
  repository_id = "zenml-containers-${var.environment}"
  format        = "DOCKER"
  depends_on    = [google_project_service.common_services]
}

# Create service account with required permissions and key

resource "google_service_account" "zenml_sa" {
  account_id   = "zenml-${var.environment}"
  display_name = "ZenML Service Account"
}

resource "google_service_account_key" "zenml_sa_key" {
  service_account_id = google_service_account.zenml_sa.name
}

resource "google_project_iam_member" "storage_object_user" {
  project = data.google_client_config.current.project
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.zenml_sa.email}"

  condition {
    title       = "Restrict access to the ZenML bucket"
    description = "Grants access only to the ZenML bucket"
    expression  = "resource.name.startsWith('projects/_/buckets/${google_storage_bucket.artifacts.name}')"
  }
}

resource "google_project_iam_member" "artifact_registry_writer" {
  project = data.google_client_config.current.project
  role    = "roles/artifactregistry.createOnPushWriter"
  member  = "serviceAccount:${google_service_account.zenml_sa.email}"

  condition {
    title       = "Restrict access to the ZenML container registry"
    description = "Grants access only to the ZenML container registry"
    expression  = "resource.name.startsWith('projects/${data.google_project.project.number}/locations/${data.google_client_config.current.region}/repositories/${google_artifact_registry_repository.containers.repository_id}')"
  }
}

resource "google_project_iam_member" "ai_platform_service_agent" {
  project = data.google_client_config.current.project
  role    = "roles/aiplatform.serviceAgent"
  member  = "serviceAccount:${google_service_account.zenml_sa.email}"
}


# ZenML Service Connector for GCP
resource "zenml_service_connector" "gcp" {
  name        = "gcp-${var.environment}"
  type        = "gcp"
  auth_method = "service-account"

  configuration = {
    project_id = var.project_id
    region     = var.region
    service_account_json = google_service_account_key.zenml_sa_key.private_key
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