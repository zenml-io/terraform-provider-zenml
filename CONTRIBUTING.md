# Contributing to the ZenML Terraform Provider

Thank you for your interest in contributing to the ZenML Terraform Provider! This document provides guidelines and instructions for contributing.

## Development Requirements

- [Go](https://golang.org/doc/install) >= 1.20
- [GNU Make](https://www.gnu.org/software/make/)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

## Setting Up Development Environment

1. Fork the repository and clone your fork:
```bash
git clone git@github.com:YOUR_USERNAME/terraform-provider-zenml.git
cd terraform-provider-zenml
```

2. Add the upstream remote:
```bash
git remote add upstream git@github.com:zenml-io/terraform-provider-zenml.git
```

3. Install dependencies:
```bash
go mod download
```

## Making Changes

1. Create a new branch for your changes:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and ensure tests pass:
```bash
make test
make testacc
```

3. Update documentation if needed:
```bash
make docs
```

4. Commit your changes:
```bash
git add .
git commit -m "Description of your changes"
```

## Pull Request Process

1. Update your branch with the latest upstream changes:
```bash
git fetch upstream
git rebase upstream/main
```

2. Push your changes:
```bash
git push origin feature/your-feature-name
```

3. Create a pull request through the GitHub UI.

4. Ensure your PR includes:
   - A clear description of the changes
   - Any updates to documentation
   - Tests for new functionality
   - All existing tests passing

## Running Tests

### Unit Tests
```bash
make test
```

### Acceptance Tests
```bash
export ZENML_SERVER_URL="your-test-server"
export ZENML_API_KEY="your-test-key"
make testacc
```

## Documentation

- Update the README.md if you're changing user-facing functionality
- Update or add documentation in the `docs/` directory
- Add examples for new features in the `examples/` directory

## Release Process

Releases are automated through GitHub Actions when a new tag is pushed:

1. Update version in Makefile and provider version
2. Create and push a new tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## Getting Help

- Open an issue for bug reports or feature requests
- Join the ZenML Slack community for discussions
- Check existing documentation and issues before starting work

## Code of Conduct

Please be respectful of others and follow our [Code of Conduct](https://github.com/zenml-io/zenml/blob/main/CODE-OF-CONDUCT.md).