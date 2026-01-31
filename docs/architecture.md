# Atabeh Architecture

This document describes the high-level architecture of Atabeh, explaining how components interact and why the system is designed this way.

## Overview

Atabeh follows a pipeline architecture where configurations flow through distinct stages: fetching, parsing, normalization, testing, and ranking. Each stage is independent and can be extended without affecting others.

## Core Principles

**Modularity**: Components communicate through well-defined interfaces. You can swap implementations without breaking the system.

**Extensibility**: Adding new parsers, source types, or test methods doesn't require modifying existing code.

**Separation**: The core engine has no knowledge of clients. All client interaction happens through a standard API.

**Performance**: Operations that can run concurrently do. Testing hundreds of configs should take seconds, not minutes.

**Resilience**: Malformed data or failed connections shouldn't crash the system. Atabeh degrades gracefully.

## System Components

### 1. Source Fetcher

**Responsibility**: Retrieve configuration data from various sources.

**Inputs**: Source specification (URL, file path, API endpoint)

**Outputs**: Raw data as bytes or strings

**Key Features**:
- HTTP/HTTPS fetching with retry logic
- File system reading with proper error handling
- GitHub raw link resolution
- Future: Telegram channel integration

**Why Separate**: Sources can fail or change. Isolating fetch logic means we can add retry strategies, caching, or new source types without touching parsers.

### 2. Parser System

**Responsibility**: Convert raw configuration data into structured objects.

**Inputs**: Raw configuration strings in various formats

**Outputs**: Structured configuration objects specific to each protocol

**Key Features**:
- Format detection (auto-detecting protocol types)
- Base64 decoding for encoded configs
- JSON/YAML parsing for structured formats
- Error recovery for partially malformed data
- Registry system for parser discovery

**Design**: Each parser implements a common interface:

```go
type Parser interface {
    Parse(data []byte) ([]Config, error)
    Supports(data []byte) bool
    Name() string
}
```

This allows the system to try parsers until one succeeds, and makes adding new parsers straightforward.

**Why Separate Per Protocol**: Different VPN protocols have completely different config formats. Keeping them separate means changes to VMess parsing don't risk breaking Trojan parsing.

### 3. Normalizer

**Responsibility**: Convert protocol-specific configs into a unified data model.

**Inputs**: Parsed configurations in various formats

**Outputs**: Normalized configuration objects with consistent fields

**Key Features**:
- Standardized field mapping
- Default value injection
- Validation of required fields
- Metadata enrichment

**Why Needed**: Clients and testers shouldn't need to understand every protocol's quirks. The normalized model provides a consistent interface for everything downstream.

**Normalized Model Fields**:
```
- ID: unique identifier
- Protocol: vmess, vless, trojan, etc.
- Server: hostname or IP
- Port: connection port
- Transport: tcp, ws, http, grpc, etc.
- Security: tls, reality, none
- Credentials: authentication data
- Extra: protocol-specific parameters
- Metadata: source, parsing timestamp, etc.
```

### 4. Test Engine

**Responsibility**: Evaluate configuration quality through network tests.

**Inputs**: Normalized configurations

**Outputs**: Test results with metrics

**Key Features**:
- Connection availability testing
- Latency measurement (ping)
- Timeout handling
- Concurrent testing with rate limiting
- Result aggregation

**Test Types**:

**Connectivity**: Can we establish a connection through this config?

**Latency**: How long does it take to get a response?

**Stability**: Does the connection succeed consistently over multiple attempts?

**Future Tests**:
- Bandwidth measurement
- Geographic location detection
- Protocol-specific health checks

**Why Separate**: Testing is expensive (network operations take time). By separating testing from parsing, we can test configs in parallel, cache results, and skip retesting unchanged configs.

### 5. Ranking System

**Responsibility**: Score and sort configurations by quality.

**Inputs**: Test results and normalized configs

**Outputs**: Ranked list of configurations

**Key Features**:
- Multi-factor scoring algorithm
- Weighted ranking (latency, availability, stability)
- Configurable ranking criteria
- Historical performance consideration

**Ranking Factors**:
- Latency (lower is better)
- Success rate (higher is better)
- Consistency (stable performance over time)
- Server location (optional, based on user preference)

**Why Separate**: Ranking logic might change as we learn what metrics matter most. Keeping it separate means we can experiment with different algorithms without touching the testing code.

### 6. Storage Layer

**Responsibility**: Persist configurations, test results, and historical data.

**Inputs**: Any serializable data

**Outputs**: Stored and retrieved data

**Key Features**:
- Pluggable backends (memory, SQLite, future: PostgreSQL)
- Efficient queries for config lookup
- Historical data tracking
- Cache invalidation strategies

**Why Needed**: Users don't want to retest configs every time they run Atabeh. Storage provides persistence and enables features like historical performance tracking.

### 7. API Server

**Responsibility**: Expose Atabeh functionality to client applications.

**Inputs**: Client requests via HTTP/gRPC

