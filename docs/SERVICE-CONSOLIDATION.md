# Service Consolidation and Go Channels Queue

## Overview

The slides-to-video application now supports two deployment architectures:

1. **All-in-One Mode**: Single process with embedded workers using Go channels
2. **Distributed Mode**: Separate services communicating via NATS or Google Pub/Sub

This provides flexibility for different deployment scenarios:
- **Development**: All-in-one mode simplifies local development
- **Small Deployments**: All-in-one mode reduces operational complexity
- **Production**: Distributed mode enables horizontal scaling and fault isolation

## What Changed

### New Components

1. **Go Channels Queue** (`queue/channels.go`)
   - In-memory queue implementation using Go channels
   - Lightweight alternative to NATS/Pub/Sub for single-process deployments
   - Buffered channels prevent blocking on message publish
   - Context-aware operations support graceful shutdown

2. **Workers Package** (`cmd/slides-to-video-manager/workers/`)
   - Generic worker runner that consumes queue messages
   - Worker wrappers for pdf-splitter, image-to-video, and concatenate-video
   - Reuses existing processor logic from worker services

3. **Configuration Extensions**
   - New `workers` section controls embedded worker behavior
   - New `channels` queue type for in-process communication
   - Added folder configurations for worker blob storage paths

### Architecture Modes

#### Mode 1: All-in-One (Channels Queue)

```
┌─────────────────────────────────────────┐
│      slides-to-video-manager            │
│                                          │
│  ┌──────────┐    ┌─────────────────┐   │
│  │   API    │───▶│  Go Channels    │   │
│  │  Server  │    │  (in-memory)    │   │
│  └──────────┘    └─────────────────┘   │
│                           │              │
│         ┌─────────────────┴─────────┐   │
│         ▼                 ▼         ▼   │
│  ┌──────────┐    ┌──────────┐  ┌────┐  │
│  │ PDF      │    │ Image-to │  │Vid │  │
│  │ Splitter │    │ Video    │  │Cat │  │
│  │ Worker   │    │ Worker   │  │    │  │
│  └──────────┘    └──────────┘  └────┘  │
│                                          │
└─────────────────────────────────────────┘
        │
        ▼
  MySQL + Minio
```

**Configuration:**
```yaml
workers:
  enabled: true
  pdfSplitter:
    enabled: true
    concurrency: 1
  imageToVideo:
    enabled: true
    concurrency: 2
  concatenateVideo:
    enabled: true
    concurrency: 1

queue:
  type: "channels"
  channels:
    bufferSize: 100
    pdfToImageTopic: "pdf-splitter"
    imageToVideoTopic: "image-to-video"
    videoConcatTopic: "concatenate-video"
```

**Use Cases:**
- Local development
- Single-server deployments
- Cost-sensitive environments
- Testing and CI/CD pipelines

**Pros:**
- Simple deployment (single binary)
- No external queue infrastructure required
- Lower latency (no network calls)
- Lower resource usage

**Cons:**
- Cannot scale workers independently
- Manager crash affects all workers
- No distributed processing
- Limited by single machine resources

#### Mode 2: Distributed (NATS/Pub/Sub)

```
┌────────────────┐      ┌──────────┐
│ Manager        │─────▶│  NATS /  │
└────────────────┘      │  Pub/Sub │
                        └──────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────┐
    │ PDF Splitter │  │ Image2Video  │  │ VideoCat │
    │  (worker)    │  │   (worker)   │  │ (worker) │
    └──────────────┘  └──────────────┘  └──────────┘
```

**Configuration (Manager):**
```yaml
workers:
  enabled: false  # Workers run as separate services

queue:
  type: "nats"
  nats:
    endpoint: "nats://localhost:4222"
    pdfToImageTopic: "pdf-splitter"
    imageToVideoTopic: "image-to-video"
    videoConcatTopic: "concatenate-video"
```

**Use Cases:**
- Production deployments
- High-availability requirements
- Need for horizontal scaling
- Fault isolation between components

