# ZenML Pro Terraform Provider Example

This example demonstrates the ZenML Pro Terraform Provider extension capabilities, including:

- **Multi-workspace orchestration** via control plane API
- **Team management** with role-based access control
- **Project management** within workspaces
- **Infrastructure-as-Code** management of ML infrastructure
- **Role assignments** at workspace, project, and stack levels

## Architecture

This example implements the scenario described in the ZenML Pro design document:

> **Scenario**: Platform team receives request: "Team Alpha needs GPU environment for recommendation engine with data lake access. They need a separate ZenML workspace with three sets of roles. Configure stack access based on these roles/teams (prod stack only accessible to specific role/team)."

## Features Demonstrated

### 1. Workspace Management
- Create workspace via control plane API
- Configure workspace metadata and tags
- Manage workspace lifecycle

### 2. Team Management
- Create teams with members
- Consistent team definitions across control plane and workspaces
- Team-based access control

### 3. Project Organization
- Create projects within workspaces
- Configure project metadata and tags
- Team-based project access

### 4. Role-Based Access Control
- Workspace-level role assignments
- Project-level role assignments
- Stack-level permissions with granular access control

### 5. Infrastructure Management
- Development and production stacks
- GPU-enabled production orchestration with SageMaker
- Shared components (artifact store, container registry)
- Team-based infrastructure access

## Prerequisites

1. **ZenML Pro Account**: Active ZenML Pro subscription
2. **Service Account**: Created in ZenML Pro control plane with appropriate permissions
3. **AWS Account**: For infrastructure provisioning
4. **AWS Resources**: Pre-existing S3 bucket, ECR repository, and IAM role

## Usage

### 1. Configure Variables

Create a `terraform.tfvars` file:

```hcl
zenml_control_plane_url   = "https://cloudapi.zenml.io"
zenml_service_account_key = "your-service-account-key"

aws_region                = "us-west-2"
aws_role_arn             = "arn:aws:iam::123456789012:role/zenml-role"
aws_access_key_id        = "AKIA..."
aws_secret_access_key    = "..."

s3_bucket_name       = "your-zenml-artifacts-bucket"
ecr_repository_url   = "123456789012.dkr.ecr.us-west-2.amazonaws.com"

mlflow_username = "admin"
mlflow_password = "password"
```

### 2. Deploy Infrastructure

```bash
terraform init
terraform plan
terraform apply
```

### 3. Verify Deployment

The deployment creates:

- **1 Workspace** (`alpha-workspace`)
- **3 Teams** (developers, ML engineers, ML ops)
- **1 Project** (`recommendation-engine`)
- **2 Stacks** (development and production)
- **5 Stack Components** (orchestrators, artifact store, container registry, model registry)
- **Multiple Role Assignments** with team-based permissions

## Team Access Matrix

| Team | Workspace Role | Project Role | Dev Stack | Prod Stack |
|------|----------------|--------------|-----------|------------|
| Developers | Member | Contributor | Write | Read |
| ML Engineers | Member | Admin | Write | Read |
| ML Ops | Admin | Admin | Admin | Admin |

## Stack Configuration

### Development Stack
- **Orchestrator**: Kubernetes (development cluster)
- **Artifact Store**: S3 (shared)
- **Container Registry**: ECR (shared)
- **Access**: All teams can write

### Production Stack
- **Orchestrator**: SageMaker (GPU-enabled)
- **Artifact Store**: S3 (shared)
- **Container Registry**: ECR (shared)
- **Model Registry**: MLflow
- **Access**: Only ML Ops team has admin access

## Key Benefits

1. **Single `terraform apply`**: Creates complete multi-team ML infrastructure
2. **Team-based Access**: Handles user identity complexity through teams
3. **Environment Separation**: Clear dev/prod boundaries with appropriate access
4. **Scalable Architecture**: Easy to replicate for additional teams
5. **Compliance Ready**: Audit trails and role-based access control

## API Interactions

This configuration interacts with multiple APIs:

- **Control Plane API**: Workspace and team management
- **Workspace API**: Project and stack management within specific workspaces
- **Cross-API Dependencies**: Teams created in control plane are referenced in workspace resources

## Next Steps

After deployment, teams can:

1. **Developers**: Access development stack for experimentation
2. **ML Engineers**: Manage models and development workflows, read production metrics
3. **ML Ops**: Deploy to production, manage infrastructure, monitor all environments

## Cleanup

```bash
terraform destroy
```

**Note**: Ensure all ML workflows are completed before destroying infrastructure to avoid data loss.

## Support

For issues with the ZenML Pro Terraform provider, contact ZenML support or file an issue in the [terraform-provider-zenml](https://github.com/zenml-io/terraform-provider-zenml) repository.