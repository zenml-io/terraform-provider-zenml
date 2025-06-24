# ZenML Terraform Provider Extension Plan

## Overview
Extend the existing Terraform provider to support ZenML Pro Control Plane API resources (users, teams, workspaces) in addition to existing workspace API resources (stacks, components, service_connectors). Also add support for projects in the workspace API.

## Current Architecture Analysis

### Existing Files Structure
- `internal/provider/client.go` - Single client for workspace API only
- `internal/provider/provider.go` - Provider configuration and resource registration
- `internal/provider/models.go` - Data models for existing resources
- Resources: `resource_stack.go`, `resource_stack_component.go`, `resource_service_connector.go`
- Data sources: `data_source_*.go` files

### Current Authentication
- Single authentication method via API key/token
- Login endpoint: `/api/v1/login` (workspace API)
- Bearer token authentication in `client.go:134`

## Key Architectural Decisions

### 1. Client Architecture Decision
**Decision Required**: How to handle dual API support?

**Options**:
- **Option A**: Extend existing `Client` struct with control plane methods
- **Option B**: Create separate `ControlPlaneClient` struct
- **Option C**: Abstract client interface with workspace/control plane implementations

**Current Implementation**: Single `Client` struct in `client.go:22-28`

**Recommendation**: Option B - separate clients for clear separation of concerns

### 2. Authentication Strategy Decision
**Decision Required**: How to handle different authentication methods?

**Current**: Workspace API uses API key → access token flow (`client.go:75-109`)
**Control Plane API**: Uses OAuth2 device flow (per OpenAPI spec)

**Options**:
- **Option A**: Add OAuth2 methods to existing client
- **Option B**: Separate authentication per client type
- **Option C**: Unified authentication abstraction

**Files to modify**: `client.go`, `provider.go:54-135`

### 3. Provider Configuration Decision
**Decision Required**: How to configure dual API endpoints?

**Current Schema** (`provider.go:14-37`):
- `server_url` - workspace API URL
- `api_key`/`api_token` - workspace authentication

**New Requirements**:
- Control plane URL (likely `cloudapi.zenml.io`)
- Control plane authentication credentials
- Backward compatibility

**Options**:
- **Option A**: Add `control_plane_url` and `control_plane_*` auth fields
- **Option B**: Auto-detect control plane URL from workspace URL
- **Option C**: Separate provider blocks

### 4. Resource Naming Convention Decision
**Decision Required**: How to name new resources?

**Current Resources**:
- `zenml_stack` (workspace API)
- `zenml_stack_component` (workspace API)  
- `zenml_service_connector` (workspace API)

**New Resources Needed**:
- `zenml_project` (workspace API - add to existing pattern)
- `zenml_user` (control plane API)
- `zenml_team` (control plane API)
- `zenml_workspace` (control plane API)

**Options**:
- **Option A**: Flat naming (current approach)
- **Option B**: Prefix with API type (`zenml_cp_user`, `zenml_ws_project`)
- **Option C**: Separate provider namespaces

### 5. Resource Relationship Decision
**Decision Required**: How to handle cross-API resource relationships?

**ZenML Hierarchy** (from docs):
```
Organization → Workspace → Project → Stack/Component/ServiceConnector
```

**Terraform Challenges**:
- Control plane manages workspaces
- Workspace API manages projects
- Cross-API resource dependencies

**Files to consider**: All `resource_*.go` files for dependency modeling

### 6. Error Handling Decision
**Decision Required**: How to handle API-specific errors?

**Current**: Generic error handling in `client.go:167-169`
**New**: Different error formats between APIs

### 7. Testing Strategy Decision
**Decision Required**: How to test dual API functionality?

**Current Tests**:
- `resource_*_test.go` files
- Mock server approach vs real API testing

**New Considerations**:
- Control plane requires OAuth2 flow
- Cross-API resource testing
- Authentication testing

## Implementation Strategy

### Phase 1: Foundation
1. **Decision**: Finalize client architecture approach
2. **Files**: Extend or refactor `client.go`
3. **Tasks**: 
   - Implement control plane authentication
   - Create control plane client methods
   - Update provider configuration schema

### Phase 2: Workspace API Extension
1. **Resource**: Add `zenml_project` 
2. **Files**: Create `resource_project.go`, `data_source_project.go`
3. **Models**: Add project models to `models.go`
4. **Client**: Add project methods to workspace client

### Phase 3: Control Plane Resources
1. **Resources**: Add `zenml_user`, `zenml_team`, `zenml_workspace`
2. **Files**: Create corresponding `resource_*.go` and `data_source_*.go`
3. **Models**: Add control plane models
4. **Client**: Implement control plane API methods

### Phase 4: Integration & Testing
1. **Cross-API**: Handle resource relationships
2. **Tests**: Update all `*_test.go` files
3. **Docs**: Update documentation and examples
4. **Validation**: Add resource validation logic

## Files That Will Need Changes

### Core Files
- `internal/provider/client.go` - Add control plane client
- `internal/provider/provider.go` - Update schema and registration
- `internal/provider/models.go` - Add new resource models

### New Resource Files
- `internal/provider/resource_project.go`
- `internal/provider/resource_user.go` 
- `internal/provider/resource_team.go`
- `internal/provider/resource_workspace.go`

### New Data Source Files
- `internal/provider/data_source_project.go`
- `internal/provider/data_source_user.go`
- `internal/provider/data_source_team.go` 
- `internal/provider/data_source_workspace.go`

### Test Files
- All existing `*_test.go` files may need updates
- New test files for new resources

### Documentation
- `docs/resources/` - Add new resource documentation
- `docs/data-sources/` - Add new data source documentation
- `examples/` - Add example configurations

## Open Questions for Decision

1. Should the provider auto-detect control plane URL or require explicit configuration?
2. How should cross-API resource dependencies be modeled in Terraform?
3. What's the backward compatibility strategy for existing configurations?
4. Should there be separate authentication flows or unified approach?
5. How to handle different API rate limits and error responses?
6. Should resources be prefixed to indicate which API they use?
7. How to handle workspace deployment states in Terraform lifecycle?
8. What's the testing strategy for OAuth2 flows in CI/CD?

## Next Steps

1. Make architectural decisions listed above
2. Create detailed implementation plan based on decisions  
3. Begin implementation starting with Phase 1
4. Iterate and test each phase before proceeding