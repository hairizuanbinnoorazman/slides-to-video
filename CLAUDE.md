# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Slides-to-video is a microservices-based application that converts PDF slides into videos with text-to-speech narration. The system uses a queue-based architecture where workers process different stages of video generation.

## Build and Development Commands

### Local Development
```bash
# Build all binaries (required before building images)
make build-bin

# Build all Docker images
make build-images

# Start the full stack (manager, workers, MySQL, Minio S3, NATS queue)
make stack-up

# Stop the stack
make stack-down

# Rebuild everything and restart
make reup
```

### Running Tests
```bash
# Integration tests (requires stack to be running)
cd tests/
pipenv shell
pipenv install
pytest test_app.py

# Go unit tests in specific packages
cd acl/ && go test -v
cd project/ && go test -v
```

### Frontend Development
```bash
# Elm frontend development
cd cmd/slides-to-video-frontend
elm reactor
# Navigate to Reactor.elm in browser

# Format Elm code before committing
make format
```

### Cloud Deployment
```bash
# Build versioned images for GCP
make build-all-versioned

# Push versioned images to GCR
make push-all-versioned

# Deploy to Cloud Run (from deployment/cloud-run/)
cd deployment/cloud-run
make deploy-all
```

## Architecture

The application consists of five main services (deployable together or separately) that communicate via queues:

1. **slides-to-video-manager** (port 8880): Main API server
   - Handles HTTP requests for project CRUD operations
   - Manages project state and coordinates worker jobs
   - Enqueues jobs to worker services

2. **pdf-splitter** (port 8881): Worker service
   - Receives PDF split jobs from queue
   - Extracts individual slide images from PDFs
   - Stores images in blob storage
   - Updates manager with extracted slide information

3. **image-to-video** (port 8882): Worker service
   - Converts individual slide images to video segments
   - Generates audio narration using Google Text-to-Speech
   - Uses ffmpeg to combine image and audio into video segment
   - Requires Google Cloud credentials for TTS

4. **concatenate-video** (port 8883): Worker service
   - Concatenates all video segments into final output video
   - Uses ffmpeg for video concatenation
   - Updates project with final video location

5. **frontend** (port 8081): Elm-based UI
   - Alternative **slides-to-video-frontend-alt**: Go-based template UI

### Supporting Infrastructure
- **MySQL** (port 3306): Primary data store for local dev
- **Minio** (port 9999): S3-compatible blob storage for local dev
- **NATS** (port 4222): Message queue for distributed mode
- **Channels**: In-memory Go channels for all-in-one mode

### Production Alternatives
- Google Cloud Datastore replaces MySQL
- Google Cloud Storage replaces Minio
- Google Pub/Sub replaces NATS
- Channels queue for single-process deployments

### Deployment Modes

The application supports flexible deployment:

1. **All-in-One Mode** (single process)
   - Manager runs with embedded workers in the same process
   - Uses in-memory channels for queue communication
   - Simpler deployment, lower resource overhead
   - Config: `config-all-in-one.yaml`

2. **Distributed Mode** (separate services)
   - Each service runs independently
   - Uses NATS (local) or Pub/Sub (cloud) for messaging
   - Better scalability and isolation
   - Config: `config-distributed.yaml`

See docs/SERVICE-CONSOLIDATION.md for detailed architecture and migration guidance.

## Code Organization

