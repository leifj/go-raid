# Docker Guide

This guide covers building and running go-RAiD using Docker and Docker Compose.

## Quick Start

```bash
# Build minimal Docker image (file storage only)
make docker-build

# Run the container
make docker-run

# Access the API at http://localhost:8080
```

## Table of Contents

- [Docker Images](#docker-images)
- [Building Images](#building-images)
- [Running Containers](#running-containers)
- [Docker Compose](#docker-compose)
- [Environment Variables](#environment-variables)
- [Volumes and Data Persistence](#volumes-and-data-persistence)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Docker Images

The project provides two Docker images:

### Minimal Image (Default)

- **Tag**: `go-raid:latest-minimal` or `go-raid:VERSION-minimal`
- **Size**: ~33MB
- **Storage Backends**: File, File+Git
- **Use Case**: Development, testing, simple deployments
- **No external dependencies**: Pure Go build

### Full Image

- **Tag**: `go-raid:latest` or `go-raid:VERSION`
- **Size**: ~45MB
- **Storage Backends**: File, File+Git, CockroachDB
- **Use Case**: Production with multiple storage backend options
- **Note**: FoundationDB requires additional setup

## Building Images

### Using Make (Recommended)

```bash
# Build minimal image (default)
make docker-build
# or explicitly
make docker-build-minimal

# Build full image with all storage backends
make docker-build-full

# Build both images
make docker-build-all

# Build with custom version
make docker-build VERSION=1.0.0
```

### Using Docker Directly

```bash
# Minimal build
docker build -f Dockerfile -t go-raid:minimal .

# Full build
docker build -f Dockerfile.full -t go-raid:full .

# With version argument
docker build -f Dockerfile --build-arg VERSION=1.0.0 -t go-raid:1.0.0 .
```

## Running Containers

### Using Make

```bash
# Run minimal container (file storage)
make docker-run

# Run full container
make docker-run-full

# Run with git storage
make docker-run-git

# Stop all running containers
make docker-stop

# Clean up containers and images
make docker-clean
```

### Using Docker Directly

```bash
# Basic run with file storage
docker run -p 8080:8080 go-raid:latest-minimal

# With persistent volume
docker run -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  go-raid:latest-minimal

# With git storage
docker run -p 8080:8080 \
  -e STORAGE_TYPE=file-git \
  -e GIT_USER_NAME="RAiD Server" \
  -e GIT_USER_EMAIL="raid@example.com" \
  -v $(pwd)/data:/app/data \
  go-raid:latest-minimal

# Interactive mode with shell access
docker run -it --rm \
  -p 8080:8080 \
  go-raid:latest-minimal \
  /bin/sh

# Run in background (detached)
docker run -d --name raid-server \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  go-raid:latest-minimal
```

## Docker Compose

Docker Compose provides orchestration for running multiple services.

### Basic Usage

```bash
# Start minimal service (file storage)
make compose-up
# or
docker-compose up -d raid-server

# Start full stack (all backends + databases)
make compose-up-full
# or
docker-compose --profile full up -d

# View logs
make compose-logs
# or
docker-compose logs -f

# Check status
make compose-ps

# Stop services
make compose-down
```

### Available Services

The `docker-compose.yml` defines several services:

1. **raid-server** (default)
   - File storage
   - Port: 8080
   - Always included

2. **raid-server-git** (profile: full)
   - File+Git storage
   - Port: 8081

3. **cockroachdb** (profile: full, cockroach)
   - CockroachDB database
   - SQL Port: 26257
   - Admin UI: 8082

4. **raid-server-cockroach** (profile: full, cockroach)
   - RAiD with CockroachDB storage
   - Port: 8083

### Using Profiles

```bash
# Full stack with all services
docker-compose --profile full up -d

# Only CockroachDB services
docker-compose --profile cockroach up -d

# Combine profiles
docker-compose --profile full --profile custom up -d
```

### Service URLs

When running with `compose-up-full`:
- File storage: http://localhost:8080
- Git storage: http://localhost:8081
- CockroachDB UI: http://localhost:8082
- CockroachDB storage: http://localhost:8083

## Environment Variables

### Server Configuration

- `SERVER_HOST` - Server bind address (default: `0.0.0.0`)
- `SERVER_PORT` - Server port (default: `8080`)

### Storage Configuration

- `STORAGE_TYPE` - Storage backend type
  - `file` - File-based JSON storage
  - `file-git` - File storage with Git versioning
  - `cockroach` - CockroachDB (full image only)
  - `fdb` - FoundationDB (requires special setup)

### File Storage Options

- `STORAGE_FILE_DATADIR` - Data directory (default: `/app/data`)

### Git Storage Options

- `GIT_USER_NAME` - Git commit author name
- `GIT_USER_EMAIL` - Git commit author email
- `GIT_AUTO_COMMIT` - Auto-commit changes (default: `true`)

### CockroachDB Storage Options

- `STORAGE_COCKROACH_HOST` - Database host
- `STORAGE_COCKROACH_PORT` - Database port (default: `26257`)
- `STORAGE_COCKROACH_DATABASE` - Database name
- `STORAGE_COCKROACH_USER` - Database user
- `STORAGE_COCKROACH_PASSWORD` - Database password
- `STORAGE_COCKROACH_SSLMODE` - SSL mode (default: `require`)

### Example

```bash
docker run -p 8080:8080 \
  -e SERVER_PORT=8080 \
  -e STORAGE_TYPE=file-git \
  -e STORAGE_FILE_DATADIR=/app/data \
  -e GIT_USER_NAME="My RAiD Server" \
  -e GIT_USER_EMAIL="admin@example.com" \
  -v $(pwd)/raid-data:/app/data \
  go-raid:latest-minimal
```

## Volumes and Data Persistence

### Named Volumes (Docker Compose)

Docker Compose creates named volumes:
- `raid-data` - File storage data
- `raid-git-data` - Git storage data
- `cockroach-data` - CockroachDB data

```bash
# List volumes
docker volume ls | grep raid

# Inspect volume
docker volume inspect raid-data

# Backup volume
docker run --rm -v raid-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/raid-backup.tar.gz /data

# Restore volume
docker run --rm -v raid-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/raid-backup.tar.gz -C /
```

### Bind Mounts

```bash
# Mount local directory
docker run -v $(pwd)/local-data:/app/data go-raid:latest-minimal

# Mount with specific permissions
docker run -v $(pwd)/local-data:/app/data:ro go-raid:latest-minimal
```

## Production Deployment

### Best Practices

1. **Use specific version tags**
   ```bash
   docker run go-raid:1.0.0 # Not :latest
   ```

2. **Use health checks**
   ```yaml
   healthcheck:
     test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
     interval: 30s
     timeout: 3s
     retries: 3
   ```

3. **Set resource limits**
   ```bash
   docker run --memory=512m --cpus=1 go-raid:1.0.0
   ```

4. **Use secrets for sensitive data**
   ```bash
   docker run \
     --secret db-password \
     -e STORAGE_COCKROACH_PASSWORD_FILE=/run/secrets/db-password \
     go-raid:1.0.0
   ```

5. **Run as non-root user** (already configured in Dockerfile)

6. **Use multi-stage builds** (already implemented)

### Kubernetes Deployment

Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-raid
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-raid
  template:
    metadata:
      labels:
        app: go-raid
    spec:
      containers:
      - name: go-raid
        image: go-raid:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: STORAGE_TYPE
          value: "cockroach"
        - name: STORAGE_COCKROACH_HOST
          value: "cockroachdb-service"
        volumeMounts:
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: raid-data-pvc
```

### Docker Swarm

```bash
# Create service
docker service create \
  --name raid-server \
  --replicas 3 \
  --publish 8080:8080 \
  --mount type=volume,source=raid-data,target=/app/data \
  go-raid:1.0.0

# Scale service
docker service scale raid-server=5

# Update service
docker service update --image go-raid:1.1.0 raid-server
```

## Registry and Distribution

### Push to Registry

```bash
# Using Make
make docker-push DOCKER_REGISTRY=registry.example.com

# Push all tags
make docker-push-all DOCKER_REGISTRY=registry.example.com

# Using Docker directly
docker tag go-raid:1.0.0 registry.example.com/go-raid:1.0.0
docker push registry.example.com/go-raid:1.0.0
```

### Private Registry

```bash
# Login
docker login registry.example.com

# Build and push
make docker-build VERSION=1.0.0
make docker-push DOCKER_REGISTRY=registry.example.com

# Pull from private registry
docker pull registry.example.com/go-raid:1.0.0
```

### GitHub Container Registry (ghcr.io)

```bash
# Login with GitHub token
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Tag and push
docker tag go-raid:1.0.0 ghcr.io/USERNAME/go-raid:1.0.0
docker push ghcr.io/USERNAME/go-raid:1.0.0
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs <container-id>

# Run with interactive shell
docker run -it --rm go-raid:latest-minimal /bin/sh

# Check health status
docker inspect --format='{{.State.Health.Status}}' <container-id>
```

### Permission Issues

```bash
# Container runs as user 1000:1000
# Fix local directory permissions
sudo chown -R 1000:1000 ./data

# Or run as root (not recommended)
docker run --user root go-raid:latest-minimal
```

### Storage Issues

```bash
# Check volume contents
docker run --rm -v raid-data:/data alpine ls -la /data

# Clear volume data
docker volume rm raid-data

# Check disk space
docker system df
```

### Network Issues

```bash
# Check container networking
docker network inspect raid-network

# Test connectivity
docker run --rm --network raid-network alpine ping raid-server

# Port already in use
# Change host port: -p 8081:8080
```

### Build Issues

```bash
# Clean build cache
docker builder prune

# Build without cache
docker build --no-cache -f Dockerfile .

# Check build context size
du -sh .

# Use .dockerignore to exclude unnecessary files
```

### Image Size Issues

```bash
# Analyze image layers
docker history go-raid:latest-minimal

# Use dive for detailed analysis
dive go-raid:latest-minimal

# Use multi-stage builds (already implemented)
```

## Advanced Topics

### Custom Dockerfile

Create your own Dockerfile based on the provided ones:

```dockerfile
FROM go-raid:latest-minimal

# Add custom configuration
COPY custom-config.yaml /app/config.yaml

# Add custom scripts
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
```

### BuildKit Features

```bash
# Use BuildKit for faster builds
DOCKER_BUILDKIT=1 docker build .

# Use cache mounts
# Add to Dockerfile:
# RUN --mount=type=cache,target=/go/pkg/mod go mod download
```

### Multi-Architecture Builds

```bash
# Build for multiple architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t go-raid:multi-arch .

# Create and use builder
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 --push .
```

## Security Considerations

1. **Image Scanning**
   ```bash
   # Scan for vulnerabilities
   docker scan go-raid:latest-minimal
   ```

2. **Non-root User** - Container runs as user `raid` (UID 1000)

3. **Minimal Base Image** - Uses Alpine Linux (~33MB total)

4. **No Secrets in Image** - Use environment variables or Docker secrets

5. **Read-only Root Filesystem**
   ```bash
   docker run --read-only -v /app/data go-raid:latest-minimal
   ```

## Monitoring and Observability

### Health Checks

```bash
# Manual health check
curl http://localhost:8080/health

# Docker health status
docker ps --filter health=healthy
```

### Logging

```bash
# View logs
docker logs -f raid-server

# Configure logging driver
docker run --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  go-raid:latest-minimal
```

### Metrics

```bash
# Container stats
docker stats raid-server

# Continuous monitoring
docker stats --no-stream
```

## Further Reading

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)

## Getting Help

```bash
# Show all Docker make targets
make help | grep docker

# Show all Compose targets
make help | grep compose

# Docker version
docker --version
docker-compose --version
```

For more information, see:
- [`README.md`](../README.md) - Project overview
- [`docs/MAKEFILE_GUIDE.md`](MAKEFILE_GUIDE.md) - Makefile documentation
- [`docs/QUICK_START.md`](QUICK_START.md) - Quick start guide
- [`docs/STORAGE_BACKENDS.md`](STORAGE_BACKENDS.md) - Storage backend details
