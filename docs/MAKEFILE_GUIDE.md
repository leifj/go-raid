# Makefile Guide

This document provides a comprehensive guide to using the Makefile for the go-RAiD project.

## Quick Start

```bash
# Show all available targets
make help

# Build the project (minimal, file storage only)
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run all checks (format, vet, test)
make check
```

## Build Targets

### Basic Build

- **`make build`** - Build minimal binary (file storage only, no external dependencies)
  - Uses `noexternal` build tag
  - Creates `./bin/raid-server`
  - Default target for development

- **`make build-minimal`** - Alias for `build` (explicit minimal build)

- **`make build-full`** - Build with all storage backends
  - Requires CockroachDB (lib/pq) and FoundationDB dependencies
  - Use for production deployments with multiple storage options

### Cross-Platform Builds

- **`make build-linux`** - Build for Linux (amd64)
- **`make build-darwin`** - Build for macOS (amd64)
- **`make build-windows`** - Build for Windows (amd64)
- **`make build-all`** - Build for all platforms

All binaries are created in `./bin/` directory.

### Versioned Builds

```bash
# Build with version tag
make build VERSION=1.0.0

# Version information is embedded in binary via LDFLAGS
# Check with: ./bin/raid-server --version
```

## Test Targets

### Basic Testing

- **`make test`** - Run all tests
  - Uses `noexternal` tag for tests without external dependencies
  - Verbose output

- **`make test-short`** - Run short tests only
  - Skips long-running tests marked with `testing.Short()`

- **`make test-verbose`** - Run tests with very verbose output

### Coverage Testing

- **`make test-coverage`** - Run tests and generate coverage report
  - Creates `coverage.out` file
  - Uses atomic coverage mode

- **`make coverage-html`** - Generate HTML coverage report
  - Runs tests with coverage
  - Creates `./coverage/coverage.html`
  - Opens in browser for visualization

### Advanced Testing

- **`make test-race`** - Run tests with race detector
  - Detects race conditions
  - Use before merging concurrent code

- **`make benchmark`** - Run benchmark tests
  - Performance testing
  - Shows memory allocations

## Development Targets

### Running the Server

- **`make run`** - Run server with file storage
  - Builds minimal binary first
  - Sets `STORAGE_TYPE=file`

- **`make run-dev`** - Run server in development mode
  - Uses file-git storage
  - Data directory: `./dev-data`
  - Port: 8080

### Environment Setup

- **`make init-dev`** - Initialize development environment
  - Downloads minimal dependencies
  - Creates `dev-data/` and `test-data/` directories
  - Copies `.env.example` to `.env` (if exists)

## Code Quality Targets

### Formatting and Linting

- **`make fmt`** - Format all Go code
  - Runs `go fmt ./...`
  - Auto-fixes formatting issues

- **`make vet`** - Run go vet
  - Static analysis
  - Catches common mistakes

- **`make lint`** - Run golangci-lint (if installed)
  - Comprehensive linting
  - Requires: `golangci-lint` binary

### Combined Checks

- **`make check`** - Run all checks
  - Runs: `fmt`, `vet`, `test`
  - Use before committing code

## Dependency Management

### Installing Dependencies

- **`make deps`** - Download minimal dependencies
  - Only go-chi/chi and core libraries
  - Sufficient for file storage development

- **`make deps-full`** - Download all dependencies
  - Includes lib/pq for CockroachDB
  - Note: FoundationDB requires manual installation

### Module Management

- **`make mod-tidy`** - Tidy and verify modules
  - Runs `go mod tidy` and `go mod verify`
  - Cleans up unused dependencies

- **`make mod-vendor`** - Vendor dependencies
  - Creates `./vendor` directory
  - Use for offline builds

- **`make update-deps`** - Update all dependencies
  - Updates to latest versions
  - Runs `go get -u ./...`

## Docker Targets

- **`make docker-build`** - Build Docker image
  - Tags: `go-raid:dev` and `go-raid:latest`
  - Use VERSION variable for custom tags

- **`make docker-run`** - Run in Docker container
  - Exposes port 8080
  - Uses file storage

```bash
# Build versioned Docker image
make docker-build VERSION=1.2.3
```

## Maintenance Targets

### Cleaning

- **`make clean`** - Clean all build artifacts
  - Removes binaries from `./bin/`
  - Removes coverage files
  - Cleans Go test cache

### Installation

- **`make install`** - Install binary to GOPATH/bin
  - Installs minimal binary
  - Available in PATH as `raid-server`

## CI/CD Targets

### Continuous Integration

