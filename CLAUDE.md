# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make build          # Build provider binary (terraform-provider-zenml)
make test           # Run unit tests (go test ./... -v)
make testacc        # Run acceptance tests (TF_ACC=1, requires running ZenML server, 120m timeout)
make install        # Build and install to local Terraform plugin directory (~/.terraform.d/plugins/...)
make docs           # Regenerate provider documentation (go generate ./...)
make clean          # Remove build artifacts
```

Run a single test:
```bash
go test ./internal/provider -run TestAccStack_basic -v
```

Acceptance tests require environment variables:
```bash
export ZENML_SERVER_URL="https://your-zenml-server.com"
export ZENML_API_KEY="your-api-key"    # or ZENML_API_TOKEN
```

## Architecture

This is a **Terraform Provider** for ZenML, built on the modern **Terraform Plugin Framework** (not the legacy SDKv2). It uses Protocol V6 via `providerserver.Serve()` in `main.go`.

### Core flow

```
main.go → provider.New(version) → ZenMLProvider
  ├── Configure() → creates Client (custom HTTP client for ZenML REST API)
  ├── Resources() → Stack, StackComponent, ServiceConnector, Project
  └── DataSources() → Server, Stack, StackComponent, ServiceConnector
```

All code lives in a single Go package: `internal/provider/`.

### Key files and their roles

- **provider.go** — Provider schema, authentication (API key or token), server version check (>= 0.80.0), passes `Client` to resources/data sources via `resp.ResourceData`/`resp.DataSourceData`
- **client.go** — Custom HTTP client for ZenML API v1. Handles token acquisition (API key → password flow → JWT), automatic token refresh (5-minute buffer before expiry), and all CRUD operations. No external ZenML SDK is used.
- **models.go** — Request/response structs with JSON tags. Separate types for create requests vs read responses. Response bodies nest `Body` (timestamps) and `Metadata` (labels, configs). Uses `Page[T]` for pagination.
- **validation.go** — Central source of truth for valid enum values: connector types, auth methods per connector, component types, resource types. Also contains `MergeOrCompareConfiguration()` which reconciles Terraform state with server responses (important: the ZenML server can mutate configurations on create/update).
- **resource_*.go** — Each implements `resource.Resource` + `resource.ResourceWithImportState` with full CRUD + import
- **data_source_*.go** — Read-only lookups by ID or name

### Configuration reconciliation

A critical design pattern: the ZenML server may add/modify configuration keys when creating or updating components and connectors. `MergeOrCompareConfiguration()` in `validation.go` handles this by merging server responses on top of Terraform state while preserving user-provided keys, and warning about keys the server doesn't recognize.

### File naming conventions

- `resource_<name>.go` / `data_source_<name>.go` / `*_test.go`
- Test functions: `TestAcc<Subject>_<Case>` with helper config functions `testAcc...Config`

## Adding a New Resource

1. Add request/response structs to `models.go`
2. Add CRUD client methods to `client.go`
3. Create `resource_<name>.go` implementing `resource.Resource` with Schema, Create, Read, Update, Delete
4. Register in `provider.go` → `Resources()`
5. Add validation constants to `validation.go` if applicable
6. Add acceptance tests in `resource_<name>_test.go`
7. Add examples in `examples/` and run `make docs`

## Releases

Automated via GoReleaser on git tag push (`v*`). Version is injected via ldflags (`-X main.version={{.Version}}`). Cross-compiles for linux/darwin/windows/freebsd across amd64/386/arm/arm64. GPG-signed checksums.
