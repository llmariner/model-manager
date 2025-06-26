# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running

Build the server component:
```bash
make build-server
./bin/server run --config configs/server/config.yaml
```

Build the loader component:
```bash
make build-loader
./bin/loader run --config configs/loader/config.yaml
```

Build Docker images:
```bash
make build-docker-server
make build-docker-loader
```

### Testing and Linting

Run all tests:
```bash
make test
# Or specifically:
go test ./...
go test -race ./...
```

Run full linting suite:
```bash
make lint
```

Individual linting commands:
```bash
go fmt ./...
go vet ./...
go mod tidy
errcheck ./...
golint -set_exit_status ./...
```

### Protocol Buffer Generation

Generate protobuf and TypeScript files:
```bash
make generate
```

This runs:
- `buf generate` for Go protobuf files
- `buf generate --template buf.gen.ts.yaml` for TypeScript files
- `tsc --skipLibCheck` for TypeScript compilation

### Docker Compose Development

Full development environment:
```bash
docker-compose build
docker-compose up
```

Access services:
```bash
# Database
docker exec -it <postgres container ID> psql -h localhost -U user --no-password -p 5432 -d model_manager

# HTTP API
curl http://localhost:8080/v1/models

# S3/MinIO
docker exec -it <aws-cli container ID> bash
export AWS_ACCESS_KEY_ID=llmariner-key
export AWS_SECRET_ACCESS_KEY=llmariner-secret
aws --endpoint-url http://minio:9000 s3 ls s3://llmariner
```

## Architecture Overview

### Core Components

**Server** (`server/`): Main gRPC/HTTP service handling:
- Model CRUD operations (Create, List, Get, Delete)
- Model activation/deactivation
- OpenAI API compatibility
- Multi-tenant authentication and authorization

**Loader** (`loader/`): Background service for:
- Downloading models from various sources (HuggingFace, S3, Ollama)
- Converting model formats (GGUF, HuggingFace, etc.)
- Uploading processed models to object storage

### Dual-Server Architecture

1. **Public API Server** (`server/internal/server/server.go`): User-facing REST/gRPC endpoints
2. **Internal Worker Server** (`server/internal/server/ws_server.go`): Backend service communication

### Model Hierarchy

**Base Models** (`server/internal/store/base_model.go`):
- Foundation models shared within tenant
- Support multiple formats: GGUF, HuggingFace, Nvidia Triton, Ollama
- State machine: REQUESTED → LOADING → SUCCEEDED/FAILED

**Fine-Tuned Models** (`server/internal/store/model.go`):
- Project-scoped models derived from base models
- Support LoRA/QLoRA adapters
- Publishing workflow before becoming available

### Multi-Tenancy

- **Tenant-level isolation**: Base models shared within tenant
- **Project-level isolation**: Fine-tuned models scoped to projects
- **Composite unique keys**: (ModelID, TenantID) prevent conflicts

## API Structure

### Protocol Buffers

Primary service definitions in `api/v1/model_manager_service.proto`:
- `ModelsService`: Public API (OpenAI compatible)
- `ModelsWorkerService`: Internal service for backend operations

### Source Repositories

Supported model sources:
- Object Store (S3)
- HuggingFace Hub
- Ollama repositories
- Fine-tuning outputs

## Configuration

### Server Config (`configs/server/config.yaml`)
```yaml
httpPort: 8080
grpcPort: 8081
internalGrpcPort: 8082
objectStore:
  s3:
    pathPrefix: models
debug:
  standalone: true
  sqlitePath: /tmp/model_manager.db
```

### Loader Config (`configs/loader/config.yaml`)
Required for model loading operations with S3 credentials and source repository configuration.

## Database Schema

Key entities:
- `BaseModel`: Foundation models with format metadata
- `Model`: Fine-tuned models with base model references
- `ModelActivationStatus`: Separate activation control
- `StorageConfig`: Per-tenant S3 configuration
- `HFModelRepo`: HuggingFace repository tracking

## Development Patterns

### Authentication Modes
- **RBAC Integration**: Production mode with role-based access control
- **Fake Auth**: Development/testing mode (`debug.standalone: true`)

### State Management
- Atomic state transitions with concurrent update protection
- Worker coordination through acquire/update patterns
- Transactional consistency for multi-table operations

### Error Handling
Check `server/internal/store/errors.go` for domain-specific error types and handling patterns.

## Testing

Tests are co-located with source files using `_test.go` suffix. Key test patterns:
- Store layer tests with database fixtures
- Server layer tests with mock stores
- Integration tests using `testing.go` utilities

## Deployment

Helm charts available in `deployments/`:
- `deployments/server/`: Server component
- `deployments/loader/`: Loader component

Generate chart schemas:
```bash
make generate-chart-schema
```

Lint Helm charts:
```bash
make helm-lint
```