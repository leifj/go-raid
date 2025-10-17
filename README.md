# go-RAiD

A cleanroom implementation of the RAiD (Research Activity Identifier) service in Go, based on the [RAiD OpenAPI specification](https://github.com/au-research/raid-au/blob/main/api-svc/idl-raid-v2/src/raido-openapi-3.0.yaml).

## About RAiD

RAiD (Research Activity Identifier) is a persistent identifier system for research projects and activities. It provides:

- **Persistent Identifiers**: Mint unique, persistent identifiers for research activities
- **Rich Metadata**: Track comprehensive metadata including contributors, organizations, dates, subjects, and related resources
- **Access Control**: Manage access levels (open, embargoed, closed)
- **Versioning**: Full version history and change tracking
- **Service Points**: Multi-tenant support for different organizations

## Project Goals

This is a cleanroom implementation that:

- ✅ Implements the full RAiD OpenAPI 3.0 specification *(in progress, see [API Compliance Analysis](docs/API_COMPATIBILITY_ANALYSIS.md))*
- ✅ Provides a cloud-agnostic solution (no AWS lock-in)
- ✅ Uses standard Go practices and modern libraries
- ✅ Focuses on simplicity and maintainability
- ✅ **Four storage backends**: File, File+Git, CockroachDB, FoundationDB
- ✅ Authentication via JWT with feature flag support (Phase 1 complete)
- 🔄 Full API compatibility with reference implementation (currently ~40%, see [compliance checklist](docs/OPENAPI_COMPLIANCE_CHECKLIST.md))

## Architecture

```
go-RAiD/
├── main.go                      # Application entry point & HTTP server
├── internal/
│   ├── models/                  # Data models based on OpenAPI spec
│   │   └── raid.go             # RAiD, ServicePoint, and related types
│   ├── handlers/                # HTTP request handlers
│   │   ├── raid.go             # RAiD CRUD operations
│   │   ├── servicepoint.go     # Service point management
│   │   └── raid_test.go        # Handler tests (36.3% coverage)
│   ├── storage/                 # Storage abstraction layer
│   │   ├── repository.go       # Storage interface
│   │   ├── file/               # File-based storage (JSON)
│   │   ├── fdb/                # FoundationDB storage
│   │   ├── cockroach/          # CockroachDB storage
│   │   └── testutil/           # Test utilities and mocks
│   ├── config/                  # Configuration management
│   │   └── config.go           # Environment-based configuration
│   └── middleware/              # HTTP middleware
│       └── auth.go             # JWT authentication & authorization
├── docs/                        # Comprehensive documentation
├── raido-openapi-3.0.yaml      # Original OpenAPI specification
└── go.mod                       # Go module definition
```

## Getting Started

### Prerequisites

- **Required**: Go 1.21 or higher
- **Optional**: 
  - CockroachDB or PostgreSQL 14+ (for production SQL storage)
  - FoundationDB (for FDB storage backend)
  - Git (for file-git storage with version control)

### Quick Start with Makefile

This project includes a comprehensive Makefile for easy building and testing. See [`docs/MAKEFILE_GUIDE.md`](docs/MAKEFILE_GUIDE.md) for full documentation.

```bash
# Clone the repository
git clone https://github.com/leifj/go-raid.git
cd go-raid

# Show all available commands
make help

# Install dependencies (minimal - file storage only)
make deps

# Build the application
make build

# Build the application (minimal build - no external dependencies)
make build

# Run tests
make test

# Run in development mode (file-git storage)
make run-dev
```

The API will be available at `http://localhost:8080`

### Manual Installation

If you prefer not to use the Makefile:

```bash
# Install dependencies
go mod download

# Build minimal binary (file storage only, no external dependencies)
go build -tags noexternal -o bin/raid-server .

# Or build with all storage backends (requires lib/pq for CockroachDB)
go build -o bin/raid-server .

# Run the server
./bin/raid-server
```

### Configuration

Configure via environment variables (see [`.env.example`](.env.example) for full list):

```bash
# Server configuration
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Storage backend selection
export STORAGE_TYPE=file              # Options: file, file-git, cockroach, fdb
export STORAGE_FILE_DATADIR=./data    # For file/file-git storage

# CockroachDB configuration (when STORAGE_TYPE=cockroach)
export STORAGE_COCKROACH_HOST=localhost
export STORAGE_COCKROACH_PORT=26257
export STORAGE_COCKROACH_DATABASE=raid
export STORAGE_COCKROACH_USER=root
export STORAGE_COCKROACH_SSLMODE=disable

# Authentication (feature flag - optional)
export AUTH_ENABLED=false              # Set to true to enable JWT auth
export JWT_SECRET=your-secret-key      # Required if AUTH_ENABLED=true
export JWT_ISSUER=https://raid.org
export JWT_AUDIENCE=raid-api

# Handle generation
export HANDLE_PREFIX=10.82481          # Your DOI-like prefix
```

### Storage Backend Options

| Backend | Use Case | Dependencies | Git Integration |
|---------|----------|--------------|-----------------|
| `file` | Development, testing | None | No |
| `file-git` | Development with history | Git (optional) | Yes |
| `cockroach` | Production, distributed | lib/pq | No |
| `fdb` | High-performance | FoundationDB | No |

See [`docs/STORAGE_BACKENDS.md`](docs/STORAGE_BACKENDS.md) for detailed comparison and configuration.

## API Endpoints

### RAiD Operations

- `POST /raid/` - Mint a new RAiD
- `GET /raid/` - List all RAiDs (with filtering: `contributorId`, `contributorRole`, `organisationId`, `organisationRole`)
- `GET /raid/all-public` - List all public RAiDs
- `GET /raid/{prefix}/{suffix}` - Get a specific RAiD
- `PUT /raid/{prefix}/{suffix}` - Update a RAiD
- `PATCH /raid/{prefix}/{suffix}` - Partially update a RAiD (JSON Patch - planned)
- `GET /raid/{prefix}/{suffix}/history` - Get RAiD change history
- `GET /raid/{prefix}/{suffix}/{version}` - Get a specific RAiD version

### Service Point Operations

- `POST /service-point/` - Create a service point
- `GET /service-point/` - List all service points
- `GET /service-point/{id}` - Get a specific service point
- `PUT /service-point/{id}` - Update a service point

### Health Check

- `GET /health` - Service health check

## Development Status

### ✅ Completed (Phase 0 - Foundation)

- [x] Project structure and architecture
- [x] OpenAPI specification integration
- [x] Data models (RAiD, ServicePoint, etc.)
- [x] HTTP server with Chi router
- [x] Route definitions for all endpoints
- [x] Configuration management (environment-based)
- [x] Storage abstraction layer with **four backends** (file, file-git, CockroachDB, FoundationDB)
- [x] RAiD handlers implementation (CRUD operations)
- [x] Service point handlers implementation
- [x] Testing infrastructure with MockRepository
- [x] Handler unit tests (36.3% coverage)
- [x] Build tags for optional dependencies (`-tags noexternal`)
- [x] Comprehensive Makefile for building and testing
- [x] Documentation suite (15+ documents)
- [x] Git repository with contribution guidelines
- [x] JWT authentication middleware with bearer token validation
- [x] Authentication configuration (JWT secret, issuer, audience)
- [x] Feature flag support (`AUTH_ENABLED`) for gradual rollout

### 🔄 In Progress (Phase 1 - API Compliance)

- [ ] Input validation framework (validator tags on models) - **Week 1-2**
- [ ] Request validation middleware
- [ ] Standardized error handling (RFC 7807 Problem Details)
- [ ] Separate request/response types per OpenAPI spec
- [ ] Model field corrections:
  - [ ] `Language.ID` → `Language.Code`
  - [ ] Add `ServicePoint.Password` to create requests only
  - [ ] Fix `Metadata` timestamp format
- [ ] Storage backend unit tests (file, git, FDB, CockroachDB)
- [ ] Integration tests with real storage backends
- [ ] Improve test coverage to 80%+

### 📋 Planned (Phase 2-4)

**Phase 2: High Priority (Weeks 3-4)**
- [ ] Query parameter filtering (contributor/org roles)
- [ ] Field filtering (`includeFields` parameter)
- [ ] Access control enforcement (closed/embargoed RAiDs)
- [ ] Configuration updates (handle generation, registration agency)

**Phase 3: Enhanced Features (Weeks 5-6)**
- [ ] JSON Patch implementation (RFC 6902) for PATCH endpoint
- [ ] RAiD history enhancement (JSON Patch diffs, Base64 encoding)
- [ ] Version-specific retrieval verification
- [ ] Service point completion

**Phase 4: Testing & Documentation (Weeks 7-8)**
- [ ] OpenAPI contract tests (100% schema validation)
- [ ] End-to-end tests with authentication
- [ ] API documentation (Swagger UI)
- [ ] Migration guide from reference implementation
- [ ] Performance benchmarks
- [ ] CI/CD pipeline (GitHub Actions)

**Future**
- [ ] OAuth2/OIDC integration examples
- [ ] Role-based access control (RBAC)
- [ ] Additional storage backends (PostgreSQL, MongoDB)
- [ ] Caching layer (Redis)
- [ ] Search integration (Elasticsearch)
- [ ] Metrics and monitoring (Prometheus)

See [`docs/API_COMPATIBILITY_ANALYSIS.md`](docs/API_COMPATIBILITY_ANALYSIS.md) for detailed roadmap.

## Development

### Using the Makefile

The project includes a comprehensive Makefile. See [`docs/MAKEFILE_GUIDE.md`](docs/MAKEFILE_GUIDE.md) for full documentation.

```bash
# Show all available commands
make help

# Development workflow
make deps              # Install minimal dependencies
make build             # Build (file storage only)
make test              # Run tests
make run-dev           # Run in dev mode

# Quality checks
make fmt               # Format code
make vet               # Run go vet
make lint              # Run golangci-lint (if installed)
make check             # Run all checks

# Testing
make test-coverage     # Run tests with coverage
make coverage-html     # Generate HTML coverage report
make test-race         # Run with race detector

# Full builds
make deps-full         # Install all dependencies
make build-full        # Build with all storage backends
make build-all         # Cross-compile for all platforms

# Docker
make docker-build      # Build Docker image
make compose-up        # Start with Docker Compose
```

### Manual Testing

If not using the Makefile:

```bash
# Run all tests (no external dependencies)
go test -tags noexternal -v ./...

# Run tests with coverage
go test -tags noexternal -cover ./...

# Run with race detector
go test -tags noexternal -race ./...

# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run
```

### Testing API Endpoints

See [`docs/QUICKSTART.md`](docs/QUICKSTART.md) for complete examples.

```bash
# Create a RAiD
curl -X POST http://localhost:8080/raid/ \
  -H "Content-Type: application/json" \
  -d @examples/raid-create.json

# List RAiDs
curl http://localhost:8080/raid/

# Get specific RAiD
curl http://localhost:8080/raid/10.82481/1234567890
```

## Contributing

Contributions are welcome! This is a cleanroom implementation, so:

1. ✅ All code must be written independently based on the OpenAPI spec
2. ✅ Follow Go best practices and idioms (see [`.github/copilot-instructions.md`](.github/copilot-instructions.md))
3. ✅ Add tests for new functionality
4. ✅ Update documentation as needed
5. ✅ Use conventional commits (see [`.gitmessage`](.gitmessage))

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for detailed contribution guidelines.

## Documentation

### Getting Started
- **[Quick Start Guide](docs/QUICKSTART.md)** - Get up and running in 5 minutes
- **[Makefile Guide](docs/MAKEFILE_GUIDE.md)** - Comprehensive guide to the Makefile
- **[Docker Guide](docs/DOCKER_GUIDE.md)** - Container deployment

### Implementation & Design
- **[API Compatibility Analysis](docs/API_COMPATIBILITY_ANALYSIS.md)** - Gap analysis and implementation roadmap (18 sections)
- **[OpenAPI Compliance Checklist](docs/OPENAPI_COMPLIANCE_CHECKLIST.md)** - Implementation tracking
- **[Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md)** - Storage abstraction layer details
- **[Architecture Decisions](docs/architecture-decisions.md)** - Key architectural choices (ADR-001 through ADR-006)

### Features & Configuration
- **[Authentication Guide](docs/AUTHENTICATION.md)** - JWT authentication and security
- **[Storage Backends](docs/STORAGE_BACKENDS.md)** - Comparison and configuration
- **[Testing Plan](docs/TESTING_PLAN.md)** - Comprehensive testing strategy

### Development
- **[Implementation Notes](docs/implementation-notes.md)** - Technical implementation details
- **[Cleanup Summary](docs/CLEANUP_SUMMARY.md)** - Code cleanup and analysis
- **[Git Setup](docs/GIT_SETUP.md)** - Repository configuration

## Project Status

- **API Compliance**: ~40% complete (targeting 100%)
- **Test Coverage**: 36.3% (targeting 80%+)
- **Authentication**: ✅ Phase 1 complete (JWT with feature flag)
- **Storage Backends**: ✅ 4 backends fully functional
- **Documentation**: ✅ 15+ comprehensive guides

**Estimated completion**: 8-10 weeks (with current velocity)

See [`docs/OPENAPI_COMPLIANCE_CHECKLIST.md`](docs/OPENAPI_COMPLIANCE_CHECKLIST.md) for detailed progress tracking.

## License

Apache 2.0 License - see the LICENSE file for details.

## Reference Implementation

This project implements the same OpenAPI specification as the reference implementation at:
<https://github.com/au-research/raid-au>

For more information about RAiD, visit: <https://www.raid.org.au/>

## Contact

For questions or contributions, please open an issue on GitHub.
