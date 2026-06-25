##############################
# VARIABLES
##############################
# Load .env if exists (loaded first so defaults can override if not set)
-include .env

# Default values (can be overridden by .env or command line)
APP_NAME ?= bbrew
APP_VERSION ?= 0.0.1-local
CONTAINER_IMAGE_NAME ?= bbrew
BUILD_GOVERSION ?= 1.25
BUILD_GOOS ?= $(shell go env GOOS)
BUILD_GOARCH ?= $(shell go env GOARCH)

# Container runtime command
CONTAINER_RUN = podman run --rm -v $(PWD):/app:Z $(CONTAINER_IMAGE_NAME)

##############################
# HELP
##############################
.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help

##############################
# CONTAINER
##############################
.PHONY: container-build-image
container-build-image: ## Build container image
	@podman build -f Containerfile -t $(CONTAINER_IMAGE_NAME) .

.PHONY: container-build-force
container-build-force: ## Force rebuild container image (no cache)
	@podman build --no-cache -f Containerfile -t $(CONTAINER_IMAGE_NAME) .

.PHONY: container-clean
container-clean: ## Remove container image
	@podman rmi $(CONTAINER_IMAGE_NAME) 2>/dev/null || true

##############################
# RELEASE
##############################
.PHONY: release-snapshot
release-snapshot: container-build-image ## Build and release snapshot (testing)
	@$(CONTAINER_RUN) goreleaser release --snapshot --clean

.PHONY: build-snapshot
build-snapshot: container-build-image ## Build snapshot without release
	@$(CONTAINER_RUN) goreleaser build --snapshot --clean

##############################
# BUILD
##############################
.PHONY: build
build: container-build-image ## Build the application binary
	@$(CONTAINER_RUN) env GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) \
		go build -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: build-local
build-local: ## Build locally without container (requires Go installed)
	@go build -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: run
run: build ## Build and run the application
	@./$(APP_NAME)

.PHONY: clean
clean: ## Clean build artifacts
	@rm -f $(APP_NAME)
	@rm -rf dist/

##############################
# QUALITY
##############################
.PHONY: quality
quality: container-build-image ## Run linter checks
	@$(CONTAINER_RUN) golangci-lint run

.PHONY: quality-local
quality-local: ## Run linter locally (requires golangci-lint installed)
	@golangci-lint run

.PHONY: test
test: ## Run tests
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

##############################
# SECURITY
##############################
.PHONY: security
security: security-vuln security-scan ## Run all security checks

.PHONY: security-vuln
security-vuln: container-build-image ## Check for known vulnerabilities
	@$(CONTAINER_RUN) govulncheck ./...

.PHONY: security-vuln-local
security-vuln-local: ## Check vulnerabilities locally (requires govulncheck)
	@govulncheck ./...

.PHONY: security-scan
security-scan: container-build-image ## Run security scanner
	@$(CONTAINER_RUN) gosec ./...

.PHONY: security-scan-local
security-scan-local: ## Run security scanner locally (requires gosec)
	@gosec ./...

##############################
# WEBSITE
##############################
.PHONY: build-site
build-site: ## Build the static website
	@node build.js

.PHONY: serve-site
serve-site: ## Serve the website locally
	@npx http-server docs -p 3000

.PHONY: dev-site
dev-site: build-site serve-site ## Build and serve the website

##############################
# UTILITY
##############################
.PHONY: install
install: build-local ## Install binary to $GOPATH/bin
	@go install ./cmd/$(APP_NAME)

.PHONY: deps
deps: ## Download and tidy dependencies
	@go mod download
	@go mod tidy

.PHONY: deps-update
deps-update: ## Update all dependencies
	@go get -u ./...
	@go mod tidy