- **`make ci`** - Run all CI checks
  - Downloads dependencies
  - Tidies modules
  - Formats code
  - Runs vet
  - Runs tests with race detector
  - Generates coverage report

Use this target in CI pipelines:

```yaml
# Example GitHub Actions
- name: Run CI checks
  run: make ci
```

## Information Targets

- **`make help`** - Show all available targets with descriptions

- **`make info`** - Show project information
  - Version
  - Go version
  - Build configuration
  - Available storage backends

- **`make version`** - Show version information
  - Current version
  - Go version
  - Build timestamp

- **`make docs`** - Start godoc server (if installed)
  - Opens documentation at http://localhost:6060
  - Requires: `godoc` binary

## Security

- **`make security`** - Run security checks (if gosec installed)
  - Static security analysis
  - Vulnerability scanning
  - Requires: `gosec` binary

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
make security
```

## Common Workflows

### Daily Development

```bash
# Start of day
make init-dev         # First time only
make deps            # Ensure dependencies
make fmt             # Format code
make test            # Run tests
make run-dev         # Start server

# During development
make check           # Before each commit
```

### Before Committing

```bash
make check           # Format, vet, test
make test-race       # Check for race conditions
git add .
git commit
```

### Preparing a Release

```bash
# Build for all platforms
make build-all VERSION=1.0.0

# Run full test suite
make test-coverage
make coverage-html   # Review coverage

# Build Docker image
make docker-build VERSION=1.0.0

# Verify all binaries
ls -lh ./bin/
```

### CI/CD Pipeline

```bash
# Single command for CI
make ci

# Or step by step
make deps
make mod-tidy
make fmt
make vet
make test-race
make test-coverage
make build
```

## Customization

### Environment Variables

The Makefile respects these environment variables:

- **`VERSION`** - Build version (default: `dev`)
- **`BUILD_FLAGS`** - Additional build flags
- **`TEST_FLAGS`** - Additional test flags

Example:

```bash
# Custom version build
VERSION=1.2.3 make build

# Custom test flags
TEST_FLAGS="-timeout 30s" make test
```

### Build Tags

- **`noexternal`** - Exclude optional storage backends (default for minimal builds)
- Custom tags can be added to `BUILD_TAGS_MINIMAL` variable

### LDFLAGS

The Makefile automatically injects:
- `main.Version` - Version string
- `main.BuildTime` - UTC build timestamp

Access in code:

```go
var (
    Version   string
    BuildTime string
)

func main() {
    fmt.Printf("Version: %s, Built: %s\n", Version, BuildTime)
}
```

## Troubleshooting

### Build Fails with Missing Dependencies

```bash
# For minimal build (file storage only)
make deps
make build

# For full build (all storage backends)
make deps-full
make build-full
```

### Tests Fail

```bash
# Run with verbose output
make test-verbose

# Run specific test
go test -tags noexternal -v ./internal/handlers -run TestMintRAiD
```

### golangci-lint or gosec Not Found

```bash
# Install linters
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Add to PATH if needed
export PATH=$PATH:$(go env GOPATH)/bin
```

### Cross-Platform Build Fails

Ensure CGO is disabled for pure Go builds:

```bash
CGO_ENABLED=0 make build-all
```

## Tips and Tricks

1. **Tab Completion**: Most shells support tab completion for Makefile targets

2. **Parallel Builds**: Speed up with `-j` flag:
   ```bash
   make -j4 build-all
   ```

3. **Dry Run**: See commands without executing:
   ```bash
   make -n build
   ```

4. **Typo Aliases**: The Makefile includes aliases for common typos:
   - `make buidl` → `make build`
   - `make tset` → `make test`

5. **Combining Targets**:
   ```bash
   make clean build test
   ```

6. **Default Target**: Running `make` without arguments runs `make all` which executes `test` and `build`

## Integration with IDEs

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": "build"
    },
    {
      "label": "Test",
      "type": "shell",
      "command": "make test",
      "group": "test"
    }
  ]
}
```

### GoLand/IntelliJ

1. Run → Edit Configurations
2. Add → Makefile
3. Select target from dropdown

## Further Reading

- [GNU Make Manual](https://www.gnu.org/software/make/manual/)
- [Go Build Tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [Go Testing Documentation](https://pkg.go.dev/testing)

## Getting Help

```bash
# Show all targets
make help

# Show project info
make info

# Show version info
make version
```

For more information, see the other documentation files:
- `README.md` - Project overview
- `docs/QUICK_START.md` - Getting started guide
- `docs/TESTING_PLAN.md` - Testing strategy
- `docs/STORAGE_BACKENDS.md` - Storage backend details
