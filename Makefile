# Makefile
VERSION ?= 0.1.0

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run unit tests
.PHONY: test
test:
	go test ./... -v $(TESTARGS)

# Build provider
.PHONY: build
build:
	go build -o terraform-provider-zenml

# Install provider locally
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/zenml/zenml/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-zenml ~/.terraform.d/plugins/registry.terraform.io/zenml/zenml/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

# Generate docs
.PHONY: docs
docs:
	go generate ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-zenml