**Pros:**
- Scale workers independently
- Fault isolation (worker crash doesn't affect manager)
- Horizontal scaling
- Distributed processing across machines

**Cons:**
- More complex deployment
- Requires NATS/Pub/Sub infrastructure
- Higher latency (network calls)
- Higher resource usage

#### Mode 3: Hybrid

You can selectively enable/disable workers:

```yaml
workers:
  enabled: true
  pdfSplitter:
    enabled: true   # Embedded
    concurrency: 1
  imageToVideo:
    enabled: false  # Separate service
  concatenateVideo:
    enabled: true   # Embedded
    concurrency: 1

queue:
  type: "nats"  # Must use NATS/Pub/Sub for separate workers
  nats:
    endpoint: "nats://localhost:4222"
    # ...
```

## Configuration Reference

### Workers Section

```yaml
workers:
  enabled: true/false          # Enable/disable all embedded workers
  pdfSplitter:
    enabled: true/false        # Enable PDF splitter worker
    concurrency: 1             # Number of concurrent workers (default: 1)
  imageToVideo:
    enabled: true/false        # Enable image-to-video worker
    concurrency: 2             # Number of concurrent workers (default: 1)
  concatenateVideo:
    enabled: true/false        # Enable concatenate-video worker
    concurrency: 1             # Number of concurrent workers (default: 1)
```

### Channels Queue Section

```yaml
queue:
  type: "channels"
  channels:
    bufferSize: 100            # Channel buffer size (default: 100)
    pdfToImageTopic: "pdf-splitter"
    imageToVideoTopic: "image-to-video"
    videoConcatTopic: "concatenate-video"
```

### Blob Storage Folder Configuration

New folder fields required for workers:

```yaml
blobStorage:
  type: "minio"  # or "gcs"
  minio:         # or gcs:
    bucket: "videos"
    endpoint: "localhost:9000"
    accessKeyId: "minioadmin"
    secretAccessKey: "minioadmin"
    pdfFolder: "pdf"                           # Required
    imagesFolder: "images"                     # NEW - Required for workers
    videoSnippetsFolder: "video-snippets"      # NEW - Required for workers
    videoFolder: "videos"                      # NEW - Required for workers
```

## Migration Guide

### From Distributed to All-in-One

1. Update configuration to enable workers:
   ```yaml
   workers:
     enabled: true
     pdfSplitter:
       enabled: true
       concurrency: 1
     # ... enable other workers
   ```

2. Change queue type to channels:
   ```yaml
   queue:
     type: "channels"
     channels:
       bufferSize: 100
       # ... configure topics
   ```

3. Add folder configurations to blobStorage

4. Stop separate worker services

5. Restart manager with new configuration

### From All-in-One to Distributed

1. Update configuration to disable workers:
   ```yaml
   workers:
     enabled: false
   ```

2. Change queue type to nats or google_pubsub:
   ```yaml
   queue:
     type: "nats"
     nats:
       endpoint: "nats://localhost:4222"
       # ... configure topics
   ```

3. Deploy separate worker services with their own configurations

4. Restart manager with new configuration

## Running the Application

### All-in-One Mode

```bash
# Build the manager
make build-bin

# Run with all-in-one configuration
./bin/slides-to-video-manager server -c cmd/slides-to-video-manager/configuration/config-all-in-one.yaml
```

### Distributed Mode

```bash
# Build all binaries
make build-bin

# Start infrastructure (MySQL, Minio, NATS)
docker-compose -f docker-compose-infra.yaml up -d

# Start manager
./bin/slides-to-video-manager server -c cmd/slides-to-video-manager/configuration/config-distributed.yaml

# Start workers (in separate terminals)
./bin/pdf-splitter server -c cmd/pdf-splitter/configuration/config.yaml
./bin/image-to-video server -c cmd/image-to-video/configuration/config.yaml
./bin/concatenate-video server -c cmd/concatenate-video/configuration/config.yaml
```

## Performance Considerations

### Buffer Size

The `bufferSize` parameter controls the channel buffer:
- **Larger buffer**: Better throughput, more memory usage
- **Smaller buffer**: Less memory, may cause blocking
- **Recommended**: 100 for general use, tune based on workload

### Worker Concurrency

Each worker type can have multiple concurrent instances:
- **PDF Splitter**: CPU-intensive (ImageMagick), keep concurrency low
- **Image-to-Video**: CPU-intensive (ffmpeg), moderate concurrency
- **Concatenate Video**: CPU-intensive (ffmpeg), usually 1 is sufficient

Example tuning:
```yaml
workers:
  pdfSplitter:
    concurrency: 1   # CPU-bound, limit to CPU cores
  imageToVideo:
    concurrency: 2   # Can benefit from parallelism
  concatenateVideo:
    concurrency: 1   # Usually sequential
```

## Troubleshooting

### Workers not processing jobs

Check:
1. Workers are enabled in configuration
2. Queue type matches configuration (channels requires embedded workers)
3. Blob storage folder paths are configured
4. Manager logs show worker startup messages

### Channel buffer full

Symptoms:
- Jobs timeout
- Slow job processing

Solutions:
- Increase `bufferSize` in channels configuration
- Increase worker `concurrency`
- Check if workers are processing successfully

### Manager doesn't start

Common issues:
- Text-to-speech credentials missing for image-to-video worker
- Blob storage configuration incomplete
- Database connection issues

Check manager logs for specific errors.

## Testing

### Unit Tests

```bash
# Test channels queue
go test ./queue -v -run TestChannels

# Test all packages
go test ./... -v
```

### Integration Tests

The existing integration tests in `tests/` work with both modes:

```bash
cd tests/
pipenv shell
pipenv install

# Test all-in-one mode
# (configure manager with workers enabled)
pytest test_app.py

# Test distributed mode
# (configure manager with workers disabled, start worker services)
pytest test_app.py
```

## Monitoring

### All-in-One Mode

Monitor single process:
- Manager logs include worker activity
- All workers share manager's resource limits
- Single health check endpoint

### Distributed Mode

Monitor multiple processes:
- Each service has its own logs
- Independent resource monitoring
- Separate health checks for each service

## Security Considerations

### All-in-One Mode

- Single process runs all operations
- Credentials needed for all worker operations (TTS, blob storage)
- Process crash affects entire system

### Distributed Mode

- Credential isolation between services
- More attack surface (multiple processes)
- Better fault containment

## Future Improvements

Planned enhancements:
1. **Internal Client**: Replace HTTP mgrclient with direct store access for embedded workers
2. **Metrics**: Expose queue depth and worker metrics
3. **Health Checks**: Enhanced health checks for embedded workers
4. **Graceful Shutdown**: Improved worker shutdown handling
5. **Rate Limiting**: Per-worker rate limiting configuration

## Related Files

- `queue/channels.go` - Channels queue implementation
- `queue/channels_test.go` - Channels queue tests
- `cmd/slides-to-video-manager/workers/` - Worker wrappers
- `cmd/slides-to-video-manager/config.go` - Configuration structures
- `cmd/slides-to-video-manager/serve.go` - Worker initialization
- `cmd/slides-to-video-manager/configuration/config-all-in-one.yaml` - Example all-in-one config
- `cmd/slides-to-video-manager/configuration/config-distributed.yaml` - Example distributed config
