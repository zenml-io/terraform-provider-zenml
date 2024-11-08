terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
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

provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Create S3 bucket for ZenML artifacts
resource "aws_s3_bucket" "artifacts" {
  bucket = "${data.aws_caller_identity.current.account_id}-zenml-artifacts-${var.environment}"
}

# Create ECR repository for ZenML containers

resource "aws_ecr_repository" "containers" {
  name = "zenml-containers-${var.environment}"
}

# Create IAM user and role with required permissions and keys

resource "aws_iam_user" "iam_user" {
  name = "zenml-${var.environment}"
}

resource "aws_iam_user_policy" "assume_role_policy" {
  name = "AssumeRole"
  user = aws_iam_user.iam_user.name

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
  user = aws_iam_user.iam_user.name
}

resource "aws_iam_role" "zenml" {
  name               = "zenml-${var.environment}"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = aws_iam_user.iam_user.arn
        }
        Action = "sts:AssumeRole"
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "sagemaker.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}


resource "aws_iam_role_policy" "s3_policy" {
  name = "S3Policy"
  role = aws_iam_role.zenml.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:GetBucketVersioning"
        ]
        Resource = [
          aws_s3_bucket.artifacts.arn,
          "${aws_s3_bucket.artifacts.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "ecr_policy" {
  name = "ECRPolicy"
  role = aws_iam_role.zenml.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:DescribeRegistry",
          "ecr:BatchGetImage",
          "ecr:DescribeImages",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage"
        ]
        Resource = aws_ecr_repository.containers.arn
      },
      {
        Effect = "Allow"
        Action = "ecr:GetAuthorizationToken"
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:DescribeRepositories",
          "ecr:ListRepositories"
        ]
        Resource = "arn:aws:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/*"
      }
    ]
  })
}

resource aws_iam_role_policy_attachment "sagemaker_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSageMakerFullAccess"
  role = aws_iam_role.zenml.name
}

# ZenML Service Connector for AWS
resource "zenml_service_connector" "aws" {
  name           = "aws-${var.environment}"
  type           = "aws"
  auth_method    = "iam-role"

  configuration = {
    region   = var.region
    role_arn = aws_iam_role.zenml.arn
    aws_access_key_id = aws_iam_access_key.iam_user_access_key.id
    aws_secret_access_key = aws_iam_access_key.iam_user_access_key.secret
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
    uri = regex("^([^/]+)/?", aws_ecr_repository.containers.repository_url)[0]
    default_repository = "${aws_ecr_repository.containers.name}"
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
      region = data.aws_region.current.name
      execution_role = aws_iam_role.zenml.arn
      output_data_s3_uri = "s3://${aws_s3_bucket.artifacts.bucket}/sagemaker"
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