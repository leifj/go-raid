# Quick Start Guide - Storage Abstraction Layer

## Getting Started in 5 Minutes

### 1. Clone and Setup
```bash
cd /home/leifj/work/sunet.se/RAiD/go-raid
go mod download
```

### 2. Choose Your Storage Backend

#### Option A: File Storage (Easiest - No Dependencies)
```bash
export STORAGE_TYPE=file
export STORAGE_FILE_DATADIR=./data
go run main.go
```

#### Option B: File Storage with Git (Development)
```bash
export STORAGE_TYPE=file-git
export STORAGE_FILE_DATADIR=./data
go run main.go
```

#### Option C: CockroachDB (Production)

First, start CockroachDB:
```bash
# Using Docker
docker run -d \
  --name cockroachdb \
  -p 26257:26257 \
  -p 8080:8080 \
  cockroachdb/cockroach:latest \
  start-single-node --insecure
```

Then configure and run:
```bash
export STORAGE_TYPE=cockroach
export STORAGE_COCKROACH_HOST=localhost
export STORAGE_COCKROACH_PORT=26257
export STORAGE_COCKROACH_DATABASE=raid
export STORAGE_COCKROACH_USER=root
export STORAGE_COCKROACH_PASSWORD=
export STORAGE_COCKROACH_SSLMODE=disable
go run main.go
```

#### Option D: FoundationDB (High Performance)

First, install and start FoundationDB:
```bash
# See: https://apple.github.io/foundationdb/getting-started-linux.html
# Install FDB packages for your OS
# Start the service

# Install Go bindings
go get github.com/apple/foundationdb/bindings/go
```

Then configure and run:
```bash
export STORAGE_TYPE=fdb
export STORAGE_FDB_CLUSTER_FILE=/etc/foundationdb/fdb.cluster
go run main.go
```

### 3. Test the API

The server will start on `http://localhost:8080`

#### Create a Service Point
```bash
curl -X POST http://localhost:8080/service-point/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Service Point",
    "identifierOwner": "https://ror.org/12345",
    "prefix": "10.25.1.1",
    "techEmail": "tech@example.org",
    "adminEmail": "admin@example.org",
    "enabled": true
  }'
```

#### Create a RAiD
```bash
curl -X POST http://localhost:8080/raid/ \
  -H "Content-Type: application/json" \
  -d '{
    "title": [{
      "text": "My Research Project",
      "type": {
        "id": "https://vocabulary.raid.org/title.type.schema/5",
        "schemaUri": "https://vocabulary.raid.org/title.type.schema"
      },
      "startDate": "2025-01-01"
    }],
    "date": {
      "startDate": "2025-01-01"
    },
    "access": {
      "type": {
        "id": "https://vocabulary.raid.org/access.type.schema/82",
        "schemaUri": "https://vocabulary.raid.org/access.type.schema"
      }
    }
  }'
```

#### List RAiDs
```bash
curl http://localhost:8080/raid/
```

#### Get Specific RAiD
```bash
# Replace with actual prefix/suffix from previous response
curl http://localhost:8080/raid/10.25.1.1/1234567890
```

#### Update RAiD
```bash
curl -X PUT http://localhost:8080/raid/10.25.1.1/1234567890 \
  -H "Content-Type: application/json" \
  -d '{
    "title": [{
      "text": "My Updated Research Project",
      "type": {
        "id": "https://vocabulary.raid.org/title.type.schema/5",
        "schemaUri": "https://vocabulary.raid.org/title.type.schema"
      },
      "startDate": "2025-01-01"
    }],
    "date": {
      "startDate": "2025-01-01"
    },
    "access": {
      "type": {
        "id": "https://vocabulary.raid.org/access.type.schema/82",
        "schemaUri": "https://vocabulary.raid.org/access.type.schema"
      }
    }
  }'
```

#### Get Version History
```bash
curl http://localhost:8080/raid/10.25.1.1/1234567890/history
```

### 4. Inspect the Data

