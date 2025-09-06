# 🚀 kaf - A Unified CLI for Kafka

A comprehensive Kafka productivity CLI tool that simplifies Kafka operations with a **kubectl-like design philosophy**. Replace complex native `kafka-*` shell scripts with intuitive, short commands and intelligent tab completion.

## 🎯 Features

- **Context-aware operations** - Switch between dev, staging, and prod clusters seamlessly
- **Unified command structure** - kubectl-inspired commands for all Kafka operations
- **Intelligent tab completion** - Auto-complete topics, consumer groups, brokers, and more
- **Multiple output formats** - Human-readable tables, JSON, and YAML for automation
- **Comprehensive topic management** - Create, list, describe, delete, and configure topics
- **Message operations** - Produce and consume messages with various formats
- **Consumer group management** - Monitor, reset, and manage consumer groups
- **Broker monitoring** - Inspect brokers, configurations, and health
- **Advanced configuration** - Topic-level and broker-level config management
- **Offset management** - Show and reset partition offsets
- **Health monitoring** - Complete cluster health diagnostics
- **Configuration management** - Store and switch between multiple cluster configs

## 🛠 Installation

### Get Binary from Releases

**Recommended for most users** - Download pre-built binaries from the [GitHub Releases](https://github.com/KLogicHQ/kaf/releases) page:

1. **Visit the releases page**: Go to [https://github.com/KLogicHQ/kaf/releases](https://github.com/KLogicHQ/kaf/releases)
2. **Download for your platform**:
   - **Linux**: `kaf-v1.0.0-linux-amd64.tar.gz` (x86_64) or `kaf-v1.0.0-linux-arm64.tar.gz` (ARM64)
   - **macOS**: `kaf-v1.0.0-darwin-amd64.tar.gz` (Intel) or `kaf-v1.0.0-darwin-arm64.tar.gz` (Apple Silicon)
   - **Windows**: `kaf-v1.0.0-windows-amd64.zip`
3. **Extract the archive**:
   ```bash
   # Linux/macOS
   tar -xzf kaf-v1.0.0-linux-amd64.tar.gz

   # Windows (PowerShell)
   Expand-Archive kaf-v1.0.0-windows-amd64.zip
   ```
4. **Move to PATH** (optional but recommended):
   ```bash
   # Linux/macOS
   sudo mv kaf /usr/local/bin/

   # Windows: Add the extracted folder to your PATH environment variable
   ```
5. **Verify installation**:
   ```bash
   kaf --help
   ```

### Build from Source

**For developers and contributors** - Build from source code:

#### Prerequisites

- **Go 1.21+** - Required for compilation
- **librdkafka development libraries** - Required for Kafka connectivity
- **Git** - For cloning the repository
- **Docker** (optional) - For cross-platform builds

#### Build Instructions

```bash
# Clone the repository
git clone https://github.com/KLogicHQ/kaf.git
cd kaf

# Download Go dependencies
go mod tidy

# Build for current platform
go build -o kaf .

# Optional: Install globally
sudo mv kaf /usr/local/bin/  # Linux/macOS
# Or add to PATH on Windows
```

#### Cross-Platform Build

Use the included build script for multi-platform binaries:

```bash
# Make executable and run
chmod +x build.sh
./build.sh

# Find built packages in release/dist/
ls release/dist/
```

The build script will:
- Build for Linux (amd64, arm64) using Docker when available
- Attempt native cross-compilation for macOS and Windows
- Create distribution packages (tar.gz for Linux/macOS, zip for Windows)
- Provide detailed build status for each platform


## ⚡ Tab Completion Setup

kaf provides intelligent tab completion for all shells. Set it up once and get auto-completion for topics, consumer groups, broker IDs, cluster names, and command flags.

### Bash Completion

```bash
# For current session only
source <(kaf completion bash)

# Install permanently on Linux
kaf completion bash | sudo tee /etc/bash_completion.d/kaf

# Install permanently on macOS (with Homebrew bash-completion)
kaf completion bash > $(brew --prefix)/etc/bash_completion.d/kaf
```

### Zsh Completion

```bash
# For current session only
source <(kaf completion zsh)

# Install permanently
kaf completion zsh > ~/.zsh/completions/_kaf
# Then add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)

# Or for oh-my-zsh users
kaf completion zsh > ~/.oh-my-zsh/completions/_kaf
```

### Fish Completion

```bash
kaf completion fish > ~/.config/fish/completions/kaf.fish
```

### PowerShell Completion

```powershell
kaf completion powershell | Out-String | Invoke-Expression
```

### What Gets Auto-Completed

- **Topics**: `kaf topics describe <TAB>` → Shows available topics
- **Consumer Groups**: `kaf groups describe <TAB>` → Shows active groups
- **Brokers**: `kaf brokers describe <TAB>` → Shows broker IDs
- **Clusters**: `kaf config use <TAB>` → Shows configured clusters
- **Flags**: `--output <TAB>` → Shows table, json, yaml options

## 🚀 Quick Start

### 1. Configure Your First Cluster

```bash
# Add a development cluster with metrics support
kaf config add dev --bootstrap "localhost:9092" --broker-metrics-port 9308

# Add a production cluster with authentication and metrics
kaf config add prod --bootstrap "kafka-prod:9092" --broker-metrics-port 9308

# List configured clusters
kaf config list

# Switch between clusters
kaf config use prod
kaf config current
```

### 2. Work with Topics

```bash
# List all topics
kaf topics list

# View partition details with insync status
kaf topics partitions              # All topics
kaf topics partitions orders       # Specific topic

# Create a new topic
kaf topics create orders --partitions 3 --replication 2

# Describe a topic
kaf topics describe orders

# Delete a topic (with confirmation)
kaf topics delete test-topic

# Manage topic configurations
kaf topics configs list      # List configs for ALL topics
kaf topics configs get orders # Get configs for specific topic
kaf topics configs set orders retention.ms=86400000
```

### 3. Produce Messages

```bash
# Interactive message production
kaf produce orders

# Produce from a file
echo '{"order_id": 123, "amount": 99.99}' > order.json
kaf produce orders --file order.json --format json

# Generate test messages
kaf produce orders --count 10

# Produce with a specific key
kaf produce orders --key "customer-123"
```

### 4. Consume Messages

```bash
# Consume messages interactively
kaf consume orders

# Consume from beginning with limit
kaf consume orders --from-beginning --limit 20

# Consume from latest messages only
kaf consume orders --from-latest

# Consume with specific consumer group
kaf consume orders --group my-service

# Output in JSON format
kaf consume orders --output json --limit 5

# Tail messages in real-time (like tail -f)
kaf tail orders
```

## 🆕 Recent Enhancements

### Enhanced Partition Management

- **`kaf topics partitions`** - New command showing detailed partition information:
  - **INSYNC Status**: Automatically calculates "yes/no" based on REPLICAS and ISR array matches
  - **All Topics**: View partition details across all topics in the cluster
  - **Single Topic**: Focus on partitions for a specific topic
  - **Complete Details**: Shows topic, partition ID, leader, replicas, ISRs, and sync status

```bash
# Example output:
# TOPIC     | PARTITION | LEADER | REPLICAS | ISRS     | INSYNC
# orders    | 0         | 1      | [1,2,3]  | [1,2,3]  | yes
# orders    | 1         | 2      | [2,3,1]  | [2,3]    | no
```

### Enhanced Configuration Management

- **`kaf config update`** - New command for modifying existing cluster configurations:
  - **Update Any Field**: Modify bootstrap servers, zookeeper, metrics port, or cluster name
  - **Metrics Port Support**: Add or update `--broker-metrics-port` for Prometheus monitoring
  - **Rename Clusters**: Change cluster names while preserving all settings
  - **Context Preservation**: Automatically updates current context when renaming active cluster

- **Enhanced Display**: All config commands now show complete cluster information:
  - **Metrics Port**: Displays configured Prometheus metrics port in all views
  - **Zookeeper Settings**: Shows zookeeper configuration when present
  - **Complete Export**: Import/export preserves all fields including metrics settings

```bash
# Update cluster settings:
kaf config update prod --broker-metrics-port 9308 --zookeeper zk-prod:2181

# Rename cluster:
kaf config update old-name --name new-name

# Enhanced display includes metrics port:
# Context  Bootstrap        Zookeeper  Metrics Port  Current
# dev      localhost:9092   -          9308          *
# prod     kafka-prod:9092  zk:2181    9308
```

### Enhanced Consumer Group Management

- **`kaf groups list`** - Now shows comprehensive group information:
  - **Group State**: Real-time state (Active, PreparingRebalance, etc.)
  - **Member Count**: Total number of active group members
  - **Real-time Data**: Uses Kafka admin API for accurate information

- **`kaf groups describe`** - Enhanced member details display:
  - **Member Information**: Shows Member ID, Client ID, and Host
  - **Topic Assignments**: Displays assigned topic:partition pairs
  - **Two-table Format**: Summary table + detailed members table
  - **Complete Visibility**: Full consumer group health monitoring

```bash
# Enhanced groups list output:
# Group ID          | State    | Members
# payment-service   | Active   | 3
# user-processor    | Rebalancing | 0

# Enhanced groups describe output includes member details:
# === MEMBERS ===
# Member ID        | Client ID    | Host           | Assignments
# consumer-1-abc   | my-service   | app-1.local    | orders:0, orders:1
# consumer-2-def   | my-service   | app-2.local    | orders:2
```

## 📖 Complete Command Reference

### Configuration Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf config list` | List all configured clusters | Show cluster overview with metrics port and current context |
| `kaf config current` | Show current active cluster | Display active cluster details including metrics port |
| `kaf config use <name>` | Switch to different cluster | `kaf config use prod` |
| `kaf config add <name> --bootstrap <server>` | Add new cluster | `kaf config add dev --bootstrap localhost:9092 --broker-metrics-port 9308` |
| `kaf config update <name>` | Update existing cluster configuration | `kaf config update dev --broker-metrics-port 9309 --zookeeper zk:2181` |
| `kaf config delete <name>` | Remove cluster | `kaf config delete old-cluster` |
| `kaf config rename <old> <new>` | Rename cluster | `kaf config rename dev development` |
| `kaf config export` | Export config to YAML/JSON | Backup or share configurations |
| `kaf config import <file>` | Import config from file | Restore from backup |

### Topic Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf topics list` | List all topics | Show topics with partition/replication info |
| `kaf topics describe <topic>` | Show detailed topic information | `kaf topics describe orders` |
| `kaf topics partitions [topic]` | Show partition details with insync status | `kaf topics partitions orders` or `kaf topics partitions` |
| `kaf topics create <topic>` | Create new topic | `kaf topics create events --partitions 6 --replication 3` |
| `kaf topics delete <topic>` | Delete topic | `kaf topics delete test-topic --force` |
| `kaf topics alter <topic>` | Modify topic settings | `kaf topics alter orders --partitions 10` |

### Topic Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf topics configs list` | List configurations for all topics | Shows configs for every topic in cluster |
| `kaf topics configs get <topic>` | Show topic-specific configs | `kaf topics configs get orders` |
| `kaf topics configs set <topic> <key>=<value>` | Set topic configuration | `kaf topics configs set orders retention.ms=86400000` |
| `kaf topics configs delete <topic> <key>` | Remove config override | `kaf topics configs delete orders cleanup.policy` |

### Consumer Group Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf groups list` | List all consumer groups | Show groups with states and member counts |
| `kaf groups describe <group>` | Show detailed group information with members | `kaf groups describe my-service` |
| `kaf groups lag <group>` | Show consumer lag metrics | `kaf groups lag payment-processor` |
| `kaf groups reset <group>` | Reset consumer offsets | `kaf groups reset my-group --to-earliest` |
| `kaf groups delete <group>` | Delete consumer group | `kaf groups delete inactive-group` |

### Message Operations

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf produce <topic>` | Produce messages interactively | `kaf produce orders` |
| `kaf produce <topic> --file <path>` | Produce from file | `kaf produce orders --file data.json` |
| `kaf produce <topic> --count <n>` | Generate test messages | `kaf produce orders --count 100` |
| `kaf produce <topic> --key <key>` | Produce with specific key | `kaf produce orders --key user-123` |
| `kaf consume <topic>` | Consume messages | `kaf consume orders --limit 50` |
| `kaf consume <topic> --from-beginning` | Consume from start | `kaf consume orders --from-beginning` |
| `kaf consume <topic> --from-latest` | Consume from latest messages | `kaf consume orders --from-latest` |
| `kaf consume <topic> --group <group>` | Consume with group | `kaf consume orders --group my-app` |

### Offset Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf offsets show <topic>` | Show partition offsets for topic | `kaf offsets show orders` |
| `kaf offsets reset <topic>` | Reset partition offsets | `kaf offsets reset orders --to-earliest` |

### Broker Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf brokers list` | List all brokers | Show broker IDs and connection info |
| `kaf brokers describe <broker-id>` | Show broker details | `kaf brokers describe 1` |
| `kaf brokers metrics <broker-id>` | Show broker Prometheus metrics | `kaf brokers metrics 1` (requires --broker-metrics-port) |

### Broker Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf brokers configs list` | List configurations for all brokers | Shows 45+ comprehensive settings per broker |
| `kaf brokers configs get <broker-id>` | Show specific broker config | `kaf brokers configs get 1` |
| `kaf brokers configs set <broker-id> <key>=<value>` | Update broker config | `kaf brokers configs set 1 log.retention.hours=72` |

### Health & Monitoring

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf health check` | Run comprehensive health checks | Full cluster diagnostics |
| `kaf health brokers` | Check broker connectivity | `kaf health brokers` |
| `kaf health topics` | Validate topic health | Check topic accessibility |
| `kaf health groups` | Check consumer group health | Monitor group status |

### Utility Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf util random-key` | Generate random message key | For testing message keys |
| `kaf tail <topic>` | Tail messages in real-time | `kaf tail orders` |
| `kaf util version` | Show version information | Display CLI version |
| `kaf completion <shell>` | Generate completion scripts | `kaf completion bash` |

## 🔧 Configuration

The CLI stores configuration in `~/.kaf/config.yml`:

```yaml
current-context: dev

clusters:
  dev:
    bootstrap: localhost:9092
    broker-metrics-port: 9308

  staging:
    bootstrap: kafka-staging:9092
    broker-metrics-port: 9308
    security:
      sasl:
        mechanism: PLAIN
        username: stage-user
        password: stage-pass

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

## 📊 Output Formats

All commands support multiple output formats:

```bash
# Default table format (human-readable)
kaf topics list

# JSON format (for automation)
kaf topics list --output json

# YAML format
kaf topics list --output yaml
```

## 🔒 Security

### SASL Authentication

```bash
# The tool supports various SASL mechanisms
# Configure through the YAML config file or environment variables
kaf config add secure-cluster --bootstrap "secure-kafka:9092"
# Then edit ~/.kaf/config.yml to add security settings
```

### SSL/TLS

Enable SSL in your cluster configuration:

```yaml
clusters:
  secure:
    bootstrap: kafka-ssl:9093
    security:
      ssl: true
      sasl:
        mechanism: SCRAM-SHA-512
        username: username
        password: password
```

## 🎯 Common Use Cases

### Development Workflow

```bash
# Set up development environment
kaf config add local --bootstrap "localhost:9092"
kaf config use local

# Create test topic and generate data
kaf topics create test-events --partitions 1 --replication 1
kaf produce test-events --count 100

# Monitor the data
kaf consume test-events --from-beginning --limit 10

# Or tail real-time messages
kaf tail test-events

# Configure topic settings
kaf topics config set test-events retention.ms=3600000
kaf topics config list test-events
```

### Production Operations

```bash
# Switch to production
kaf config use prod

# Check cluster health
kaf health check

# Monitor consumer lag
kaf groups list
kaf groups lag critical-processor

# Inspect broker configurations
kaf brokers configs list
kaf brokers describe 1

# List topic details and configurations
kaf topics list
kaf topics describe critical-topic
kaf topics config list critical-topic
```

### Automation & Scripting

```bash
#!/bin/bash
# Get all topics as JSON for processing
TOPICS=$(kaf topics list --output json)

# Check if specific topic exists
if kaf topics describe user-events --output json > /dev/null 2>&1; then
    echo "Topic exists"
else
    echo "Creating topic..."
    kaf topics create user-events --partitions 6 --replication 3

    # Configure the topic
    kaf topics config set user-events retention.ms=86400000
    kaf topics config set user-events cleanup.policy=delete
fi

# Monitor consumer group lag
LAG=$(kaf groups lag my-service --output json)
echo "Current lag: $LAG"
```

### Topic Configuration Management

```bash
# List configurations for all topics (comprehensive view)
kaf topics configs list

# Get detailed configs for specific topic
kaf topics configs get orders

# Update topic settings
kaf topics configs set orders retention.ms=604800000  # 7 days
kaf topics configs set orders segment.ms=3600000      # 1 hour

# Remove configuration overrides
kaf topics configs delete orders cleanup.policy
```

### Broker Configuration Management

```bash
# List all broker configurations
kaf brokers configs list

# View specific broker configuration
kaf brokers configs get 1

# Update broker settings
kaf brokers configs set 1 log.retention.hours=72
```

### Broker Metrics Monitoring

```bash
# Configure cluster with metrics port
kaf config add prod --bootstrap kafka-prod:9092 --broker-metrics-port 9308

# View Prometheus metrics for specific broker
kaf brokers metrics 1

# Sample metrics include:
# - Kafka server metrics (requests, throughput)
# - JVM memory usage and garbage collection
# - Network I/O statistics
# - Process CPU and memory utilization
```

## 🆘 Troubleshooting

### Connection Issues

```bash
# Test broker connectivity
kaf health brokers

# Verify current configuration
kaf config current

# Test with specific cluster
kaf config use dev
kaf brokers list
```

### Topic Issues

```bash
# Check if topic exists and view details
kaf topics describe my-topic

# List all topics to verify
kaf topics list

# Check topic configurations
kaf topics configs list      # All topics
kaf topics configs get my-topic  # Specific topic
```

### Consumer Issues

```bash
# Check consumer groups and lag
kaf groups list
kaf groups describe my-group
kaf groups lag my-group

# Reset consumer offsets if needed
kaf groups reset my-group --to-earliest
```

### Configuration Issues

```bash
# Verify cluster configuration
kaf config current

# Export and inspect full config
kaf config export --output yaml

# Test cluster connectivity
kaf health check
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Test thoroughly with tab completion
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Happy Kafka-ing! 🎉**

For more help with any command, use `kaf <command> --help` or enable tab completion for the best experience.