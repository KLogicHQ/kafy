# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**kafy** is a comprehensive Kafka productivity CLI tool written in Go that simplifies Kafka operations with a kubectl-like design philosophy. It provides context-aware operations, intelligent tab completion, and multiple output formats for managing Kafka clusters, topics, consumer groups, brokers, and messages.

## Development Commands

### Building and Testing
```bash
# Development build for current platform
go build -o kafy .

# Run directly from source
go run . --help
go run . topics list

# Multi-platform production builds
./build.sh

# Docker build for Linux platforms
docker buildx build --platform linux/amd64 -t kafy:latest .

# Test basic functionality
go run . config add test --bootstrap localhost:9092
go run . topics list --output json
```

### Dependencies Management
```bash
# Install/update dependencies
go mod tidy
go mod download

# Verify module integrity
go mod verify
```

### Testing Against Real Kafka
```bash
# Set up local test cluster
go run . config add local --bootstrap localhost:9092
go run . config use local
go run . health check

# Run integration tests (requires running Kafka cluster)
go run . topics create test-topic --partitions 1 --replication 1
go run . produce test-topic --count 10
go run . consume test-topic --from-beginning --limit 5
```

## Architecture Overview

### Project Structure
```
kafy/
├── main.go                 # Application entry point - simple cmd.Execute() wrapper
├── cmd/                    # CLI commands using cobra framework
│   ├── root.go            # Root command definition, global flags, subcommand registration
│   ├── topics.go          # Topic management (create, list, describe, delete, configs)
│   ├── groups.go          # Consumer group operations and lag monitoring
│   ├── produce.go         # Message production with various input methods
│   ├── consume.go         # Message consumption with filtering and formatting
│   ├── brokers.go         # Broker inspection, metrics, and configuration
│   ├── config.go          # Cluster configuration and context switching
│   ├── health.go          # Cluster health checks and diagnostics
│   ├── offsets.go         # Partition offset management
│   ├── cp.go              # Message copying between topics
│   ├── tail.go            # Real-time message tailing
│   ├── util.go            # Utility commands
│   ├── version.go         # Version information
│   └── completions.go     # Shell tab completion logic
├── internal/
│   ├── kafka/client.go    # Kafka client wrapper using confluent-kafka-go
│   ├── output/formatter.go # Output formatting (table, JSON, YAML)
│   └── ai/client.go       # AI integration for metrics analysis
├── config/config.go       # Configuration file management (~/.kafy/config.yml)
├── build.sh              # Multi-platform build script
└── Dockerfile            # Multi-stage Docker build
```

### Key Architectural Patterns

1. **Command Structure**: Uses cobra framework with a kubectl-inspired command hierarchy
   - Global flags: `--output`, `--cluster` available on all commands
   - Consistent subcommand patterns: `kafy <resource> <action> [args] [flags]`

2. **Configuration Management**:
   - YAML-based config stored in `~/.kafy/config.yml`
   - Context switching between multiple clusters (dev/staging/prod)
   - Temporary cluster override with `-c/--cluster` flag

3. **Kafka Client Abstraction**:
   - `internal/kafka/client.go` wraps confluent-kafka-go library
   - Handles connection management, authentication (SASL/SSL)
   - Provides high-level operations for all Kafka resources

4. **Output Formatting**:
   - Unified formatter in `internal/output/formatter.go`
   - Supports table (default), JSON, and YAML output
   - Consistent formatting across all commands

5. **Tab Completion**:
   - Dynamic completion for Kafka resources (topics, groups, brokers)
   - Context-aware completion based on current cluster
   - Implemented in `cmd/completions.go`

## Key Components

### Version Management
- Version constant in `cmd/root.go` (currently "0.0.6")
- Must match git tags for releases (validated by CI/CD)
- Update version when making releases

### Kafka Client Integration
- Uses confluent-kafka-go/v2 library for Kafka operations
- Client configuration supports SASL authentication, SSL/TLS
- Connection pooling and error handling managed internally

### AI Integration (Optional)
- `internal/ai/client.go` provides AI-powered metrics analysis
- Supports OpenAI, Claude, Grok, and Gemini providers
- Used with `kafy brokers metrics --analyze` command

### Configuration Schema
```yaml
current-context: dev
clusters:
  dev:
    bootstrap: localhost:9092
    broker-metrics-port: 9308
  prod:
    bootstrap: kafka-prod:9092
    broker-metrics-port: 9308
    security:
      sasl:
        mechanism: SCRAM-SHA-512
        username: prod-user
        password: prod-pass
      ssl: true
```

## Development Guidelines

### Adding New Commands
1. Create new command file in `cmd/` directory
2. Follow existing patterns from `cmd/topics.go`, `cmd/groups.go`
3. Register command in `cmd/root.go` init() function
4. Add tab completion logic in `cmd/completions.go`
5. Use consistent flag naming and output formatting

### Code Conventions
- Import internal Kafka client as `kafkaClient`
- Use the existing `getFormatter()` pattern for output
- Add appropriate error handling with `handleError()`
- Follow Go naming conventions and add documentation
- Use cobra command patterns with consistent flag definitions

### Output Formatting
- Always support `--output` flag with table/json/yaml formats
- Use `getFormatter().Output()` for structured data
- Use `getFormatter().OutputTable()` for tabular data
- Ensure JSON/YAML output is automation-friendly

### Testing Patterns
```bash
# Test command structure
go run . <command> --help

# Test with different output formats
go run . <command> <subcommand> --output json
go run . <command> <subcommand> --output yaml

# Test cluster switching
go run . <command> <subcommand> -c different-cluster
```

## Release Process

### Version Updates
1. Update version constant in `cmd/root.go`
2. Commit change: `git commit -m "Bump version to x.y.z"`
3. Create and push tag: `git tag -a vx.y.z -m "Release version x.y.z"`
4. Push: `git push origin main && git push origin vx.y.z`

### Build Pipeline
- GitHub Actions automatically builds for all platforms on tag push
- Validates version match between tag and source code
- Creates distribution packages: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- Publishes to GitHub Releases with auto-generated release notes

### Manual Build Testing
```bash
# Test build script locally
./build.sh

# Check generated artifacts
ls -la release/dist/

# Test Docker build
docker buildx build --platform linux/amd64 --target final .
```

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/confluentinc/confluent-kafka-go/v2` - Kafka client
- `github.com/jedib0t/go-pretty/v6` - Table formatting
- `gopkg.in/yaml.v3` - YAML parsing

### System Dependencies
- **Go 1.21+** - Required for building
- **librdkafka** - C library for Kafka connectivity
- **Docker** (optional) - For cross-platform builds

### Platform-Specific Setup
- **Linux**: `apt-get install librdkafka-dev`
- **macOS**: `brew install librdkafka`
- **Windows**: Uses vcpkg for librdkafka in CI/CD

## Common Development Tasks

### Adding a New Kafka Operation
1. Add the operation to `internal/kafka/client.go`
2. Create/extend command in appropriate `cmd/*.go` file
3. Add tab completion for new flags/arguments
4. Test with real Kafka cluster
5. Update documentation if needed

### Debugging Connection Issues
```bash
# Check current config
go run . config current

# Test broker connectivity
go run . health brokers

# Test with specific cluster
go run . health check -c cluster-name
```

### Testing Output Formats
```bash
# Verify all output formats work
go run . topics list --output table
go run . topics list --output json
go run . topics list --output yaml
```