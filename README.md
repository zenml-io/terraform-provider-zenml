# Terraform Provider for ZenML

[![Tests](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/test.yml/badge.svg)](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/test.yml)
[![Release](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/release.yml/badge.svg)](https://github.com/zenml-io/terraform-provider-zenml/actions/workflows/release.yml)

This Terraform provider allows you to manage ZenML resources using Infrastructure as Code. It provides the ability to manage:

## ZenML OSS Resources
- ZenML Stacks
- Stack Components
- Service Connectors

## ZenML Pro Resources
- Workspaces (via Control Plane API)
- Teams and Team Management
- Projects within Workspaces
- Role Assignments (workspace, project, and stack-level)
- Multi-workspace orchestration

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20
- [ZenML Server](https://docs.zenml.io/) >= 0.70.0
- For Pro features: ZenML Pro subscription with Control Plane access

## Building The Provider

1. Clone the repository
```bash
git clone git@github.com:zenml-io/terraform-provider-zenml.git
```

2. Build the provider
```bash
make build
```

## Installing The Provider

To use the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    zenml = {
      source = "zenml-io/zenml"
    }
  }
}
```

## Using the Provider

### Authentication

The provider supports multiple authentication methods depending on whether you're using ZenML OSS or Pro features:

#### ZenML OSS Authentication (Workspace-only)

For basic ZenML OSS usage, configure with your ZenML server URL and API key:

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}
```

#### ZenML Pro Authentication (Control Plane + Workspace)

For ZenML Pro features that require control plane access, provide both control plane and workspace credentials:

```hcl
provider "zenml" {
  # Control Plane Configuration (for workspaces, teams, projects)
  control_plane_url   = "https://zenml.cloud"
  service_account_key = "your-service-account-key"
  
  # Workspace Configuration (for stacks, components, connectors)
  server_url = "https://your-workspace-url.zenml.io"
  api_key    = "your-workspace-api-key"
}
```

#### Environment Variables

You can also use environment variables:

```bash
# For Control Plane (Pro features)
export ZENML_CONTROL_PLANE_URL="https://zenml.cloud"
export ZENML_SERVICE_ACCOUNT_KEY="your-service-account-key"

# For Workspace (OSS + Pro features)
export ZENML_SERVER_URL="https://your-workspace-url.zenml.io"
export ZENML_API_KEY="your-workspace-api-key"
```

### Authentication Flow

The provider automatically chooses the appropriate authentication method:

1. **Control Plane requests** (workspaces, teams, projects, role assignments) use `service_account_key`
2. **Workspace requests** (stacks, components, connectors) use `api_key` or `api_token`
3. **Mixed scenarios** use both authentication methods as needed

### Getting Your Credentials

#### Service Account Key (Pro)

1. Log in to your ZenML Pro dashboard
2. Navigate to Settings → Service Accounts
3. Create a new service account or use an existing one
4. Copy the service account key

#### Workspace API Key

1. Install ZenML CLI:
```bash
pip install zenml
```

2. Login to your ZenML server:
```bash
zenml login --url <WORKSPACE_URL>
```

3. Create a service account:
```bash
zenml service-account create <SERVICE_ACCOUNT_NAME>
```

### Provider Configuration Options

The provider supports flexible configuration:

```hcl
provider "zenml" {
  # Control Plane (required for Pro features)
  control_plane_url   = "https://zenml.cloud"           # Optional, defaults to https://zenml.cloud
  service_account_key = "your-service-account-key"      # Required for Pro features
  
  # Workspace (required for OSS features, optional for Pro)
  server_url = "https://workspace.zenml.io"             # Required for OSS, optional for Pro
  api_key    = "your-api-key"                           # Alternative: api_token
  api_token  = "your-api-token"                         # Alternative: api_key
}
```

### Example Usage

#### Basic OSS Usage

