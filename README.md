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

- ✅ Implements the full RAiD OpenAPI 3.0 specification
- ✅ Provides a cloud-agnostic solution (no AWS lock-in)
- ✅ Uses standard Go practices and modern libraries
- ✅ Focuses on simplicity and maintainability
- ✅ Supports PostgreSQL for data persistence
- 🔄 Authentication via JWT (future: OAuth2/OIDC)
- 🔄 Full API compatibility with the reference implementation

## Architecture

```
go-RAiD/
├── main.go                      # Application entry point & HTTP server
├── internal/
│   ├── models/                  # Data models based on OpenAPI spec
│   │   └── raid.go             # RAiD, ServicePoint, and related types
│   ├── handlers/                # HTTP request handlers
│   │   ├── raid.go             # RAiD CRUD operations
│   │   └── servicepoint.go     # Service point management
│   ├── config/                  # Configuration management
│   │   └── config.go           # Environment-based configuration
│   ├── database/                # Database layer (TODO)
│   │   ├── postgres.go         # PostgreSQL connection
│   │   └── repository.go       # Data access layer
│   └── middleware/              # HTTP middleware (TODO)
│       └── auth.go             # Authentication & authorization
├── raido-openapi-3.0.yaml      # Original OpenAPI specification
└── go.mod                       # Go module definition
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Optional: PostgreSQL 14+ or CockroachDB (for production storage backends)
- Optional: FoundationDB (for FDB storage backend)

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

# Build minimal binary (file storage only)
go build -tags noexternal -o bin/raid-server .

# Or build with all storage backends
go build -o bin/raid-server .

# Run the server
./bin/raid-server
```

### Configuration

Configure via environment variables:

```bash
# Server configuration
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=raid
export DB_USER=raid
export DB_PASSWORD=secret
export DB_SSLMODE=disable

# Authentication (optional, for future use)
export AUTH_ENABLED=false
export JWT_SECRET=your-secret-key
```

## API Endpoints

### RAiD Operations

- `POST /raid/` - Mint a new RAiD
- `GET /raid/` - List all RAiDs (with filtering)
- `GET /raid/all-public` - List all public RAiDs
- `GET /raid/{prefix}/{suffix}` - Get a specific RAiD
- `PUT /raid/{prefix}/{suffix}` - Update a RAiD
- `PATCH /raid/{prefix}/{suffix}` - Partially update a RAiD
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

### ✅ Completed
- [x] Project structure and architecture
- [x] OpenAPI specification integration
- [x] Data models (RAiD, ServicePoint, etc.)
- [x] HTTP server with Chi router
- [x] Route definitions for all endpoints
- [x] Configuration management
- [x] Storage abstraction layer with three backends (file, file-git, FoundationDB, CockroachDB)
- [x] RAiD handlers implementation
- [x] Service point handlers implementation
- [x] Testing infrastructure with MockRepository
- [x] Handler unit tests (36.3% coverage)
- [x] Comprehensive Makefile for building and testing
- [x] Documentation suite (storage backends, testing plan, quick start guide)
- [x] Git repository with contribution guidelines

### 🔄 In Progress
- [ ] Storage backend unit tests (file, git, FDB, CockroachDB)
- [ ] Integration tests with real storage backends
- [ ] Change history tracking (JSON Patch RFC 6902)
- [ ] Improve test coverage to 80%+

### 📋 Planned
- [ ] Authentication & authorization (JWT, OAuth2/OIDC)
- [ ] End-to-end tests
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Docker deployment
- [ ] Kubernetes manifests
- [ ] API documentation (Swagger UI)
- [ ] Migration tools from reference implementation
- [ ] Performance benchmarks

## Development

### Using the Makefile

The project includes a comprehensive Makefile. See [`docs/MAKEFILE_GUIDE.md`](docs/MAKEFILE_GUIDE.md) for full documentation.

```bash
# Show all available commands
make help

# Build the project
make build

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Generate HTML coverage report
make coverage-html

# Format, vet, and test
make check

# Run in development mode
make run-dev

# Clean build artifacts
make clean
```

### Manual Testing

If not using the Makefile:

```bash
# Run all tests
go test -tags noexternal -v ./...

# Run tests with coverage
go test -tags noexternal -cover ./...

# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run
```

## Contributing

Contributions are welcome! This is a cleanroom implementation, so:

1. All code must be written independently based on the OpenAPI spec
2. Follow Go best practices and idioms
3. Add tests for new functionality
4. Update documentation as needed

See [`CONTRIBUTING.md`](.github/CONTRIBUTING.md) for detailed contribution guidelines.

## Documentation

- **[Quick Start Guide](docs/QUICK_START.md)** - Get started with go-RAiD
- **[Makefile Guide](docs/MAKEFILE_GUIDE.md)** - Comprehensive guide to the Makefile
- **[Storage Backends](docs/STORAGE_BACKENDS.md)** - Storage backend options and configuration
- **[Testing Plan](docs/TESTING_PLAN.md)** - Testing strategy and guidelines
- **[Implementation Notes](docs/implementation-notes.md)** - Technical implementation details
- **[Architecture Decisions](docs/architecture-decisions.md)** - Key architectural choices

## License

Apache 2.0 License - see the LICENSE file for details.

## Reference Implementation

This project implements the same OpenAPI specification as the reference implementation at:
https://github.com/au-research/raid-au

For more information about RAiD, visit: https://www.raid.org.au/

## Contact

For questions or contributions, please open an issue on GitHub.
