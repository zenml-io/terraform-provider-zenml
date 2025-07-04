# Complete ZenML Pro Example

**⚠️ IMPORTANT NOTICE**: This example is **CONCEPTUAL ONLY** and does not work with the real ZenML Cloud API. The real API uses OAuth2 authentication and has different endpoints and data structures. See [../../AUTHENTICATION_ANALYSIS.md](../../AUTHENTICATION_ANALYSIS.md) for details.

This example demonstrates the complete Team Alpha scenario from the ZenML Pro Terraform Provider design document, showcasing how platform teams can provision an entire ML infrastructure setup with a single `terraform apply`.

## What This Example Creates

This configuration creates a complete multi-team ML environment:

- **1 Workspace**: `team-alpha-workspace` for Team Alpha
- **3 Teams**: Developers, ML Engineers, and ML Ops teams
- **1 Project**: `recommendation-engine` for the customer recommendation pipeline
- **2 Stacks**: Development and production environments
- **Role-based Access Control**: Proper permissions for each team

## Prerequisites

**⚠️ IMPORTANT**: These prerequisites are for the conceptual implementation only and do not reflect the real ZenML Cloud API requirements.

1. **ZenML Pro Subscription**: Access to ZenML Pro with control plane features
2. **Service Account Key**: Generated from your ZenML Pro dashboard
3. **Terraform**: Version 1.0 or later
4. **Provider Configuration**: The ZenML Terraform provider installed

## Quick Start

**⚠️ IMPORTANT**: The following steps are conceptual and will not work against the real ZenML Cloud API.

1. **Set up your environment**:
```bash
export ZENML_CONTROL_PLANE_URL="https://zenml.cloud"
export ZENML_SERVICE_ACCOUNT_KEY="your-service-account-key"
```

2. **Initialize Terraform**:
```bash
terraform init
```

3. **Review the plan**:
```bash
terraform plan -var="service_account_key=your-service-account-key"
```

4. **Apply the configuration**:
```bash
terraform apply -var="service_account_key=your-service-account-key"
```

## Configuration Details

### Team Alpha Access Matrix

**⚠️ Note**: This access matrix is conceptual and doesn't reflect the real ZenML Cloud API role structure.

| Team | Workspace Access | Project Access | Dev Stack | Prod Stack |
|------|------------------|----------------|-----------|------------|
| Developers | Editor | Editor | Admin | Read-only |
| ML Engineers | Editor | Admin | Admin | Editor |
| ML Ops | Admin | Admin | Admin | Admin |

### Infrastructure Components

**⚠️ Note**: These components are conceptual and may not match the real ZenML Cloud API structure.

- **Development Stack**: 
  - Local artifact store
  - Local orchestrator
  - Basic compute resources
  
- **Production Stack**:
  - S3 artifact store
  - SageMaker orchestrator
  - GPU-enabled compute nodes
  - Production-grade security

## Outputs

After successful deployment, you'll get:

```bash
# Workspace Information
workspace_id = "uuid-of-team-alpha-workspace"
workspace_url = "https://team-alpha-workspace.zenml.io"

# Team Information
team_ids = {
  developers = "uuid-of-developers-team"
  ml_engineers = "uuid-of-ml-engineers-team"
  ml_ops = "uuid-of-ml-ops-team"
}

# Project Information
project_id = "uuid-of-recommendation-engine-project"

# Stack Information
stack_ids = {
  development = "uuid-of-dev-stack"
  production = "uuid-of-prod-stack"
}
```

## Customization

**⚠️ Note**: These customization options are conceptual only.

### Adding More Teams

```hcl
resource "zenml_team" "data_scientists" {
  name        = "alpha-data-scientists"
  description = "Team Alpha data scientists"
  
  members = [
    "scientist1@company.com",
    "scientist2@company.com"
  ]
}
```

### Adding More Projects

```hcl
resource "zenml_project" "fraud_detection" {
  workspace_id = zenml_workspace.team_alpha.id
  name         = "fraud-detection"
  description  = "Fraud detection ML pipeline"
  
  tags = {
    project-type = "ml-pipeline"
    priority     = "medium"
  }
}
```

### Customizing Stack Configuration

```hcl
resource "zenml_stack" "staging" {
  name = "staging-stack"
  
  components = {
    artifact_store = zenml_stack_component.s3_staging.id
    orchestrator   = zenml_stack_component.sagemaker_staging.id
    container_registry = zenml_stack_component.ecr_staging.id
  }
  
  labels = {
    environment = "staging"
    team        = "alpha"
    cost-center = "ml-research"
  }
}
```

## Cost Considerations

**⚠️ Note**: These cost considerations are conceptual and may not reflect actual ZenML Pro pricing.

This configuration provisions resources that may incur costs:

- **ZenML Pro Workspace**: Based on your subscription plan
- **Team Seats**: 6 team members across 3 teams
- **AWS Resources**: S3 buckets, SageMaker endpoints, ECR registries
- **Compute Resources**: GPU-enabled instances for production workloads

## Security Best Practices

**⚠️ Note**: These security practices are conceptual and may not apply to the real ZenML Cloud API.

1. **Least Privilege**: Teams only get the minimum required access
2. **Environment Separation**: Development and production stacks are isolated
3. **Secret Management**: Use encrypted secrets for sensitive configurations
4. **Access Auditing**: All role assignments are tracked and logged

## Troubleshooting

**⚠️ Note**: These troubleshooting steps are conceptual only.

### Common Issues

1. **Authentication Errors**: Verify your service account key is correct
2. **Team Member Issues**: Ensure email addresses are valid ZenML Pro users
3. **Resource Conflicts**: Check for naming conflicts with existing resources
4. **Permission Denied**: Verify your service account has organization admin rights

### Verification Steps

```bash
# Check workspace status
terraform show | grep workspace_id

# Verify team creation
terraform show | grep team_ids

# Check role assignments
terraform show | grep role_assignment
```

## Cleanup

To remove all resources:

```bash
terraform destroy -var="service_account_key=your-service-account-key"
```

**⚠️ Important**: This will delete the workspace, teams, projects, and all associated resources. Make sure to backup any important data first.

## Support

For issues with this example:

1. Check the [Authentication Analysis](../../AUTHENTICATION_ANALYSIS.md) for known limitations
2. Review the [main README](../../README.md) for current status
3. File issues in the GitHub repository

Remember that this is a conceptual example that demonstrates the intended architecture but does not work with the real ZenML Cloud API.