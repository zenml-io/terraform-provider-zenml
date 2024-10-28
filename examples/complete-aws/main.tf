# examples/complete-aws/main.tf
terraform {
  required_providers {
    zenml = {
      source = "zenml/zenml"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "zenml" {
  server_url = var.zenml_server_url
  api_key    = var.zenml_api_key
}

provider "aws" {
  region = var.region
}

# Create AWS resources if needed
resource "aws_s3_bucket" "artifacts" {
  bucket = "${var.name_prefix}-zenml-artifacts"
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_ecr_repository" "containers" {
  name = "${var.name_prefix}-zenml-containers"
}

# IAM Role for ZenML
resource "aws_iam_role" "zenml" {
  name = "${var.name_prefix}-zenml-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = [
            "sagemaker.amazonaws.com",
            "lambda.amazonaws.com"
          ]
        }
      }
    ]
  })
}

# ZenML Service Connector for AWS
resource "zenml_service_connector" "aws" {
  name           = "aws-${var.environment}"
  type           = "aws"
  auth_method    = "iam-role"
  user           = var.user_id
  workspace      = var.workspace_id

  resource_types = [
    "artifact-store",
    "container-registry",
    "orchestrator",
    "step-operator"
  ]

  configuration = {
    region   = var.region
    role_arn = aws_iam_role.zenml.arn
  }

  labels = {
    environment = var.environment
    managed_by  = "terraform"
  }
}

# Artifact Store Component
resource "zenml_stack_component" "artifact_store" {
  name      = "s3-${var.environment}"
  type      = "artifact_store"
  flavor    = "s3"
  user      = var.user_id
  workspace = var.workspace_id

  configuration = {
    path = "s3://${aws_s3_bucket.artifacts.bucket}/artifacts"
  }

  connector = zenml_service_connector.aws.id

  labels = {
    environment = var.environment
  }
}

# Container Registry Component
resource "zenml_stack_component" "container_registry" {
  name      = "ecr-${var.environment}"
  type      = "container_registry"
  flavor    = "aws"
  user      = var.user_id
  workspace = var.workspace_id

  configuration = {
    uri = aws_ecr_repository.containers.repository_url
  }

  connector = zenml_service_connector.aws.id

  labels = {
    environment = var.environment
  }
}

# SageMaker Orchestrator
resource "zenml_stack_component" "orchestrator" {
  name      = "sagemaker-${var.environment}"
  type      = "orchestrator"
  flavor    = "sagemaker"
  user      = var.user_id
  workspace = var.workspace_id

  configuration = {
    role_arn = aws_iam_role.zenml.arn
  }

  connector = zenml_service_connector.aws.id

  labels = {
    environment = var.environment
  }
}

# Complete Stack
resource "zenml_stack" "aws_stack" {
  name = "aws-${var.environment}"

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