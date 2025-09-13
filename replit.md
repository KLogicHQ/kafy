# Overview

This project is a Kafka Productivity CLI tool called "kkl" - designed to be the "kubectl for Kafka". It provides a comprehensive, unified command-line interface that replaces the complex native `kafka-*` shell scripts with intuitive, short commands.

**Status**: ✅ FULLY IMPLEMENTED - All core functionality is working

The tool simplifies Kafka operations by providing declarative commands, context switching between environments (dev, staging, prod), and unified output formats for both human operators and automation scripts.

## Key Features Implemented:
- Complete config management with YAML-based cluster configuration
- Full topic management (create, list, describe, delete)
- Consumer group operations
- Message production and consumption with multiple formats
- Broker inspection and health monitoring  
- Multiple output formats (table, JSON, YAML)
- Context switching similar to kubectl
- Built using confluent-kafka-go v2 library

# User Preferences

Preferred communication style: Simple, everyday language.

# System Architecture

## CLI Command Structure
The application follows a hierarchical command structure similar to kubectl:
- Root command: `kkl [command] [subcommand] [flags]`
- Main command categories: config, topics, groups
- Context-aware operations that work across different environments

## Configuration Management
- **Config Storage**: Uses YAML configuration files stored in `~/.kkl/config.yml`
- **Context System**: Implements a kubectl-like context switching mechanism
- **Multi-Environment Support**: Manages multiple Kafka clusters (dev, staging, prod) through named contexts
- **Cluster Configuration**: Stores bootstrap servers and optional Zookeeper endpoints per cluster

## Command Categories

### Config Management (`kkl config`)
- Cluster configuration CRUD operations
- Context switching and management
- Import/export functionality for configuration portability

### Topic Management (`kkl topics`) 
- Topic lifecycle operations (list, create, delete, alter)
- Topic configuration management
- Partition and replication factor configuration

### Consumer Group Management (`kkl groups`)
- Consumer group monitoring and management
- Lag analysis and offset management
- Group reset operations for different offset strategies

## Output Strategy
- **Dual Format Support**: Human-readable tables for interactive use, JSON/YAML for automation
- **Consistency**: Unified output formatting across all commands
- **Scriptability**: Machine-readable formats for CI/CD integration

# Technical Implementation

## Go Dependencies
- **confluent-kafka-go v2**: Primary Kafka client library for Go
- **cobra**: CLI framework for command structure and parsing
- **go-pretty/v6**: Table formatting for human-readable output
- **yaml.v3**: YAML configuration file parsing

## External Dependencies
- **Apache Kafka**: Primary message broker system (tested with modern versions)
- **Bootstrap Servers**: Kafka broker endpoints for cluster connectivity
- **Optional Zookeeper**: For legacy cluster management (modern Kafka doesn't require this)

## Build and Usage
```bash
# Build the CLI
go build -o kkl .

# Get help
./kkl --help

# Add a cluster configuration
./kkl config add my-cluster --bootstrap "kafka-server:9092"

# List topics
./kkl topics list

# Create a topic
./kkl topics create orders --partitions 3 --replication 2

# Produce test messages
./kkl produce orders --count 10

# Consume messages
./kkl consume orders --from-beginning --limit 5
```