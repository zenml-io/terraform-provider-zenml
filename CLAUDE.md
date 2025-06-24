# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Development
- `make build` - Build the Terraform provider binary
- `make install` - Build and install the provider locally for testing
- `make clean` - Remove build artifacts

### Testing
- `make test` - Run unit tests
- `make testacc` - Run acceptance tests (requires running ZenML server)
- `go test ./... -v` - Run all tests with verbose output
- `TF_ACC=1 go test ./internal/provider -v -run TestAccStack` - Run specific acceptance test

### Documentation
- `make docs` - Generate provider documentation using `go generate`

## Architecture

This is a Terraform provider for ZenML that enables Infrastructure as Code management of ZenML resources.

### Core Components

**Provider Entry Point (`main.go`)**
- Serves the Terraform plugin using HashiCorp's plugin framework
- Delegates to the internal provider package

**Provider Core (`internal/provider/provider.go`)**
- Defines the provider schema with authentication options (server_url, api_key, api_token)
- Configures resources and data sources
- Handles provider-level configuration and ZenML server version compatibility checks

**HTTP Client (`internal/provider/client.go`)**
- Implements ZenML REST API client with OAuth2-style authentication
- Handles API token refresh using API keys
- Provides CRUD operations for all ZenML resources
- Includes comprehensive logging and error handling

**Data Models (`internal/provider/models.go`)**
- Defines Go structs for ZenML API request/response objects
- Uses generic Page[T] type for paginated responses
- Covers Stacks, Stack Components, Service Connectors, Users, and Projects

### Resources and Data Sources

The provider manages three main ZenML resource types:

1. **Stacks** (`resource_stack.go`, `data_source_stack.go`)
   - Logical groupings of stack components
   - Components are referenced by UUID

2. **Stack Components** (`resource_stack_component.go`, `data_source_stack_component.go`)
   - Individual infrastructure components (artifact stores, orchestrators, etc.)
   - Can be linked to service connectors for authentication

3. **Service Connectors** (`resource_service_connector.go`, `data_source_service_connector.go`)
   - Authentication connectors for cloud providers (AWS, GCP, Azure)
   - Handle credentials and resource access

### Authentication Flow

The provider supports two authentication methods:
- **API Keys**: Long-term credentials for service accounts (recommended for Terraform)
- **API Tokens**: Short-lived tokens that can be refreshed using API keys

Authentication is handled in `client.go:getAPIToken()` which automatically refreshes expired tokens using the OAuth2 password flow.

### Testing Strategy

- **Unit Tests**: Test individual functions and logic (`*_test.go` files)
- **Acceptance Tests**: Test full provider functionality against a real ZenML server
- Tests require `TF_ACC=1` environment variable for acceptance tests
- Server URL and credentials must be configured via environment variables for acceptance tests

### Environment Variables

- `ZENML_SERVER_URL` - ZenML server endpoint
- `ZENML_API_KEY` - Service account API key (recommended)
- `ZENML_API_TOKEN` - Short-lived API token
- `TF_ACC=1` - Enable acceptance tests

### Version Compatibility

The provider enforces ZenML server version compatibility (>= 0.80.0) but this can be bypassed with `skip_version_check = true` in the provider configuration.