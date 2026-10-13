# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the provider entrypoint. Core implementation lives in `internal/provider/`, with files grouped by concern: `resource_*.go`, `data_source_*.go`, `provider.go`, `client.go`, and shared models/validation helpers. Tests sit beside the code in `*_test.go` files. Generated provider docs live under `docs/resources/` and `docs/data-sources/`, while longer how-to material lives in `docs/guides/`. Example Terraform configurations are in `examples/`; image assets are in `assets/`.

## Build, Test, and Development Commands
Use the `Makefile` for common workflows:

- `make build`: compile `terraform-provider-zenml`
- `make test`: run the Go test suite locally
- `make testacc`: run acceptance tests with `TF_ACC=1` and a live ZenML server
- `make lint`: run `golangci-lint`
- `make fmt`: apply `gofmt` and `goimports` through `golangci-lint`
- `make docs`: regenerate provider documentation after schema changes
- `make install`: copy the built provider into the local Terraform plugin directory

## Coding Style & Naming Conventions
This is a Go module targeting the version declared in `go.mod`. Follow standard Go formatting: tabs for indentation, short receiver names, and `gofmt`/`goimports` output as the source of truth. Keep file names descriptive and aligned with Terraform concepts, for example `resource_stack.go` or `data_source_server.go`. Use exported CamelCase names for framework-facing types and unexported helpers for internal-only logic.

## Testing Guidelines
Use `make test` for fast feedback and `make testacc` for changes that touch API behavior, lifecycle handling, or Terraform state transitions. Acceptance tests follow the `TestAcc...` naming pattern; focused unit tests use regular `Test...` names. Set `ZENML_SERVER_URL` and either `ZENML_API_KEY` or `ZENML_API_TOKEN` before running acceptance coverage. There is no published coverage threshold, so add tests for every new resource/data source path you change.

## Commit & Pull Request Guidelines
Recent history favors short, imperative commit messages such as `Fix updates to include all fields even when empty` or `Add log store to the supported component types`. Keep commits targeted and stage only relevant files. PRs should include a clear description, test coverage, and any required README or `docs/` updates. If provider schemas change, run `make docs` and commit the generated files; CI checks that generated docs are up to date.