```hcl
provider "zenml" {
  server_url = "https://your-zenml-server.com"
  api_key    = "your-api-key"
}

# Create a stack with components
resource "zenml_stack_component" "artifact_store" {
  name   = "gcs-store"
  type   = "artifact_store"
  flavor = "gcp"
  
  configuration = {
    path = "gs://my-bucket/artifacts"
  }
}

resource "zenml_stack" "ml_stack" {
  name = "production-stack"
  
  components = {
    artifact_store = zenml_stack_component.artifact_store.id
  }
}
```

#### Pro Usage with Multi-Workspace Management

```hcl
provider "zenml" {
  control_plane_url   = "https://zenml.cloud"
  service_account_key = var.service_account_key
}

# Create a workspace
resource "zenml_workspace" "team_alpha" {
  name        = "team-alpha-workspace"
  description = "Workspace for Team Alpha ML projects"
  
  tags = {
    team        = "alpha"
    environment = "production"
    cost-center = "ml-research"
  }
}

# Create teams
resource "zenml_team" "developers" {
  name        = "alpha-developers"
  description = "Team Alpha developers"
  
  members = [
    "alice@company.com",
    "bob@company.com"
  ]
}

resource "zenml_team" "ml_engineers" {
  name        = "alpha-ml-engineers"
  description = "Team Alpha ML engineers"
  
  members = [
    "charlie@company.com",
    "diana@company.com"
  ]
}

# Create a project within the workspace
resource "zenml_project" "recommendation_engine" {
  workspace_id = zenml_workspace.team_alpha.id
  name         = "recommendation-engine"
  description  = "Customer recommendation ML pipeline"
  
  tags = {
    project-type = "ml-pipeline"
    priority     = "high"
  }
}

# Assign workspace-level roles
resource "zenml_workspace_role_assignment" "dev_team_access" {
  workspace_id = zenml_workspace.team_alpha.id
  team_id      = zenml_team.developers.id
  role         = "Editor"
}

resource "zenml_project_role_assignment" "ml_team_project_access" {
  project_id = zenml_project.recommendation_engine.id
  team_id    = zenml_team.ml_engineers.id
  role       = "Admin"
}
```

See the [examples](./examples/) directory for more complete examples, including the [complete Pro example](./examples/complete-pro/).

## New Pro Resources

### Workspaces
- [Workspace Resource](./docs/resources/workspace.md)
- [Workspace Data Source](./docs/data-sources/workspace.md)

### Teams
- [Team Resource](./docs/resources/team.md)
- [Team Data Source](./docs/data-sources/team.md)

### Projects
- [Project Resource](./docs/resources/project.md)
- [Project Data Source](./docs/data-sources/project.md)

### Role Assignments
- [Workspace Role Assignment Resource](./docs/resources/workspace_role_assignment.md)
- [Project Role Assignment Resource](./docs/resources/project_role_assignment.md)
- [Stack Role Assignment Resource](./docs/resources/stack_role_assignment.md)

## Development

### Requirements

- [Go](https://golang.org/doc/install) >= 1.20
- [GNU Make](https://www.gnu.org/software/make/)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

### Building

1. Clone the repository
```bash
git clone git@github.com:zenml-io/terraform-provider-zenml.git
cd terraform-provider-zenml
```

2. Build the provider
```bash
make build
```

### Testing

Run unit tests:
```bash
make test
```

Run acceptance tests (requires a running ZenML server):
```bash
make testacc
```

### Local Development Installation

To install the provider locally for testing:

```bash
make install
```

This will build and install the provider to your local Terraform plugins directory.

### Documentation

Generate provider documentation:

```bash
make docs
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on contributing to this provider.

## Resource Documentation

### Stacks
- [Stack Resource](./docs/resources/stack.md)
- [Stack Data Source](./docs/data-sources/stack.md)

### Stack Components
- [Stack Component Resource](./docs/resources/stack_component.md)
- [Stack Component Data Source](./docs/data-sources/stack_component.md)

### Service Connectors
- [Service Connector Resource](./docs/resources/service_connector.md)
- [Service Connector Data Source](./docs/data-sources/service_connector.md)

## License

[Apache License 2.0](./LICENSE)