# examples/complete-aws/main.tf
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
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


resource "aws_iam_user" "iam_user" {
  name = "${var.name_prefix}-zenml-${var.environment}"
}

resource "aws_iam_user_policy" "assume_role_policy" {
  name = "AssumeRole"
  user = aws_iam_user.iam_user[0].name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_access_key" "iam_user_access_key" {
  user = aws_iam_user.iam_user[0].name
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    effect = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type = "AWS"
      identifiers = [aws_iam_user.iam_user[0].arn]
    }
  }
}

resource "aws_iam_role" "stack_access_role" {
  name               = "${var.name_prefix}-zenml-${var.environment}"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

resource "aws_s3_bucket" "artifacts" {
  bucket = "${var.name_prefix}-zenml-artifacts-${var.environment}"
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_ecr_repository" "containers" {
  name = "${var.name_prefix}-zenml-containers-${var.environment}"
}


# ZenML Service Connector for AWS
resource "zenml_service_connector" "aws" {
  name           = "aws-${var.environment}"
  type           = "aws"
  auth_method    = "iam-role"

  configuration = {
    region   = var.region
    role_arn = aws_iam_role.stack_access_role.arn
    aws_access_key_id = aws_iam_access_key.iam_user_access_key[0].id
    aws_secret_access_key = aws_iam_access_key.iam_user_access_key[0].secret
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

  configuration = {
    path = "s3://${aws_s3_bucket.artifacts.bucket}/artifacts"
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = var.environment
  }
}

# Container Registry Component
resource "zenml_stack_component" "container_registry" {
  name      = "ecr-${var.environment}"
  type      = "container_registry"
  flavor    = "aws"

  configuration = {
    uri = aws_ecr_repository.containers.repository_url
  }

  connector_id = zenml_service_connector.aws.id

  labels = {
    environment = var.environment
  }
}

# SageMaker Orchestrator
resource "zenml_stack_component" "orchestrator" {
  name      = "sagemaker-${var.environment}"
  type      = "orchestrator"
  flavor    = "sagemaker"

  configuration = {
    role_arn = aws_iam_role.zenml.arn
  }

  connector_id = zenml_service_connector.aws.id

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