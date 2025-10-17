# Storage Abstraction Layer

The go-raid project implements a flexible storage abstraction layer with multiple backend implementations.

## Architecture

### Repository Interface

The `storage.Repository` interface defines all CRUD operations for both RAiD and ServicePoint entities:

```go
type Repository interface {
    // RAiD operations
    CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error)
    GetRAiD(ctx context.Context, prefix, suffix string) (*models.RAiD, error)
    GetRAiDVersion(ctx context.Context, prefix, suffix string, version int) (*models.RAiD, error)
    UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error)
    ListRAiDs(ctx context.Context, filter *RAiDFilter) ([]*models.RAiD, error)
    ListPublicRAiDs(ctx context.Context, filter *RAiDFilter) ([]*models.RAiD, error)
    GetRAiDHistory(ctx context.Context, prefix, suffix string) ([]*models.RAiD, error)
    DeleteRAiD(ctx context.Context, prefix, suffix string) error
    GenerateIdentifier(ctx context.Context, servicePointID int64) (prefix, suffix string, err error)
    
    // ServicePoint operations
    CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error)
    GetServicePoint(ctx context.Context, id int64) (*models.ServicePoint, error)
    UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error)
    ListServicePoints(ctx context.Context) ([]*models.ServicePoint, error)
    DeleteServicePoint(ctx context.Context, id int64) error
    
    // Lifecycle
    Close() error
    HealthCheck(ctx context.Context) error
}
```

## Storage Implementations

### 1. File-based Storage (JSON)

Simple file-based storage using JSON serialization.

**Features:**
- No external dependencies
- Human-readable JSON files
- Organized directory structure by RAiD prefix
- Built-in version history support

**Configuration:**
```bash
export STORAGE_TYPE=file
export STORAGE_FILE_DATADIR=./data
```

**Directory Structure:**
```
data/
├── raids/
│   ├── 10.25.1.1/
│   │   ├── 12345.json        # Current version
│   │   └── .history/
│   │       └── 12345/
│   │           ├── v1.json
│   │           └── v2.json
│   └── 10.25.1.2/
└── servicepoints/
    ├── 1001.json
    └── 1002.json
```

**Use Cases:**
- Development and testing
- Small deployments
- Portable data stores
- Backup/export scenarios

### 2. File-based Storage with Git

Extends file storage with git version control.

**Features:**
- All features of file storage
- Automatic git commits on changes
- Full audit trail via git log
- Easy rollback and diffs
- Supports remote git repositories

**Configuration:**
```bash
export STORAGE_TYPE=file-git
export STORAGE_FILE_DATADIR=./data
export STORAGE_GIT_AUTOCOMMIT=true
export STORAGE_GIT_AUTHOR_NAME="RAiD System"
export STORAGE_GIT_AUTHOR_EMAIL="raid@example.org"
```

**Additional Methods:**
```go
// Get git history for a specific RAiD
commits, err := gitStorage.GetGitLog(prefix, suffix)
```

**Use Cases:**
- Development environments
- Compliance/audit requirements
- Collaboration workflows
- Distributed deployments

### 3. FoundationDB

High-performance, distributed key-value store.

**Features:**
- ACID transactions
- Horizontal scalability
- Multi-datacenter replication
- Strong consistency
- Atomic counters for ID generation

**Configuration:**
```bash
export STORAGE_TYPE=fdb
export STORAGE_FDB_CLUSTER_FILE=/etc/foundationdb/fdb.cluster
export STORAGE_FDB_API_VERSION=710
```

**Data Model:**
- Directory: `/raid/{prefix}/{suffix}/current` → Current RAiD JSON
- Directory: `/raid/{prefix}/{suffix}/version/{n}` → Historical versions
- Directory: `/servicepoint/{id}` → ServicePoint JSON
- Directory: `/counters/raid_{prefix}` → ID counter
- Directory: `/counters/servicepoint_id` → ServicePoint ID counter

**Use Cases:**
- High-throughput applications
- Multi-region deployments
- Mission-critical systems
- Large-scale deployments (millions of RAiDs)

**Dependencies:**
```bash
# Install FoundationDB client
# See: https://apple.github.io/foundationdb/
go get github.com/apple/foundationdb/bindings/go
```

### 4. CockroachDB

Distributed SQL database compatible with PostgreSQL.

**Features:**
- SQL interface
- Horizontal scalability
- Multi-region support
- JSONB for flexible metadata
- Automatic schema migration

**Configuration:**
```bash
export STORAGE_TYPE=cockroach
export STORAGE_COCKROACH_HOST=localhost
export STORAGE_COCKROACH_PORT=26257
export STORAGE_COCKROACH_DATABASE=raid
export STORAGE_COCKROACH_USER=root
export STORAGE_COCKROACH_PASSWORD=
export STORAGE_COCKROACH_SSLMODE=disable

# For production with SSL:
export STORAGE_COCKROACH_SSLMODE=verify-full
export STORAGE_COCKROACH_SSLCERT=/path/to/client.crt
export STORAGE_COCKROACH_SSLKEY=/path/to/client.key
export STORAGE_COCKROACH_SSLROOT=/path/to/ca.crt
```

