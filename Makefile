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

# Docker parameters
DOCKER_IMAGE=go-raid
DOCKER_TAG=$(VERSION)
DOCKER_IMAGE_FULL=$(DOCKER_IMAGE):$(DOCKER_TAG)
DOCKER_IMAGE_MINIMAL=$(DOCKER_IMAGE):$(DOCKER_TAG)-minimal
DOCKER_IMAGE_LATEST=$(DOCKER_IMAGE):latest
DOCKER_REGISTRY?=
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_FILE=docker-compose.yml

.PHONY: all build build-minimal build-full test test-coverage test-short clean run help deps deps-full fmt vet lint install coverage-html
.PHONY: docker-build docker-build-minimal docker-build-full docker-build-all docker-run docker-run-full docker-run-git docker-stop docker-clean docker-push docker-push-all
.PHONY: compose-up compose-down compose-up-full compose-logs compose-ps compose-restart compose-build

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
	$(GOVET) -tags $(TEST_TAGS) ./...
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

## docker-build: Build minimal Docker image (file storage only)
docker-build: docker-build-minimal

## docker-build-minimal: Build minimal Docker image without external dependencies
docker-build-minimal:
	@echo "Building minimal Docker image (file storage only)..."
	docker build \
		--file Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--tag $(DOCKER_IMAGE_MINIMAL) \
		--tag $(DOCKER_IMAGE):latest-minimal \
		.
	@echo "Minimal Docker image built: $(DOCKER_IMAGE_MINIMAL)"

## docker-build-full: Build full Docker image with all storage backends
docker-build-full:
	@echo "Building full Docker image (all storage backends)..."
	docker build \
		--file Dockerfile.full \
		--build-arg VERSION=$(VERSION) \
		--tag $(DOCKER_IMAGE_FULL) \
		--tag $(DOCKER_IMAGE):latest \
		.
	@echo "Full Docker image built: $(DOCKER_IMAGE_FULL)"

## docker-build-all: Build both minimal and full Docker images
docker-build-all: docker-build-minimal docker-build-full
	@echo "All Docker images built successfully"

## docker-run: Run minimal Docker container
docker-run:
	@echo "Running minimal Docker container..."
	docker run --rm -it \
		-p 8080:8080 \
		-e STORAGE_TYPE=file \
		-v $(PWD)/docker-data:/app/data \
		$(DOCKER_IMAGE):latest-minimal

## docker-run-full: Run full Docker container
docker-run-full:
	@echo "Running full Docker container..."
	docker run --rm -it \
		-p 8080:8080 \
		-e STORAGE_TYPE=file \
		-v $(PWD)/docker-data:/app/data \
		$(DOCKER_IMAGE):latest

## docker-run-git: Run Docker container with git storage
docker-run-git:
	@echo "Running Docker container with git storage..."
	docker run --rm -it \
		-p 8080:8080 \
		-e STORAGE_TYPE=file-git \
		-e GIT_USER_NAME="RAiD Server" \
		-e GIT_USER_EMAIL="raid@example.com" \
		-v $(PWD)/docker-data:/app/data \
		$(DOCKER_IMAGE):latest-minimal

## docker-stop: Stop all running go-raid containers
docker-stop:
	@echo "Stopping all go-raid containers..."
	@docker ps -q --filter ancestor=$(DOCKER_IMAGE) | xargs -r docker stop
	@echo "All containers stopped"

## docker-clean: Remove Docker images and containers
docker-clean:
	@echo "Cleaning Docker images and containers..."
	@docker ps -a -q --filter ancestor=$(DOCKER_IMAGE) | xargs -r docker rm -f
	@docker images $(DOCKER_IMAGE) -q | xargs -r docker rmi -f
	@rm -rf docker-data
	@echo "Docker cleanup complete"

## docker-push: Push Docker images to registry
docker-push:
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set. Use: make docker-push DOCKER_REGISTRY=your-registry.com"; \
		exit 1; \
	fi
	@echo "Pushing Docker images to $(DOCKER_REGISTRY)..."
	docker tag $(DOCKER_IMAGE_MINIMAL) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_MINIMAL)
	docker tag $(DOCKER_IMAGE_FULL) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_FULL)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_MINIMAL)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_FULL)
	@echo "Docker images pushed to registry"

## docker-push-all: Push all Docker image tags to registry
docker-push-all: docker-push
	@echo "Pushing all tags to $(DOCKER_REGISTRY)..."
	docker tag $(DOCKER_IMAGE):latest-minimal $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest-minimal
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest-minimal
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	@echo "All tags pushed to registry"

## compose-up: Start services with Docker Compose (minimal)
compose-up:
	@echo "Starting services with Docker Compose..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d raid-server
	@echo "Services started. Access at http://localhost:8080"

## compose-up-full: Start all services with Docker Compose (full stack)
compose-up-full:
	@echo "Starting full stack with Docker Compose..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) --profile full up -d
	@echo "Full stack started:"
	@echo "  - RAiD (file): http://localhost:8080"
	@echo "  - RAiD (git): http://localhost:8081"
	@echo "  - CockroachDB UI: http://localhost:8082"
	@echo "  - RAiD (cockroach): http://localhost:8083"

## compose-down: Stop all Docker Compose services
compose-down:
	@echo "Stopping Docker Compose services..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) --profile full down
	@echo "Services stopped"

## compose-logs: Show Docker Compose logs
compose-logs:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

## compose-ps: Show Docker Compose service status
compose-ps:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) ps

## compose-restart: Restart Docker Compose services
compose-restart:
	@echo "Restarting Docker Compose services..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) restart
	@echo "Services restarted"

## compose-build: Build Docker Compose images
compose-build:
	@echo "Building Docker Compose images..."
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build
	@echo "Images built"

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
