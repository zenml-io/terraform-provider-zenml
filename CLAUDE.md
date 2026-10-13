# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraform provider for managing ZenML resources (stacks, stack components, service connectors, projects) via Infrastructure as Code. Built with the **HashiCorp Terraform Plugin Framework** (not the older SDKv2). Communicates with the ZenML server REST API (`/api/v1/...`).

## Build & Development Commands

```bash
make build          # Build the provider binary
make test           # Run unit tests (no server needed)
make testacc        # Run acceptance tests (requires ZENML_SERVER_URL + ZENML_API_KEY)
make lint           # Run golangci-lint (v2, config in .golangci.yml)
make fmt            # Run formatter check
make docs           # Generate provider documentation (go generate)
make install        # Install provider to local Terraform plugins dir
make clean          # Remove build artifacts
```

Run a single test:
```bash
go test ./internal/provider/ -run TestExpandStackComponentsFromTF -v
```

Acceptance tests require env vars:
```bash
export ZENML_SERVER_URL="https://your-server"
export ZENML_API_KEY="your-key"
make testacc
```

## Architecture

All provider code lives in **one package**: `internal/provider/`. There is no separate client package.

### Key Files

- **`main.go`** — Entry point. Wires up `providerserver.Serve` with the provider factory.
- **`provider.go`** — `ZenMLProvider` struct, schema (server_url, api_key, api_token, skip_version_check), and `Configure()` which creates the API client and validates server version (>= 0.80.0).
- **`client.go`** — HTTP client for the ZenML REST API. Handles API key → token exchange via `/api/v1/login`, token expiry/refresh, and all CRUD operations for stacks, components, connectors, and projects.
- **`models.go`** — Go structs mirroring ZenML API request/response shapes. API responses use a nested `Body`/`Metadata` envelope pattern (e.g., `StackResponse.Body.Created`, `StackResponse.Metadata.Components`).
- **`validation.go`** — Valid connector types/auth methods/component types lists, plus `MergeOrCompareConfiguration()` for reconciling Terraform state with server-mutated config.

### Resources (CRUD)

| Resource | File | Notes |
|---|---|---|
| `zenml_stack` | `resource_stack.go` | Components map is `type→ID`. Changing orchestrator/artifact_store triggers replacement via custom plan modifier. |
| `zenml_stack_component` | `resource_stack_component.go` | On delete, proactively removes itself from any stacks (unless it's a required type like orchestrator/artifact_store). |
| `zenml_service_connector` | `resource_service_connector.go` | Supports optional verification with retries and configurable timeouts. `ConnectorType` in API response can be string or object (handled via `json.RawMessage`). |
| `zenml_project` | `resource_project.go` | Simple CRUD for projects. |

### Data Sources (read-only)

| Data Source | File |
|---|---|
| `zenml_server` | `data_source_server.go` |
| `zenml_stack` | `data_source_stack.go` |
| `zenml_stack_component` | `data_source_stack_component.go` |
| `zenml_service_connector` | `data_source_service_connector.go` |

### Important Design Patterns

**Configuration merging**: The ZenML server may mutate component/connector configuration on create/update (e.g., normalizing values, adding defaults). `MergeOrCompareConfiguration()` in `validation.go` ensures user-provided Terraform values are preserved while server-added keys are merged in. This prevents perpetual diffs.

**Required vs optional component types**: `requiredComponentTypes` in `validation.go` defines `orchestrator` and `artifact_store` as mandatory. The custom `requiresReplaceIfRequiredComponentChanges` plan modifier on stacks forces stack replacement when these change, while optional components can be swapped in-place.

**Component deletion lifecycle**: When deleting a `zenml_stack_component`, the provider first checks if it's used by any stacks. For optional component types, it proactively removes the component from those stacks before deletion. For required types, deletion is blocked with an error.

**404 handling**: All `Get*` client methods return `(nil, nil)` on HTTP 404, and all `Delete*` methods silently succeed on 404. Resource `Read` methods remove the resource from state when the API returns nil (drift detection).

## Testing

Two categories of tests:
- **Unit tests** (`resource_stack_helpers_test.go`): Test expand/flatten helpers and plan modifiers directly. No server needed.
- **Acceptance tests** (`resource_stack_test.go`, etc.): Use `resource.Test()` from `terraform-plugin-testing`. Require a running ZenML server. Tests use `testAccPreCheck` to verify env vars and `testAccProviderConfig()` for provider configuration.

Test naming convention: `TestAcc*` for acceptance tests, `Test*` for unit tests.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on push to main and all PRs:
1. **Build** — `go build`
2. **Lint** — `golangci-lint` v2.1.6
3. **Generate** — Ensures `make docs` output is committed (fails if docs are stale)
4. **Test** — `go test -v -count=1 -timeout 5m ./...` (unit tests only, no acceptance)

## Linting

Uses golangci-lint v2 (`.golangci.yml`). Default "standard" linter set with `govet` enable-all. Formatters: `gofmt` + `goimports`.

## Release Process

Automated via GitHub Actions on tag push. GoReleaser sets the `version` variable in `main.go`. Tags follow `vX.Y.Z` format. Provider is published to the Terraform Registry under `zenml-io/zenml`.