**Schema:**
```sql
-- RAiD table with versioning
CREATE TABLE raids (
    prefix TEXT NOT NULL,
    suffix TEXT NOT NULL,
    version INT NOT NULL,
    is_current BOOLEAN NOT NULL DEFAULT true,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (prefix, suffix, version)
);

-- Service Point table
CREATE TABLE service_points (
    id SERIAL PRIMARY KEY,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ID counters
CREATE TABLE id_counters (
    name TEXT PRIMARY KEY,
    value INT NOT NULL DEFAULT 1000
);
```

**Use Cases:**
- Cloud-native deployments
- SQL query requirements
- Geographic distribution
- Integration with existing PostgreSQL tools

**Dependencies:**
```bash
go get github.com/lib/pq
```

## Factory Pattern

Storage implementations register themselves using the factory pattern:

```go
// In each storage package's init()
func init() {
    storage.RegisterFactory(storage.StorageTypeFile, func(cfg interface{}) (storage.Repository, error) {
        // Create and return storage instance
    })
}
```

## Configuration

All storage backends are configured via environment variables:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `STORAGE_TYPE` | string | `file` | Storage backend: `file`, `file-git`, `fdb`, `cockroach` |

### File Storage Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_FILE_DATADIR` | `./data` | Data directory path |
| `STORAGE_GIT_AUTOCOMMIT` | `true` | Auto-commit changes |
| `STORAGE_GIT_AUTHOR_NAME` | `RAiD System` | Git author name |
| `STORAGE_GIT_AUTHOR_EMAIL` | `raid@example.org` | Git author email |

### FoundationDB Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_FDB_CLUSTER_FILE` | (empty) | Path to fdb.cluster |
| `STORAGE_FDB_API_VERSION` | `710` | FDB API version |

### CockroachDB Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_COCKROACH_HOST` | `localhost` | Database host |
| `STORAGE_COCKROACH_PORT` | `26257` | Database port |
| `STORAGE_COCKROACH_DATABASE` | `raid` | Database name |
| `STORAGE_COCKROACH_USER` | `root` | Database user |
| `STORAGE_COCKROACH_PASSWORD` | (empty) | Database password |
| `STORAGE_COCKROACH_SSLMODE` | `disable` | SSL mode |
| `STORAGE_COCKROACH_SSLCERT` | (empty) | Client certificate |
| `STORAGE_COCKROACH_SSLKEY` | (empty) | Client key |
| `STORAGE_COCKROACH_SSLROOT` | (empty) | CA certificate |

## Usage Examples

### Development (File Storage)
```bash
export STORAGE_TYPE=file
export STORAGE_FILE_DATADIR=./dev-data
go run main.go
```

### Development with Git History
```bash
export STORAGE_TYPE=file-git
export STORAGE_FILE_DATADIR=./dev-data
go run main.go
```

### Production with CockroachDB
```bash
export STORAGE_TYPE=cockroach
export STORAGE_COCKROACH_HOST=cockroachdb.example.com
export STORAGE_COCKROACH_PORT=26257
export STORAGE_COCKROACH_DATABASE=raid_production
export STORAGE_COCKROACH_USER=raid_app
export STORAGE_COCKROACH_PASSWORD=${COCKROACH_PASSWORD}
export STORAGE_COCKROACH_SSLMODE=verify-full
export STORAGE_COCKROACH_SSLCERT=/etc/certs/client.crt
export STORAGE_COCKROACH_SSLKEY=/etc/certs/client.key
export STORAGE_COCKROACH_SSLROOT=/etc/certs/ca.crt
go run main.go
```

### High-Performance with FoundationDB
```bash
export STORAGE_TYPE=fdb
export STORAGE_FDB_CLUSTER_FILE=/etc/foundationdb/fdb.cluster
go run main.go
```

## Performance Characteristics

| Storage Type | Read Latency | Write Latency | Scalability | Consistency |
|--------------|-------------|---------------|-------------|-------------|
| File | Low | Low | Single node | Strong |
| File+Git | Low | Medium | Single node | Strong |
| FoundationDB | Very Low | Low | Horizontal | Strong (ACID) |
| CockroachDB | Low | Low | Horizontal | Strong (ACID) |

## Migration Between Storage Types

Data can be migrated between storage backends using the common Repository interface:

```go
// Pseudocode for migration
sourceRepo := storage.NewRepository(sourceConfig)
destRepo := storage.NewRepository(destConfig)

raids, _ := sourceRepo.ListRAiDs(ctx, nil)
for _, raid := range raids {
    destRepo.CreateRAiD(ctx, raid)
}

servicePoints, _ := sourceRepo.ListServicePoints(ctx)
for _, sp := range servicePoints {
    destRepo.CreateServicePoint(ctx, sp)
}
```

## Error Handling

All implementations return standardized errors:

- `storage.ErrNotFound` - Resource not found
- `storage.ErrAlreadyExists` - Resource already exists
- `storage.ErrInvalidVersion` - Version mismatch
- `storage.ErrAccessDenied` - Access denied

## Best Practices

1. **Development**: Use file or file-git storage for quick iteration
2. **Testing**: Use file storage for easy inspection and debugging
3. **Production**: Use CockroachDB or FoundationDB for scalability
4. **Compliance**: Use file-git storage for audit trails
5. **High Performance**: Use FoundationDB for low-latency requirements

## Future Enhancements

Potential future storage backends:

- **PostgreSQL**: Traditional PostgreSQL for simpler deployments
- **MongoDB**: Document-oriented storage for flexible schemas
- **S3**: Object storage for archival and backup
- **Redis**: In-memory caching layer
- **Elasticsearch**: Full-text search capabilities
