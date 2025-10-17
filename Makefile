# Makefile for go-RAiD project
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary names
BINARY_NAME=raid-server
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_DARWIN=$(BINARY_NAME)_darwin
BINARY_WINDOWS=$(BINARY_NAME).exe

# Build flags
BUILD_FLAGS=-v
BUILD_TAGS_MINIMAL=noexternal
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Test flags
TEST_FLAGS=-v
TEST_COVERAGE_FLAGS=-coverprofile=coverage.out -covermode=atomic
TEST_TAGS=$(BUILD_TAGS_MINIMAL)

# Directories
BUILD_DIR=./bin
COVERAGE_DIR=./coverage
DOCS_DIR=./docs

# Version (can be overridden: make build VERSION=1.0.0)
VERSION?=dev

.PHONY: all build build-minimal build-full test test-coverage test-short clean run help deps deps-full fmt vet lint install docker-build docker-run coverage-html

# Default target
all: test build

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary (minimal - file storage only)
build: build-minimal

## build-minimal: Build binary without external dependencies (file storage only)
build-minimal:
	@echo "Building minimal binary (file storage only)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -tags $(BUILD_TAGS_MINIMAL) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary created at $(BUILD_DIR)/$(BINARY_NAME)"

## build-full: Build binary with all storage backends (requires dependencies)
build-full:
	@echo "Building full binary (all storage backends)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary created at $(BUILD_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags $(BUILD_TAGS_MINIMAL) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) .
	@echo "Linux binary created at $(BUILD_DIR)/$(BINARY_UNIX)"

## build-darwin: Build for macOS
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -tags $(BUILD_TAGS_MINIMAL) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) .
	@echo "macOS binary created at $(BUILD_DIR)/$(BINARY_DARWIN)"

## build-windows: Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -tags $(BUILD_TAGS_MINIMAL) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	@echo "Windows binary created at $(BUILD_DIR)/$(BINARY_WINDOWS)"

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All platform binaries built successfully"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -tags $(TEST_TAGS) $(TEST_FLAGS) ./...

## test-short: Run tests with short flag (skip long-running tests)
test-short:
	@echo "Running short tests..."
	$(GOTEST) -tags $(TEST_TAGS) -short $(TEST_FLAGS) ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -tags $(TEST_TAGS) $(TEST_COVERAGE_FLAGS) ./...
	@echo "Coverage report generated: coverage.out"

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -tags $(TEST_TAGS) -race $(TEST_FLAGS) ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	$(GOTEST) -tags $(TEST_TAGS) -v ./...

## coverage-html: Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOCMD) tool cover -html=coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "HTML coverage report generated at $(COVERAGE_DIR)/coverage.html"
	@echo "Open in browser: file://$(shell pwd)/$(COVERAGE_DIR)/coverage.html"

## fmt: Format all Go code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet check passed"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test
	@echo "All checks passed!"

## clean: Clean build artifacts and test cache
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_DARWIN)
	rm -f $(BINARY_WINDOWS)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out
	rm -rf $(COVERAGE_DIR)
	@echo "Clean complete"

## deps: Download minimal dependencies (for file storage only)
deps:
	@echo "Downloading minimal dependencies..."
	$(GOGET) github.com/go-chi/chi/v5
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Minimal dependencies downloaded"

## deps-full: Download all dependencies (including optional storage backends)
deps-full: deps
	@echo "Downloading optional dependencies..."
	$(GOGET) github.com/lib/pq
	@echo "Note: FoundationDB bindings require manual installation"
	@echo "See: https://github.com/apple/foundationdb/tree/main/bindings/go"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "All dependencies downloaded"

## run: Run the application (file storage mode)
run: build-minimal
	@echo "Starting server..."
	@export STORAGE_TYPE=file && $(BUILD_DIR)/$(BINARY_NAME)

## run-dev: Run with file-git storage in development mode
run-dev: build-minimal
	@echo "Starting server in development mode (file-git storage)..."
	@export STORAGE_TYPE=file-git && \
	export STORAGE_FILE_DATADIR=./dev-data && \
	export SERVER_PORT=8080 && \
	$(BUILD_DIR)/$(BINARY_NAME)

## install: Install binary to GOPATH/bin
install:
	@echo "Installing binary..."
	$(GOBUILD) -tags $(BUILD_TAGS_MINIMAL) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t go-raid:$(VERSION) -t go-raid:latest .
	@echo "Docker image built: go-raid:$(VERSION)"

## docker-run: Run in Docker container
docker-run:
	@echo "Running in Docker container..."
	docker run -p 8080:8080 -e STORAGE_TYPE=file go-raid:latest

## mod-tidy: Tidy and verify dependencies
mod-tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "Modules tidied and verified"

## mod-vendor: Vendor dependencies
mod-vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor
	@echo "Dependencies vendored to ./vendor"

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -tags $(TEST_TAGS) -bench=. -benchmem ./...

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Build time: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"

## docs: Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

## security: Run security checks (requires gosec)
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

## update-deps: Update all dependencies to latest versions
update-deps:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies updated"

## init-dev: Initialize development environment
init-dev: deps
	@echo "Initializing development environment..."
	@mkdir -p dev-data
	@mkdir -p test-data
	@cp -n .env.example .env 2>/dev/null || true
	@echo "Development environment initialized"
	@echo "Edit .env file to configure your environment"

## info: Show project information
info:
	@echo "Project: go-RAiD"
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell $(GOCMD) version)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Binary name: $(BINARY_NAME)"
	@echo "Test tags: $(TEST_TAGS)"
	@echo ""
	@echo "Available storage backends:"
	@echo "  - file: File-based JSON storage (always available)"
	@echo "  - file-git: File storage with Git versioning (always available)"
	@if $(GOCMD) list -m github.com/lib/pq >/dev/null 2>&1; then \
		echo "  - cockroach: CockroachDB storage (installed)"; \
	else \
		echo "  - cockroach: CockroachDB storage (not installed)"; \
	fi
	@echo "  - fdb: FoundationDB storage (check manually)"

## ci: Run all CI checks
ci: deps mod-tidy fmt vet test-race test-coverage
	@echo "All CI checks passed!"

# Aliases for common typos
.PHONY: buidl tset
buidl: build
tset: test
