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
- PostgreSQL 14+ (for full implementation)

### Installation

1. Clone this repository:
```bash
git clone https://github.com/leifj/go-raid.git
cd go-raid
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build
```

### Running

Run the development server:
```bash
go run main.go
```

Or build and run:
```bash
go build
./go-RAiD
```

The API will be available at `http://localhost:8080`

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
- [x] Basic handler stubs

### 🔄 In Progress
- [ ] Database integration (PostgreSQL)
- [ ] RAiD identifier generation logic
- [ ] Validation layer
- [ ] Authentication & authorization
- [ ] Service point management
- [ ] Change history tracking (JSON Patch RFC 6902)

### 📋 Planned
- [ ] Unit tests
- [ ] Integration tests
- [ ] Docker deployment
- [ ] Kubernetes manifests
- [ ] API documentation (Swagger UI)
- [ ] Migration tools from reference implementation
- [ ] Performance benchmarks

## Development

### Testing

Run tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

### Code Formatting

Format code:
```bash
go fmt ./...
```

### Linting

Run linter:
```bash
golangci-lint run
```

## Contributing

Contributions are welcome! This is a cleanroom implementation, so:

1. All code must be written independently based on the OpenAPI spec
2. Follow Go best practices and idioms
3. Add tests for new functionality
4. Update documentation as needed

## License

Apache 2.0 License - see the LICENSE file for details.

## Reference Implementation

This project implements the same OpenAPI specification as the reference implementation at:
https://github.com/au-research/raid-au

For more information about RAiD, visit: https://www.raid.org.au/

## Contact

For questions or contributions, please open an issue on GitHub.
