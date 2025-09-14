# 🚀 kafy - A Unified CLI for Kafka

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

**Recommended for most users** - Download pre-built binaries from the [GitHub Releases](https://github.com/KLogicHQ/kafy/releases) page:

1. **Visit the releases page**: Go to [https://github.com/KLogicHQ/kafy/releases](https://github.com/KLogicHQ/kafy/releases)
2. **Download for your platform**:
   - **Linux**: `kafy-v1.0.0-linux-amd64.tar.gz` (x86_64) or `kafy-v1.0.0-linux-arm64.tar.gz` (ARM64)
   - **macOS**: `kafy-v1.0.0-darwin-amd64.tar.gz` (Intel) or `kafy-v1.0.0-darwin-arm64.tar.gz` (Apple Silicon)
   - **Windows**: `kafy-v1.0.0-windows-amd64.zip`
3. **Extract the archive**:
   ```bash
   # Linux/macOS
   tar -xzf kafy-v1.0.0-linux-amd64.tar.gz

   # Windows (PowerShell)
   Expand-Archive kafy-v1.0.0-windows-amd64.zip
   ```
4. **Move to PATH** (optional but recommended):
   ```bash
   # Linux/macOS
   sudo mv kafy /usr/local/bin/

   # Windows: Add the extracted folder to your PATH environment variable
   ```
5. **Verify installation**:
   ```bash
   kafy --help
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
git clone https://github.com/KLogicHQ/kafy.git
cd kafy

# Download Go dependencies
go mod tidy

# Build for current platform
go build -o kafy .

# Optional: Install globally
sudo mv kafy /usr/local/bin/  # Linux/macOS
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

kafy provides intelligent tab completion for all shells. Set it up once and get auto-completion for topics, consumer groups, broker IDs, cluster names, and command flags.

### Bash Completion

```bash
# For current session only
source <(kafy completion bash)

# Install permanently on Linux
kafy completion bash | sudo tee /etc/bash_completion.d/kafy

# Install permanently on macOS (with Homebrew bash-completion)
kafy completion bash > $(brew --prefix)/etc/bash_completion.d/kafy
```

### Zsh Completion

```bash
# For current session only
source <(kafy completion zsh)

# Install permanently
kafy completion zsh > ~/.zsh/completions/_kafy
# Then add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)

# Or for oh-my-zsh users
kafy completion zsh > ~/.oh-my-zsh/completions/_kafy
```

### Fish Completion

```bash
kafy completion fish > ~/.config/fish/completions/kafy.fish
```

### PowerShell Completion

```powershell
kafy completion powershell | Out-String | Invoke-Expression
```

### What Gets Auto-Completed

- **Topics**: `kafy topics describe <TAB>` → Shows available topics
- **Consumer Groups**: `kafy groups describe <TAB>` → Shows active groups
- **Brokers**: `kafy brokers describe <TAB>` → Shows broker IDs
- **Clusters**: `kafy config use <TAB>` → Shows configured clusters
- **Flags**: `--output <TAB>` → Shows table, json, yaml options

## 🚀 Quick Start

### 1. Configure Your First Cluster

```bash
# Add a development cluster with single bootstrap server
kafy config add dev --bootstrap "localhost:9092" --broker-metrics-port 9308

# Add a production cluster with multiple bootstrap servers for high availability
kafy config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308

# List configured clusters
kafy config list

# Switch between clusters
kafy config use prod
kafy config current
```

### 2. Work with Topics

```bash
# List all topics
kafy topics list

# View partition details with insync status
kafy topics partitions              # All topics
kafy topics partitions orders       # Specific topic

# Create a new topic
kafy topics create orders --partitions 3 --replication 2

# Describe a topic
kafy topics describe orders

# Delete a topic (with confirmation)
kafy topics delete test-topic

# Manage topic configurations
kafy topics configs list      # List configs for ALL topics
kafy topics configs get orders # Get configs for specific topic
kafy topics configs set orders retention.ms=86400000
```

### 3. Produce Messages

```bash
# Interactive message production
kafy produce orders

# Produce from a file
echo '{"order_id": 123, "amount": 99.99}' > order.json
kafy produce orders --file order.json --format json

# Generate test messages
kafy produce orders --count 10

# Produce with a specific key
kafy produce orders --key "customer-123"
```

### 4. Consume Messages

```bash
# Consume messages interactively
kafy consume orders

# Consume from multiple topics simultaneously
kafy consume orders users events

# Consume from beginning with limit
kafy consume orders --from-beginning --limit 20

# Consume from latest messages only
kafy consume orders --from-latest

# Consume with specific consumer group
kafy consume orders --group my-service

# Output in JSON format
kafy consume orders --output json --limit 5

# Tail messages in real-time (like tail -f)
kafy tail orders

# Tail multiple topics simultaneously
kafy tail orders users events
```

## 📖 Complete Command Reference

### Configuration Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy config list` | List all configured clusters | Show cluster overview with metrics port and current context |
| `kafy config current` | Show current active cluster | Display active cluster details including metrics port |
| `kafy config use <name>` | Switch to different cluster | `kafy config use prod` |
| `kafy config add <name> --bootstrap <server>` | Add new cluster | `kafy config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092"` |
| `kafy config update <name>` | Update existing cluster configuration | `kafy config update dev --broker-metrics-port 9309 --zookeeper zk:2181` |
| `kafy config delete <name>` | Remove cluster | `kafy config delete old-cluster` |
| `kafy config rename <old> <new>` | Rename cluster | `kafy config rename dev development` |
| `kafy config export` | Export config to YAML/JSON | Backup or share configurations |
| `kafy config import <file>` | Import config from file | Restore from backup |

### Topic Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy topics list` | List all topics | Show topics with partition/replication info |
| `kafy topics describe <topic>` | Show detailed topic information | `kafy topics describe orders` |
| `kafy topics partitions [topic]` | Show partition details with insync status | `kafy topics partitions orders` or `kafy topics partitions` |
| `kafy topics create <topic>` | Create new topic | `kafy topics create events --partitions 6 --replication 3` |
| `kafy topics delete <topic>` | Delete topic | `kafy topics delete test-topic --force` |
| `kafy topics alter <topic>` | Modify topic settings | `kafy topics alter orders --partitions 10` |
| `kafy topics move-partition <topic>` | Move data between partitions | `kafy topics move-partition orders --source-partition 0 --dest-partition 3` |

### Topic Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy topics configs list` | List configurations for all topics | Shows configs for every topic in cluster |
| `kafy topics configs get <topic>` | Show topic-specific configs | `kafy topics configs get orders` |
| `kafy topics configs set <topic> <key>=<value>` | Set topic configuration | `kafy topics configs set orders retention.ms=86400000` |
| `kafy topics configs delete <topic> <key>` | Remove config override | `kafy topics configs delete orders cleanup.policy` |

### Consumer Group Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy groups list` | List all consumer groups | Show groups with states and member counts |
| `kafy groups describe <group>` | Show detailed group information with members | `kafy groups describe my-service` |
| `kafy groups lag <group>` | Show consumer lag metrics | `kafy groups lag payment-processor` |
| `kafy groups reset <group>` | Reset consumer offsets | `kafy groups reset my-group --to-earliest` |
| `kafy groups delete <group>` | Delete consumer group | `kafy groups delete inactive-group` |

### Message Operations

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy produce <topic>` | Produce messages interactively | `kafy produce orders` |
| `kafy produce <topic> --file <path>` | Produce from file | `kafy produce orders --file data.json` |
| `kafy produce <topic> --count <n>` | Generate test messages | `kafy produce orders --count 100` |
| `kafy produce <topic> --key <key>` | Produce with specific key | `kafy produce orders --key user-123` |
| `kafy consume <topic1> [topic2] ...` | Consume from one or more topics | `kafy consume orders users --limit 50` |
| `kafy consume <topic> --from-beginning` | Consume from start | `kafy consume orders --from-beginning` |
| `kafy consume <topic> --from-latest` | Consume from latest messages | `kafy consume orders --from-latest` |
| `kafy consume <topic> --group <group>` | Consume with group | `kafy consume orders --group my-app` |
| `kafy tail <topic1> [topic2] ...` | Tail messages in real-time | `kafy tail orders users events` |
| `kafy cp <source> <dest>` | Copy messages between topics | `kafy cp orders orders-backup --limit 1000` |
| `kafy cp <source> <dest> --begin-offset <n>` | Copy from specific offset | `kafy cp orders backup --begin-offset 100` |
| `kafy cp <source> <dest> --begin-offset <n> --end-offset <n>` | Copy offset range | `kafy cp orders backup --begin-offset 100 --end-offset 500` |

### Offset Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy offsets show <topic>` | Show partition offsets for topic | `kafy offsets show orders` |
| `kafy offsets reset <topic>` | Reset partition offsets | `kafy offsets reset orders --to-earliest` |

### Broker Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy brokers list` | List all brokers | Show broker IDs and connection info |
| `kafy brokers describe <broker-id>` | Show broker details | `kafy brokers describe 1` |
| `kafy brokers metrics <broker-id>` | Show broker Prometheus metrics | `kafy brokers metrics 1` (requires --broker-metrics-port) |
| `kafy brokers metrics <broker-id> --analyze` | AI-powered metrics analysis | `kafy brokers metrics 1 --analyze --provider openai --model gpt-4o` |

### Broker Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy brokers configs list` | List configurations for all brokers | Shows 45+ comprehensive settings per broker |
| `kafy brokers configs get <broker-id>` | Show specific broker config | `kafy brokers configs get 1` |
| `kafy brokers configs set <broker-id> <key>=<value>` | Update broker config | `kafy brokers configs set 1 log.retention.hours=72` |

### Health & Monitoring

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy health check` | Run comprehensive health checks | Full cluster diagnostics |
| `kafy health brokers` | Check broker connectivity | `kafy health brokers` |
| `kafy health topics` | Validate topic health | Check topic accessibility |
| `kafy health groups` | Check consumer group health | Monitor group status |

### Utility Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kafy util random-key` | Generate random message key | For testing message keys |
| `kafy tail <topic>` | Tail messages in real-time | `kafy tail orders` |
| `kafy util version` | Show version information | Display CLI version |
| `kafy completion <shell>` | Generate completion scripts | `kafy completion bash` |

## 🔧 Configuration

The CLI stores configuration in `~/.kafy/config.yml`:

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
kafy topics list

# JSON format (for automation)
kafy topics list --output json

# YAML format
kafy topics list --output yaml
```

## 🔄 Temporary Cluster Switching

The global `-c` or `--cluster` flag allows you to temporarily switch to a different cluster for any command without changing your current context:

```bash
# Run command against specific cluster without switching context
kafy topics list -c prod                    # List topics on prod cluster
kafy consume orders -c staging --limit 10   # Consume from staging cluster
kafy health check -c dev                    # Check dev cluster health

# Multiple commands can use different clusters
kafy topics create test-topic -c dev --partitions 1
kafy cp test-topic backup-topic -c prod --limit 100

# Your current context remains unchanged
kafy config current  # Still shows your original active cluster
```

**Key Benefits:**
- **No context switching required** - Run commands on any configured cluster instantly
- **Safe operations** - Current context is never modified
- **Script-friendly** - Perfect for automation scripts that work across multiple environments
- **Works with all commands** - Available for every kafy command that connects to Kafka

## 🔒 Security

### SASL Authentication

```bash
# The tool supports various SASL mechanisms
# Configure through the YAML config file or environment variables
kafy config add secure-cluster --bootstrap "secure-kafka:9092"
# Then edit ~/.kafy/config.yml to add security settings
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
kafy config add local --bootstrap "localhost:9092" --broker-metrics-port 9308
kafy config use local

# Create test topic and generate data
kafy topics create test-events --partitions 1 --replication 1
kafy produce test-events --count 100

# Monitor partition health and sync status
kafy topics partitions test-events

# Monitor the data
kafy consume test-events --from-beginning --limit 10

# Monitor multiple topics simultaneously
kafy consume orders test-events users --limit 20

# Or tail real-time messages from multiple topics
kafy tail test-events orders users

# Configure topic settings
kafy topics configs set test-events retention.ms=3600000
kafy topics configs get test-events

# Remove topic configuration overrides
kafy topics configs delete test-events cleanup.policy
```

### Production Operations

```bash
# Add production cluster with multiple bootstrap servers for HA
kafy config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308
kafy config use prod

# Check cluster health
kafy health check

# Monitor consumer groups with state and member information
kafy groups list
kafy groups describe critical-processor
kafy groups lag critical-processor

# Check partition health and sync status
kafy topics partitions critical-topic
kafy topics partitions  # All topics

# Inspect broker configurations and metrics
kafy brokers configs list
kafy brokers describe 1
kafy brokers metrics 1  # Requires --broker-metrics-port

# List topic details and configurations
kafy topics list
kafy topics describe critical-topic
kafy topics configs get critical-topic
```

### Automation & Scripting

```bash
#!/bin/bash
# Set up cluster with multiple bootstrap servers
kafy config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092" --broker-metrics-port 9308

# Get all topics as JSON for processing
TOPICS=$(kafy topics list --output json)

# Check if specific topic exists
if kafy topics describe user-events --output json > /dev/null 2>&1; then
    echo "Topic exists"
    # Check partition sync status
    kafy topics partitions user-events --output json
else
    echo "Creating topic..."
    kafy topics create user-events --partitions 6 --replication 3

    # Configure the topic
    kafy topics configs set user-events retention.ms=86400000
    kafy topics configs set user-events cleanup.policy=delete
fi

# Monitor multiple topics simultaneously in automation
kafy consume user-events orders payments --output json --limit 100

# Monitor consumer group with detailed member information
GROUPS=$(kafy groups list --output json)
kafy groups describe my-service --output json
LAG=$(kafy groups lag my-service --output json)
echo "Current lag: $LAG"
```

### Advanced Data Management

```bash
# Partition data movement for rebalancing or migration
kafy topics move-partition orders --source-partition 0 --dest-partition 3

# Monitor partition health before and after migration
kafy topics partitions orders

# Copy messages between topics for backup or testing
kafy cp production-events staging-events --limit 5000

# Multi-topic consumption for aggregated monitoring
kafy consume orders payments notifications --output json --limit 50

# Real-time monitoring across multiple topics
kafy tail critical-events error-logs audit-trail
```

### Configuration Management

```bash
# Add clusters with single or multiple bootstrap servers
kafy config add dev --bootstrap "localhost:9092" --broker-metrics-port 9308
kafy config add staging --bootstrap "kafka-stage-1:9092,kafka-stage-2:9092" --broker-metrics-port 9308
kafy config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308

# Update existing cluster configurations
kafy config update dev --bootstrap "localhost:9092,localhost:9093" --broker-metrics-port 9309
kafy config update prod --zookeeper "zk-prod-1:2181,zk-prod-2:2181,zk-prod-3:2181"

# View complete cluster information
kafy config list     # Shows all clusters with metrics port and zookeeper
kafy config current  # Shows detailed current cluster info

# Topic configuration management
kafy topics configs list                              # All topic configs
kafy topics configs get orders                        # Specific topic
kafy topics configs set orders retention.ms=604800000 # Update setting
kafy topics configs delete orders cleanup.policy      # Remove override
```

### Broker Configuration Management

```bash
# List all broker configurations
kafy brokers configs list

# View specific broker configuration
kafy brokers configs get 1

# Update broker settings
kafy brokers configs set 1 log.retention.hours=72
```

### Monitoring & Health Checks

```bash
# Configure cluster with metrics support
kafy config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092" --broker-metrics-port 9308

# Comprehensive cluster health monitoring
kafy health check           # Full cluster diagnostics
kafy health brokers         # Broker connectivity
kafy health topics          # Topic accessibility
kafy health groups          # Consumer group health

# Partition health and sync monitoring
kafy topics partitions                    # All topics with INSYNC status
kafy topics partitions critical-topic     # Specific topic partitions

# Consumer group monitoring with member details
kafy groups list                          # Groups with state and member count
kafy groups describe payment-service      # Detailed member information
kafy groups lag payment-service           # Partition lag metrics

# Broker metrics monitoring (Prometheus)
kafy brokers metrics 1                    # Kafka server and JVM metrics
kafy brokers metrics 2                    # Network I/O and process stats

# AI-powered metrics analysis (optional)
export OPENAI_API_KEY="your-openai-key"              # Configure API key
kafy brokers metrics 1 --analyze                      # OpenAI analysis with gpt-4o (default)
kafy brokers metrics 1 --analyze --provider claude    # Use Claude with default model
kafy brokers metrics 1 --analyze --provider grok      # Use Grok with default model
kafy brokers metrics 1 --analyze --provider gemini    # Use Gemini with default model

# Custom model examples
kafy brokers metrics 1 --analyze --model gpt-4o-mini        # Use cheaper OpenAI model
kafy brokers metrics 1 --analyze --provider claude --model claude-3-haiku-20240307  # Use faster Claude model
```

### AI-Powered Metrics Analysis

The broker metrics command supports optional AI analysis to provide intelligent recommendations and root cause analysis:

#### Supported AI Providers

| Provider | Environment Variable | Default Model |
|----------|---------------------|---------------|
| **OpenAI** (default) | `OPENAI_API_KEY` | gpt-4o |
| **Claude** | `ANTHROPIC_API_KEY` | claude-3-sonnet-20240229 |
| **Grok** | `XAI_API_KEY` | grok-beta |
| **Gemini** | `GOOGLE_API_KEY` | gemini-pro |

#### Setup Instructions

1. **Configure API Key**: Set the environment variable for your chosen provider:
   ```bash
   # For OpenAI (default)
   export OPENAI_API_KEY="your-openai-api-key"
   
   # For Claude
   export ANTHROPIC_API_KEY="your-anthropic-api-key"
   
   # For Grok
   export XAI_API_KEY="your-xai-api-key"
   
   # For Gemini
   export GOOGLE_API_KEY="your-google-api-key"
   ```

2. **Use AI Analysis**: Add the `--analyze` flag to any metrics command:
   ```bash
   # Basic AI analysis with OpenAI (uses gpt-4o by default)
   kafy brokers metrics 1 --analyze
   
   # Use specific AI provider with default model
   kafy brokers metrics 1 --analyze --provider claude
   
   # Use specific AI provider with custom model
   kafy brokers metrics 1 --analyze --provider openai --model gpt-4o-mini
   kafy brokers metrics 1 --analyze --provider claude --model claude-3-haiku-20240307
   kafy brokers metrics 1 --analyze --provider gemini --model gemini-1.5-pro
   ```

#### Analysis Output

The AI analysis provides structured insights:
- **📊 Summary**: Overall health assessment
- **⚠️ Issues Identified**: Performance problems and bottlenecks  
- **🔍 Root Cause Analysis**: Explanations of what's causing issues
- **💡 Recommendations**: Specific, actionable solutions

Example output:
```
================================================================================
🤖 AI ANALYSIS & RECOMMENDATIONS  
================================================================================
🔄 Analyzing metrics with OPENAI...

📊 SUMMARY:
   Broker shows healthy performance with normal memory usage and stable throughput

⚠️ ISSUES IDENTIFIED:
   1. High GC frequency indicates memory pressure
   2. Consumer lag building up on topic '__consumer_offsets'

💡 RECOMMENDATIONS:
   1. Increase heap size from 4GB to 6GB
   2. Tune G1GC settings for better throughput
   3. Monitor consumer group distribution
```

**Note**: AI analysis is completely optional and requires your own API keys. All metrics are processed securely through the configured AI provider.

## 🆘 Troubleshooting

### Connection Issues

```bash
# Test broker connectivity
kafy health brokers

# Verify current configuration
kafy config current

# Test with specific cluster
kafy config use dev
kafy brokers list
```

### Topic Issues

```bash
# Check if topic exists and view details
kafy topics describe my-topic

# List all topics to verify
kafy topics list

# Check topic configurations
kafy topics configs list      # All topics
kafy topics configs get my-topic  # Specific topic
```

### Consumer Issues

```bash
# Check consumer groups and lag
kafy groups list
kafy groups describe my-group
kafy groups lag my-group

# Reset consumer offsets if needed
kafy groups reset my-group --to-earliest
```

### Configuration Issues

```bash
# Verify cluster configuration
kafy config current

# Export and inspect full config
kafy config export --output yaml

# Test cluster connectivity
kafy health check
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

For more help with any command, use `kafy <command> --help` or enable tab completion for the best experience.