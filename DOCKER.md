# Docker Guide for datagen-cli

This guide covers running datagen-cli in Docker containers.

## Table of Contents

- [Quick Start](#quick-start)
- [Docker Image](#docker-image)
- [Docker Compose](#docker-compose)
- [Usage Examples](#usage-examples)
- [Building Images](#building-images)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Pull from Registry (Future)

```bash
# Pull latest image
docker pull nhaletuc/datagen-cli:latest

# Run with help
docker run --rm nhaletuc/datagen-cli:latest --help

# Generate data from a schema
docker run --rm -v $(pwd):/workspace nhaletuc/datagen-cli:latest \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### Build Locally

```bash
# Build the image
docker build -t datagen-cli:dev .

# Run the container
docker run --rm datagen-cli:dev --help
```

## Docker Image

### Image Details

- **Base**: Alpine Linux 3.19 (~10MB final image)
- **Binary**: Statically compiled Go binary (no dependencies)
- **User**: Non-root user (UID 1000)
- **Entrypoint**: `datagen` command
- **Working Directory**: `/app`

### Image Tags

- `latest`: Latest stable release
- `v1.0.0`: Specific version
- `dev`: Development build

### Image Layers

```
alpine:3.19           (~7MB)
+ ca-certificates     (~500KB)
+ datagen binary      (~6MB)
+ user setup          (minimal)
= Total: ~10-12MB
```

## Docker Compose

The `docker-compose.yml` provides a complete development environment with PostgreSQL.

### Services

- **datagen**: CLI tool for generating data
- **postgres**: PostgreSQL 16 for testing

### Usage

#### 1. Start Services

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps
```

#### 2. Generate Data

```bash
# Generate SQL dump
docker-compose run --rm datagen generate \
  --input /workspace/docs/examples/blog.json \
  --output /workspace/blog.sql \
  --format sql

# Generate with seed for reproducibility
docker-compose run --rm datagen generate \
  --input /workspace/docs/examples/blog.json \
  --output /workspace/blog.sql \
  --format sql \
  --seed 12345
```

#### 3. Import to PostgreSQL

```bash
# Import generated SQL
docker-compose exec postgres psql -U testuser -d testdb -f /workspace/blog.sql

# Verify import
docker-compose exec postgres psql -U testuser -d testdb -c "\dt"
docker-compose exec postgres psql -U testuser -d testdb -c "SELECT COUNT(*) FROM authors;"
```

#### 4. Interactive Development

```bash
# Open shell in datagen container
docker-compose run --rm datagen /bin/sh

# Inside container:
# datagen validate --input /workspace/schema.json
# datagen generate --input /workspace/schema.json --output /workspace/output.sql
# exit
```

#### 5. Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes (deletes PostgreSQL data)
docker-compose down -v
```

## Usage Examples

### Validate Schema

```bash
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  validate --input /workspace/schema.json
```

### Generate SQL Dump

```bash
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  generate \
  --input /workspace/schema.json \
  --output /workspace/output.sql \
  --format sql
```

### Generate COPY Format

```bash
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  generate \
  --input /workspace/schema.json \
  --output /workspace/output.copy.sql \
  --format copy
```

### Use Template

```bash
# List available templates
docker run --rm datagen-cli:dev template list

# Export template
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  template export ecommerce --output /workspace/ecommerce.json

# Generate from template
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  generate \
  --input /workspace/ecommerce.json \
  --output /workspace/ecommerce.sql \
  --format sql
```

### View Examples

```bash
# List example schemas
docker run --rm datagen-cli:dev /bin/sh -c "ls -la /app/examples"

# Copy examples to host
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  /bin/sh -c "cp -r /app/examples/* /workspace/"
```

### Check Version

```bash
docker run --rm datagen-cli:dev version
```

## Building Images

### Local Build

```bash
# Build with version
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S') \
  -t datagen-cli:1.0.0 \
  .

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t datagen-cli:1.0.0 \
  .
```

### Build Script

```bash
# Use the provided build script
./scripts/build.sh --version 1.0.0

# Build Docker image from binary
docker build \
  --build-arg VERSION=1.0.0 \
  -t datagen-cli:1.0.0 \
  .
```

### CI/CD Build

```yaml
# GitHub Actions example
- name: Build Docker Image
  run: |
    docker build \
      --build-arg VERSION=${{ github.ref_name }} \
      --build-arg COMMIT=${{ github.sha }} \
      --build-arg BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S') \
      -t datagen-cli:${{ github.ref_name }} \
      .

- name: Push to Registry
  run: |
    docker tag datagen-cli:${{ github.ref_name }} nhaletuc/datagen-cli:${{ github.ref_name }}
    docker push nhaletuc/datagen-cli:${{ github.ref_name }}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATAGEN_HOME` | `/app` | Application home directory |
| `VERSION` | `dev` | Build version |
| `COMMIT` | `unknown` | Git commit hash |
| `BUILD_DATE` | `unknown` | Build timestamp |

## Volume Mounts

### Input/Output Files

Mount your working directory to access schemas and outputs:

```bash
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### Custom Templates

Mount a custom templates directory:

```bash
docker run --rm \
  -v $(pwd)/my-templates:/app/templates \
  -v $(pwd):/workspace \
  datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### Configuration File

Mount a configuration file:

```bash
docker run --rm \
  -v $(pwd)/.datagen.yaml:/app/.datagen.yaml \
  -v $(pwd):/workspace \
  datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

## Security

### Non-Root User

The container runs as a non-root user (UID 1000):

```bash
# Verify user
docker run --rm datagen-cli:dev /bin/sh -c "whoami"
# Output: datagen

# Check permissions
docker run --rm datagen-cli:dev /bin/sh -c "id"
# Output: uid=1000(datagen) gid=1000(datagen) groups=1000(datagen)
```

### Read-Only Filesystem

Run with read-only filesystem for enhanced security:

```bash
docker run --rm --read-only \
  -v $(pwd):/workspace \
  datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### No Privileged Access

Container requires no privileged access or capabilities.

## Performance

### Resource Limits

Limit container resources:

```bash
docker run --rm \
  --memory=512m \
  --cpus=2 \
  -v $(pwd):/workspace \
  datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### Optimization Tips

1. **Use COPY format** for large datasets (faster than SQL INSERT)
2. **Set seed** for reproducible builds (avoids regeneration)
3. **Limit row counts** during development
4. **Use .dockerignore** to minimize build context

## Troubleshooting

### Permission Errors

If you get permission errors when writing output files:

```bash
# Run with your user ID
docker run --rm --user $(id -u):$(id -g) \
  -v $(pwd):/workspace \
  datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### File Not Found

Ensure files are mounted correctly:

```bash
# Check if file exists in container
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  /bin/sh -c "ls -la /workspace/schema.json"

# Use absolute paths
docker run --rm -v $(pwd):/workspace datagen-cli:dev \
  generate --input /workspace/schema.json --output /workspace/output.sql
```

### Image Build Fails

If build fails due to network issues:

```bash
# Use proxy
docker build \
  --build-arg HTTP_PROXY=http://proxy:port \
  --build-arg HTTPS_PROXY=http://proxy:port \
  -t datagen-cli:dev \
  .

# Use cache from registry
docker build --cache-from datagen-cli:latest -t datagen-cli:dev .
```

### Container Exits Immediately

The default command shows help and exits. For interactive use:

```bash
# Interactive shell
docker run --rm -it datagen-cli:dev /bin/sh

# Keep container running (docker-compose)
docker-compose run --rm datagen /bin/sh
```

## Integration with CI/CD

### GitHub Actions

```yaml
name: Docker Build

on:
  push:
    tags:
      - 'v*'

jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: nhaletuc/datagen-cli:${{ github.ref_name }}
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
            BUILD_DATE=${{ github.event.repository.updated_at }}
```

### GitLab CI

```yaml
build-docker:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG
  only:
    - tags
```

## Publishing to Docker Hub

```bash
# Login
docker login

# Tag image
docker tag datagen-cli:1.0.0 nhaletuc/datagen-cli:1.0.0
docker tag datagen-cli:1.0.0 nhaletuc/datagen-cli:latest

# Push to Docker Hub
docker push nhaletuc/datagen-cli:1.0.0
docker push nhaletuc/datagen-cli:latest
```

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Best Practices](https://docs.docker.com/develop/dev-best-practices/)