**Outputs**: JSON/Protocol Buffer responses

**Key Features**:
- RESTful HTTP endpoints
- gRPC service for efficiency
- Request validation
- Error handling and proper status codes
- Future: Authentication and rate limiting

**Endpoints**:
- Add source: `POST /sources`
- List configs: `GET /configs`
- Test configs: `POST /test`
- Get rankings: `GET /rankings`
- Health check: `GET /health`

**Why Both REST and gRPC**: REST is universal and easy to debug. gRPC is efficient for high-frequency requests and works better for streaming results.

## Data Flow

### Typical Operation Flow

1. **Client Request**: User adds a configuration source through the API
2. **Fetch**: Source fetcher retrieves raw data from the specified location
3. **Parse**: Parser system detects format and extracts configurations
4. **Normalize**: Normalizer converts configs to unified model
5. **Store**: Configurations are saved to storage
6. **Test**: Test engine evaluates each config's quality
7. **Update**: Test results are stored alongside configs
8. **Rank**: Ranking system scores all configs
9. **Response**: API returns ranked list to client

### Concurrent Operations

Many operations happen in parallel:
- Multiple sources can be fetched simultaneously
- Configs are tested concurrently (with rate limiting)
- Different clients can make API requests independently

The system uses Go's goroutines and channels for safe concurrent operations.

## Client-Server Architecture

### Core Engine (Server Side)

The core runs as a service (daemon, API server, or embedded in an application). It handles all the heavy work:
- Fetching data
- Parsing formats
- Testing connections
- Managing storage

### Clients (User Side)

Clients are lightweight applications that communicate with the core:
- Android/iOS mobile apps
- Desktop applications (Windows, macOS, Linux)
- Command-line tools
- Web interfaces

Clients handle only:
- User interface
- Requesting operations from the core
- Displaying results
- Platform-specific features

### Communication Protocol

Clients talk to the core via:
- HTTP REST API for simple operations
- gRPC for efficient bulk operations
- WebSocket for real-time updates (future)

This separation means:
- Core can run on a powerful server, lightweight clients on less capable devices
- Multiple clients can share one core instance
- Clients can be developed independently
- Platform-specific functionality doesn't complicate the core

## Error Handling Strategy

**Graceful Degradation**: A failed source fetch doesn't stop processing other sources. A malformed config doesn't crash the parser.

**Clear Feedback**: Errors include enough context to understand what went wrong and potentially fix it.

**Recovery**: When possible, the system extracts partial data from corrupted inputs.

**Logging**: All errors are logged with appropriate severity levels for debugging.

## Concurrency Model

**Parser Parallelism**: Multiple parsers can run simultaneously on different data sources.

**Test Parallelism**: Configs are tested concurrently with a worker pool to prevent overwhelming the network.

**Thread Safety**: Shared data structures use appropriate synchronization primitives.

**Rate Limiting**: Testing respects rate limits to avoid triggering anti-abuse mechanisms.

## Security Considerations

**Input Validation**: All external data is validated before processing.

**No Arbitrary Code Execution**: Parsers don't eval or execute configuration data as code.

**Isolation**: Test connections run in isolated contexts.

**No Data Leakage**: The core doesn't send user configs to external services without explicit consent.

## Performance Characteristics

**Parsing Speed**: Single-threaded parsers process thousands of configs per second.

**Testing Speed**: Limited by network latency, not CPU. Can test 100+ configs concurrently.

**Memory Usage**: Minimal, with streaming parsers for large datasets.

**Storage Growth**: Linear with number of configs and test history depth.

## Extension Points

The architecture explicitly supports extension:

**New Parsers**: Implement the Parser interface, register in the registry.

**New Source Types**: Implement the Fetcher interface.

**New Test Methods**: Add to the test engine without modifying existing tests.

**New Storage Backends**: Implement the Storage interface.

**New Ranking Algorithms**: Plug in different scoring functions.

## Future Architectural Considerations

**Machine Learning**: Could add a prediction layer that learns which configs will perform well based on patterns.

**Distributed Testing**: Could split testing across multiple nodes for faster results.

**Config Sharing**: P2P layer for sharing and verifying configs between users.

**Plugin System**: More formalized plugin architecture for third-party extensions.

## Design Trade-offs

**Simplicity vs. Features**: We favor simple, working implementations over feature-complete complexity. Add features when they're needed, not speculatively.

**Performance vs. Reliability**: When forced to choose, we prefer reliable operation over maximum speed. Most users care more about correctness than shaving milliseconds.

**Flexibility vs. Constraints**: The interface-based design adds some boilerplate but makes the system vastly more extensible.

## Why This Architecture?

This design emerged from real requirements:

- Users need to test many configs quickly → concurrent testing
- Configs come in many formats → pluggable parsers
- New protocols appear regularly → easy parser addition
- Clients run on different platforms → separate client/server
- Network conditions vary → robust error handling

The architecture solves these problems without over-engineering. As requirements evolve, the modular design allows us to enhance components independently without redesigning the entire system.