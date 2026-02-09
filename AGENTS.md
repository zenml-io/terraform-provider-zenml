# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the provider entrypoint.
- `internal/provider/` contains provider logic, resources, data sources, shared models, and tests.
- `docs/resources/` and `docs/data-sources/` contain provider documentation.
- `examples/` contains runnable Terraform examples by scenario (`complete-aws`, `complete-gcp`, `project_creation`, etc.).
- `assets/` stores documentation assets (for example, `assets/workspace_url.png`).
- `tools/tools.go` pins tool dependencies such as `tfplugindocs`.

## Build, Test, and Development Commands
- `make build`: Build the provider binary (`terraform-provider-zenml`).
- `make test`: Run Go tests (`go test ./... -v`).
- `make testacc`: Run acceptance tests with `TF_ACC=1` and extended timeout.
- `make docs`: Regenerate docs via `go generate ./...`.
- `make install VERSION=0.1.0`: Build and install the provider into the local Terraform plugin path.
- Example targeted run: `go test ./internal/provider -run TestAccStack_basic -v`.

## Coding Style & Naming Conventions
- Follow standard Go formatting (`gofmt`) and idiomatic import grouping.
- Keep implementation in package `provider`; prefer focused files per resource/data source.
- Use existing filename patterns: `resource_<name>.go`, `data_source_<name>.go`, and `*_test.go`.
- Use acceptance test naming patterns like `TestAcc<Subject>_<Case>` and helper config functions (`testAcc...Config`).

## Testing Guidelines
- Tests use Go `testing` and `terraform-plugin-testing/helper/resource`.
- Acceptance tests require `ZENML_SERVER_URL` and one auth variable: `ZENML_API_KEY` or `ZENML_API_TOKEN`.
- Add/update acceptance tests for provider behavior changes, including import-state checks where applicable.
- Run `make test` before every PR; run `make testacc` when changing resource or data source behavior.

## Commit & Pull Request Guidelines
- Prefer short, imperative commit subjects consistent with repo history (for example: `Add ...`, `Fix ...`, optional `(#NN)`).
- Branch from `main` with descriptive names like `feature/<topic>` or `fix/<topic>`.
- PRs should include: what changed, why, test evidence, and docs/example updates for user-facing changes.

## Security & Configuration Tips
- Never commit real credentials or tenant-specific URLs.
- Use environment variables for auth (`ZENML_SERVER_URL`, `ZENML_API_KEY`, `ZENML_API_TOKEN`).
- Keep secrets in Terraform `secrets` blocks or external secret managers, not plaintext files.