### Domain Models (cmd-agnostic packages)
- **project/**: Core project entity with status tracking
- **videosegment/**: Individual video segments with order, script text, and file locations
- **pdfslideimages/**: PDF slide images extracted from uploaded PDFs
- **user/**: User accounts and authentication
- **acl/**: Access control lists for projects

### Infrastructure Abstractions
- **blobstorage/**: Storage abstraction supporting GCS and Minio
- **queue/**: Queue abstraction supporting NATS and Google Pub/Sub
- **job/**: Job tracking for async operations
- **logger/**: Structured logging

### HTTP Layer
- **handlers/**: HTTP handlers for the manager API
- **client/**: Go client library for calling manager API from workers

### Worker-Specific Logic
Each worker service has its own package under cmd/:
- **cmd/pdf-splitter/pdfsplitter/**: PDF splitting logic
- **cmd/image-to-video/image2videoconverter/**: Image-to-video conversion with TTS
- **cmd/concatenate-video/videoconcater/**: Video concatenation logic

### Storage Layer Pattern
Each domain model package contains store interfaces and implementations:
- `store.go`: Interface definition
- `mysql.go`: MySQL implementation
- `datastore.go`: Google Cloud Datastore implementation
- Tests use common test suites to verify both implementations

## Common Workflows

### Typical User Flow
1. User uploads PDF via API → creates Project and enqueues PDF split job
2. pdf-splitter worker processes PDF → creates PDFSlideImages entries
3. User adds narration text to each slide via API
4. User requests video generation → enqueues image-to-video jobs for each slide
5. image-to-video workers generate video segments with TTS audio
6. When all segments complete → enqueue concatenate-video job
7. concatenate-video worker produces final output video
8. User downloads final video from blob storage URL

### Configuration Pattern
All services use YAML configuration files loaded via cobra CLI flags:
```bash
app server -c /path/to/config.yaml
```

Configuration controls:
- Storage backend selection (MySQL vs Datastore, Minio vs GCS)
- Queue backend selection (NATS vs Pub/Sub)
- Service URLs for inter-service communication
- Credentials and secrets

### Adding New Features

When adding new domain models:
1. Create package with struct definition
2. Implement Store interface with both MySQL and Datastore backends
3. Add handlers in handlers/ package
4. Add routes in cmd/slides-to-video-manager/serve.go
5. Add client methods in client/ package if workers need access
6. Write tests using testcontainers pattern from existing tests

### Database Migrations
```bash
# Run migrations (uses GORM AutoMigrate)
docker exec -it <manager-container> app migrate -c /path/to/config.yaml
```

Note: Migration strategy is basic - planned migration to golang-migrate for production.

## Development Workflow Requirements

After making code changes, verify:

1. **Tests pass** (if code was modified)
   ```bash
   # Run unit tests for modified packages
   cd <package>/ && go test -v

   # Run integration tests
   cd tests/ && pipenv run pytest
   ```

2. **Binaries compile**
   ```bash
   make build-bin
   ```

3. **Docker images build**
   ```bash
   make build-images
   ```

These checks ensure changes don't break the build or deployment pipeline.

## Testing Strategy

### Integration Tests
The tests/ directory contains pytest-based integration tests that:
- Start the full docker-compose stack
- Test end-to-end workflows through the API
- Verify worker processing and final video generation

### Unit Tests
Go packages use testcontainers for database-dependent tests:
- Spins up MySQL container for test isolation
- Tests both MySQL and Datastore implementations
- Located alongside source files (*_test.go)

### Test Requirements
- 4 CPU cores, 4.5GB RAM minimum for local stack
- Python 3.x with pipenv for integration tests
- Docker for testcontainers

## Deployment Notes

### Local Development (docker-compose)
Services communicate via service names, use Minio for storage, NATS for queuing.

### Kubernetes (Helm)
Helm charts in deployment/helm/slides-to-video/ support:
- StatefulSet for MySQL
- Deployments for each service
- ConfigMaps for configuration
- Secrets for credentials

### Cloud Run
Services deployed as independent Cloud Run services:
- Manager allows unauthenticated access (public API)
- Workers are authenticated (called via Pub/Sub push)
- Each worker has dedicated Pub/Sub topic and push subscription
- Uses Google Cloud Datastore, GCS, and Pub/Sub

## Key Dependencies
- **ffmpeg**: Required in worker containers for video processing
- **Google Cloud Text-to-Speech API**: Required for audio narration in image-to-video worker
- **gorilla/mux**: HTTP routing
- **spf13/cobra**: CLI framework
- **jinzhu/gorm**: ORM for MySQL
- **cloud.google.com/go/datastore**: Google Datastore client
- **minio-go**: Minio/S3 client

## Known Issues and TODOs

See README.md "Features to be developed" section for comprehensive list. Key items:
- List and Delete operations for VideoSegments/PDFSlideImages need fixes
- ACL model integration incomplete (app-managed and Keycloak modes planned)
- Need proper migration tooling (planned: golang-migrate)
- Rate limiting and API security not fully implemented
- Monitoring and distributed tracing planned but not implemented