#### File Storage
```bash
# View the JSON files
tree ./data/
cat ./data/raids/10_25_1_1/1234567890.json
```

#### File Storage with Git
```bash
# View git history
cd ./data
git log --oneline
git show HEAD
```

#### CockroachDB
```bash
# Connect to CockroachDB
docker exec -it cockroachdb ./cockroach sql --insecure

# Query the data
SELECT prefix, suffix, version, is_current, created_at FROM raids;
SELECT * FROM service_points;
```

#### FoundationDB
```bash
# Use fdbcli to inspect data
fdbcli
> getrange \xff/raid/ \xff/raid/\xff
```

## Switching Storage Backends

Just change the `STORAGE_TYPE` environment variable and restart:

```bash
# Stop the server (Ctrl+C)

# Switch to different backend
export STORAGE_TYPE=cockroach  # or: file, file-git, fdb

# Restart
go run main.go
```

## Docker Compose Setup

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  go-raid:
    build: .
    ports:
      - "8080:8080"
    environment:
      - STORAGE_TYPE=cockroach
      - STORAGE_COCKROACH_HOST=cockroachdb
      - STORAGE_COCKROACH_PORT=26257
      - STORAGE_COCKROACH_DATABASE=raid
      - STORAGE_COCKROACH_USER=root
      - STORAGE_COCKROACH_PASSWORD=
      - STORAGE_COCKROACH_SSLMODE=disable
    depends_on:
      - cockroachdb

  cockroachdb:
    image: cockroachdb/cockroach:latest
    command: start-single-node --insecure
    ports:
      - "26257:26257"
      - "8081:8080"
    volumes:
      - cockroach-data:/cockroach/cockroach-data

volumes:
  cockroach-data:
```

Run with:
```bash
docker-compose up
```

## Troubleshooting

### File Storage Issues
```bash
# Check permissions
ls -la ./data/

# Clear data
rm -rf ./data/
```

### CockroachDB Issues
```bash
# Check if running
docker ps | grep cockroach

# Check logs
docker logs cockroachdb

# Connect and check schema
docker exec -it cockroachdb ./cockroach sql --insecure
> \d
> SELECT * FROM raids LIMIT 5;
```

### FoundationDB Issues
```bash
# Check FDB status
fdbcli
> status

# Check cluster file
cat /etc/foundationdb/fdb.cluster

# Restart FDB
sudo systemctl restart foundationdb
```

### Git Storage Issues
```bash
# Check if git is installed
git --version

# Manual git operations in data directory
cd ./data
git status
git log
```

## Next Steps

1. **Read the full documentation**: `docs/storage-backends.md`
2. **Check the implementation summary**: `docs/IMPLEMENTATION_SUMMARY.md`
3. **Review the example config**: `.env.example`
4. **Explore the API**: Try all the endpoints
5. **Switch backends**: Try different storage options
6. **Add authentication**: Enable `AUTH_ENABLED=true`
7. **Deploy to production**: Use CockroachDB or FoundationDB

## Common Use Cases

### Development
```bash
export STORAGE_TYPE=file-git
export STORAGE_FILE_DATADIR=./dev-data
```

### Testing
```bash
export STORAGE_TYPE=file
export STORAGE_FILE_DATADIR=./test-data
```

### Staging
```bash
export STORAGE_TYPE=cockroach
export STORAGE_COCKROACH_HOST=staging-db.example.com
```

### Production
```bash
export STORAGE_TYPE=cockroach
export STORAGE_COCKROACH_HOST=production-cluster.example.com
export STORAGE_COCKROACH_SSLMODE=verify-full
export STORAGE_COCKROACH_SSLCERT=/etc/certs/client.crt
export STORAGE_COCKROACH_SSLKEY=/etc/certs/client.key
export STORAGE_COCKROACH_SSLROOT=/etc/certs/ca.crt
```

## Questions?

- Read the docs: `docs/`
- Check examples: `.env.example`
- Review code: `internal/storage/`
- Open an issue on GitHub
