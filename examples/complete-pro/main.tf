terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
  }
}

provider "zenml" {
  control_plane_url   = var.zenml_control_plane_url
  service_account_key = var.zenml_service_account_key
}

# Create workspace via control plane
resource "zenml_workspace" "alpha" {
  name        = "alpha-workspace"
  description = "Workspace for Team Alpha recommendation engine project"
  
  tags = [
    "ml-team",
    "recommendation-engine",
    "gpu-enabled"
  ]
  
  metadata = {
    team        = "alpha"
    environment = "production"
    cost_center = "ml-engineering"
  }
}

# Teams are consistent across control plane and workspaces
resource "zenml_team" "alpha_dev" {
  control_plane_id = data.zenml_workspace.main.control_plane_id
  name             = "alpha-developers"
  description      = "Development team for Alpha project"
  
  members = [
    "dev1@company.com",
    "dev2@company.com"
  ]
}

resource "zenml_team" "alpha_ml" {
  control_plane_id = data.zenml_workspace.main.control_plane_id
  name             = "alpha-ml-engineers"
  description      = "ML engineers for Alpha project"
  
  members = [
    "ml1@company.com",
    "ml2@company.com"
  ]
}

resource "zenml_team" "alpha_ops" {
  control_plane_id = data.zenml_workspace.main.control_plane_id
  name             = "alpha-ml-ops"
  description      = "ML Ops team for Alpha project"
  
  members = [
    "ops1@company.com"
  ]
}

# Assign teams to workspace (control plane API)
resource "zenml_workspace_role_assignment" "teams" {
  for_each = {
    dev = { team_id = zenml_team.alpha_dev.id, role = "member" }
    ml  = { team_id = zenml_team.alpha_ml.id, role = "member" }
    ops = { team_id = zenml_team.alpha_ops.id, role = "admin" }
  }

  workspace_id = zenml_workspace.alpha.id
  team_id      = each.value.team_id
  role         = each.value.role
}

# Create project in workspace (workspace API)
resource "zenml_project" "recommendation" {
  workspace_id = zenml_workspace.alpha.id
  name         = "recommendation-engine"
  description  = "ML project for building recommendation systems"
  
  tags = [
    "recommendation",
    "deep-learning",
    "production"
  ]
  
  metadata = {
    ml_framework = "tensorflow"
    data_source  = "user-behavior"
    model_type   = "neural-collaborative-filtering"
  }
}

# Assign teams to project (workspace API, using team IDs)
resource "zenml_project_role_assignment" "teams" {
  for_each = {
    dev = { team_id = zenml_team.alpha_dev.id, role = "contributor" }
    ml  = { team_id = zenml_team.alpha_ml.id, role = "admin" }
    ops = { team_id = zenml_team.alpha_ops.id, role = "admin" }
  }

  project_id = zenml_project.recommendation.id
  team_id    = each.value.team_id
  role       = each.value.role
}

# AWS Service Connector for the workspace
resource "zenml_service_connector" "aws" {
  name        = "aws-alpha-connector"
  type        = "aws"
  auth_method = "iam-role"

  configuration = {
    region                = var.aws_region
    role_arn             = var.aws_role_arn
    aws_access_key_id     = var.aws_access_key_id
    aws_secret_access_key = var.aws_secret_access_key
  }

  labels = {
    environment = "production"
    team        = "alpha"
    managed_by  = "terraform"
  }
}

# Development stack - accessible to all teams
resource "zenml_stack" "dev" {
  name = "alpha-dev-stack"

  components = {
    orchestrator      = zenml_stack_component.orchestrator_dev.id
    artifact_store    = zenml_stack_component.artifact_store.id
    container_registry = zenml_stack_component.container_registry.id
  }

  labels = {
    environment = "development"
    team        = "alpha"
  }
}

# Production stack with restricted access
resource "zenml_stack" "prod" {
  name = "alpha-prod-stack"

  components = {
    orchestrator       = zenml_stack_component.orchestrator_prod.id
    artifact_store     = zenml_stack_component.artifact_store.id
    container_registry = zenml_stack_component.container_registry.id
    model_registry     = zenml_stack_component.model_registry.id
  }

  labels = {
    environment = "production"
    team        = "alpha"
  }
}

# Stack components
resource "zenml_stack_component" "artifact_store" {
  name   = "s3-alpha-artifacts"
  type   = "artifact_store"
  flavor = "s3"

  configuration = {
    path = "s3://${var.s3_bucket_name}/artifacts"
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = "shared"
    team        = "alpha"
  }
}

resource "zenml_stack_component" "container_registry" {
  name   = "ecr-alpha-registry"
  type   = "container_registry"
  flavor = "aws"

  configuration = {
    uri                = var.ecr_repository_url
    default_repository = "alpha-ml-images"
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = "shared"
    team        = "alpha"
  }
}

resource "zenml_stack_component" "orchestrator_dev" {
  name   = "kubernetes-dev"
  type   = "orchestrator"
  flavor = "kubernetes"

  configuration = {
    context                = "dev-cluster"
    kubernetes_namespace   = "zenml-alpha-dev"
    synchronous           = false
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = "development"
    team        = "alpha"
  }
}

resource "zenml_stack_component" "orchestrator_prod" {
  name   = "sagemaker-prod"
  type   = "orchestrator"
  flavor = "sagemaker"

  configuration = {
    region             = var.aws_region
    execution_role     = var.aws_role_arn
    output_data_s3_uri = "s3://${var.s3_bucket_name}/sagemaker"
    instance_type      = "ml.g4dn.xlarge"
    volume_size        = 100
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = "production"
    team        = "alpha"
    gpu_enabled = "true"
  }
}

resource "zenml_stack_component" "model_registry" {
  name   = "mlflow-prod"
  type   = "model_registry"
  flavor = "mlflow"

  configuration = {
    tracking_uri     = "https://mlflow.company.com"
    tracking_username = var.mlflow_username
    tracking_password = var.mlflow_password
  }

  labels = {
    environment = "production"
    team        = "alpha"
  }
}

# Stack permissions use teams, not individual users
resource "zenml_stack_role_assignment" "permissions" {
  for_each = {
    dev_dev   = { stack = zenml_stack.dev.id, team = zenml_team.alpha_dev.id, role = "write" }
    dev_ml    = { stack = zenml_stack.dev.id, team = zenml_team.alpha_ml.id, role = "write" }
    dev_ops   = { stack = zenml_stack.dev.id, team = zenml_team.alpha_ops.id, role = "admin" }
    prod_dev  = { stack = zenml_stack.prod.id, team = zenml_team.alpha_dev.id, role = "read" }
    prod_ml   = { stack = zenml_stack.prod.id, team = zenml_team.alpha_ml.id, role = "read" }
    prod_ops  = { stack = zenml_stack.prod.id, team = zenml_team.alpha_ops.id, role = "admin" }
  }

  stack_id = each.value.stack
  team_id  = each.value.team
  role     = each.value.role
}

# Data source to get control plane information
data "zenml_workspace" "main" {
  # This would typically reference an existing workspace or be provided
  name = "main"
}