# 🚀 kkl - A Unified CLI for Kafka

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

**Recommended for most users** - Download pre-built binaries from the [GitHub Releases](https://github.com/KLogicHQ/kkl/releases) page:

1. **Visit the releases page**: Go to [https://github.com/KLogicHQ/kkl/releases](https://github.com/KLogicHQ/kkl/releases)
2. **Download for your platform**:
   - **Linux**: `kkl-v1.0.0-linux-amd64.tar.gz` (x86_64) or `kkl-v1.0.0-linux-arm64.tar.gz` (ARM64)
   - **macOS**: `kkl-v1.0.0-darwin-amd64.tar.gz` (Intel) or `kkl-v1.0.0-darwin-arm64.tar.gz` (Apple Silicon)
   - **Windows**: `kkl-v1.0.0-windows-amd64.zip`
3. **Extract the archive**:
   ```bash
   # Linux/macOS
   tar -xzf kkl-v1.0.0-linux-amd64.tar.gz

   # Windows (PowerShell)
   Expand-Archive kkl-v1.0.0-windows-amd64.zip
   ```
4. **Move to PATH** (optional but recommended):
   ```bash
   # Linux/macOS
   sudo mv kkl /usr/local/bin/

   # Windows: Add the extracted folder to your PATH environment variable
   ```
5. **Verify installation**:
   ```bash
   kkl --help
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
git clone https://github.com/KLogicHQ/kkl.git
cd kkl

# Download Go dependencies
go mod tidy

# Build for current platform
go build -o kkl .

# Optional: Install globally
sudo mv kkl /usr/local/bin/  # Linux/macOS
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

kkl provides intelligent tab completion for all shells. Set it up once and get auto-completion for topics, consumer groups, broker IDs, cluster names, and command flags.

### Bash Completion

```bash
# For current session only
source <(kkl completion bash)

# Install permanently on Linux
kkl completion bash | sudo tee /etc/bash_completion.d/kkl

# Install permanently on macOS (with Homebrew bash-completion)
kkl completion bash > $(brew --prefix)/etc/bash_completion.d/kkl
```

### Zsh Completion

```bash
# For current session only
source <(kkl completion zsh)

# Install permanently
kkl completion zsh > ~/.zsh/completions/_kkl
# Then add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)

# Or for oh-my-zsh users
kkl completion zsh > ~/.oh-my-zsh/completions/_kkl
```

### Fish Completion

```bash
kkl completion fish > ~/.config/fish/completions/kkl.fish
```

### PowerShell Completion

```powershell
kkl completion powershell | Out-String | Invoke-Expression
```

### What Gets Auto-Completed

- **Topics**: `kkl topics describe <TAB>` → Shows available topics
- **Consumer Groups**: `kkl groups describe <TAB>` → Shows active groups
- **Brokers**: `kkl brokers describe <TAB>` → Shows broker IDs
- **Clusters**: `kkl config use <TAB>` → Shows configured clusters
- **Flags**: `--output <TAB>` → Shows table, json, yaml options

## 🚀 Quick Start

### 1. Configure Your First Cluster

```bash
# Add a development cluster with single bootstrap server
kkl config add dev --bootstrap "localhost:9092" --broker-metrics-port 9308

# Add a production cluster with multiple bootstrap servers for high availability
kkl config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308

# List configured clusters
kkl config list

# Switch between clusters
kkl config use prod
kkl config current
```

### 2. Work with Topics

```bash
# List all topics
kkl topics list

# View partition details with insync status
kkl topics partitions              # All topics
kkl topics partitions orders       # Specific topic

# Create a new topic
kkl topics create orders --partitions 3 --replication 2

# Describe a topic
kkl topics describe orders

# Delete a topic (with confirmation)
kkl topics delete test-topic

# Manage topic configurations
kkl topics configs list      # List configs for ALL topics
kkl topics configs get orders # Get configs for specific topic
kkl topics configs set orders retention.ms=86400000
```

### 3. Produce Messages

```bash
# Interactive message production
kkl produce orders

# Produce from a file
echo '{"order_id": 123, "amount": 99.99}' > order.json
kkl produce orders --file order.json --format json

# Generate test messages
kkl produce orders --count 10

# Produce with a specific key
kkl produce orders --key "customer-123"
```

### 4. Consume Messages

```bash
# Consume messages interactively
kkl consume orders

# Consume from beginning with limit
kkl consume orders --from-beginning --limit 20

# Consume from latest messages only
kkl consume orders --from-latest

# Consume with specific consumer group
kkl consume orders --group my-service

# Output in JSON format
kkl consume orders --output json --limit 5

# Tail messages in real-time (like tail -f)
kkl tail orders
```

## 📖 Complete Command Reference

### Configuration Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl config list` | List all configured clusters | Show cluster overview with metrics port and current context |
| `kkl config current` | Show current active cluster | Display active cluster details including metrics port |
| `kkl config use <name>` | Switch to different cluster | `kkl config use prod` |
| `kkl config add <name> --bootstrap <server>` | Add new cluster | `kkl config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092"` |
| `kkl config update <name>` | Update existing cluster configuration | `kkl config update dev --broker-metrics-port 9309 --zookeeper zk:2181` |
| `kkl config delete <name>` | Remove cluster | `kkl config delete old-cluster` |
| `kkl config rename <old> <new>` | Rename cluster | `kkl config rename dev development` |
| `kkl config export` | Export config to YAML/JSON | Backup or share configurations |
| `kkl config import <file>` | Import config from file | Restore from backup |

### Topic Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl topics list` | List all topics | Show topics with partition/replication info |
| `kkl topics describe <topic>` | Show detailed topic information | `kkl topics describe orders` |
| `kkl topics partitions [topic]` | Show partition details with insync status | `kkl topics partitions orders` or `kkl topics partitions` |
| `kkl topics create <topic>` | Create new topic | `kkl topics create events --partitions 6 --replication 3` |
| `kkl topics delete <topic>` | Delete topic | `kkl topics delete test-topic --force` |
| `kkl topics alter <topic>` | Modify topic settings | `kkl topics alter orders --partitions 10` |

### Topic Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl topics configs list` | List configurations for all topics | Shows configs for every topic in cluster |
| `kkl topics configs get <topic>` | Show topic-specific configs | `kkl topics configs get orders` |
| `kkl topics configs set <topic> <key>=<value>` | Set topic configuration | `kkl topics configs set orders retention.ms=86400000` |
| `kkl topics configs delete <topic> <key>` | Remove config override | `kkl topics configs delete orders cleanup.policy` |

### Consumer Group Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl groups list` | List all consumer groups | Show groups with states and member counts |
| `kkl groups describe <group>` | Show detailed group information with members | `kkl groups describe my-service` |
| `kkl groups lag <group>` | Show consumer lag metrics | `kkl groups lag payment-processor` |
| `kkl groups reset <group>` | Reset consumer offsets | `kkl groups reset my-group --to-earliest` |
| `kkl groups delete <group>` | Delete consumer group | `kkl groups delete inactive-group` |

### Message Operations

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl produce <topic>` | Produce messages interactively | `kkl produce orders` |
| `kkl produce <topic> --file <path>` | Produce from file | `kkl produce orders --file data.json` |
| `kkl produce <topic> --count <n>` | Generate test messages | `kkl produce orders --count 100` |
| `kkl produce <topic> --key <key>` | Produce with specific key | `kkl produce orders --key user-123` |
| `kkl consume <topic>` | Consume messages | `kkl consume orders --limit 50` |
| `kkl consume <topic> --from-beginning` | Consume from start | `kkl consume orders --from-beginning` |
| `kkl consume <topic> --from-latest` | Consume from latest messages | `kkl consume orders --from-latest` |
| `kkl consume <topic> --group <group>` | Consume with group | `kkl consume orders --group my-app` |

### Offset Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl offsets show <topic>` | Show partition offsets for topic | `kkl offsets show orders` |
| `kkl offsets reset <topic>` | Reset partition offsets | `kkl offsets reset orders --to-earliest` |

### Broker Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl brokers list` | List all brokers | Show broker IDs and connection info |
| `kkl brokers describe <broker-id>` | Show broker details | `kkl brokers describe 1` |
| `kkl brokers metrics <broker-id>` | Show broker Prometheus metrics | `kkl brokers metrics 1` (requires --broker-metrics-port) |
| `kkl brokers metrics <broker-id> --analyze` | AI-powered metrics analysis | `kkl brokers metrics 1 --analyze --provider openai --model gpt-4o` |

### Broker Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl brokers configs list` | List configurations for all brokers | Shows 45+ comprehensive settings per broker |
| `kkl brokers configs get <broker-id>` | Show specific broker config | `kkl brokers configs get 1` |
| `kkl brokers configs set <broker-id> <key>=<value>` | Update broker config | `kkl brokers configs set 1 log.retention.hours=72` |

### Health & Monitoring

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl health check` | Run comprehensive health checks | Full cluster diagnostics |
| `kkl health brokers` | Check broker connectivity | `kkl health brokers` |
| `kkl health topics` | Validate topic health | Check topic accessibility |
| `kkl health groups` | Check consumer group health | Monitor group status |

### Utility Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kkl util random-key` | Generate random message key | For testing message keys |
| `kkl tail <topic>` | Tail messages in real-time | `kkl tail orders` |
| `kkl util version` | Show version information | Display CLI version |
| `kkl completion <shell>` | Generate completion scripts | `kkl completion bash` |

## 🔧 Configuration

The CLI stores configuration in `~/.kkl/config.yml`:

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
kkl topics list

# JSON format (for automation)
kkl topics list --output json

# YAML format
kkl topics list --output yaml
```

## 🔒 Security

### SASL Authentication

```bash
# The tool supports various SASL mechanisms
# Configure through the YAML config file or environment variables
kkl config add secure-cluster --bootstrap "secure-kafka:9092"
# Then edit ~/.kkl/config.yml to add security settings
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
kkl config add local --bootstrap "localhost:9092" --broker-metrics-port 9308
kkl config use local

# Create test topic and generate data
kkl topics create test-events --partitions 1 --replication 1
kkl produce test-events --count 100

# Monitor partition health and sync status
kkl topics partitions test-events

# Monitor the data
kkl consume test-events --from-beginning --limit 10

# Or tail real-time messages
kkl tail test-events

# Configure topic settings
kkl topics configs set test-events retention.ms=3600000
kkl topics configs get test-events
```

### Production Operations

```bash
# Add production cluster with multiple bootstrap servers for HA
kkl config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308
kkl config use prod

# Check cluster health
kkl health check

# Monitor consumer groups with state and member information
kkl groups list
kkl groups describe critical-processor
kkl groups lag critical-processor

# Check partition health and sync status
kkl topics partitions critical-topic
kkl topics partitions  # All topics

# Inspect broker configurations and metrics
kkl brokers configs list
kkl brokers describe 1
kkl brokers metrics 1  # Requires --broker-metrics-port

# List topic details and configurations
kkl topics list
kkl topics describe critical-topic
kkl topics configs get critical-topic
```

### Automation & Scripting

```bash
#!/bin/bash
# Set up cluster with multiple bootstrap servers
kkl config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092" --broker-metrics-port 9308

# Get all topics as JSON for processing
TOPICS=$(kkl topics list --output json)

# Check if specific topic exists
if kkl topics describe user-events --output json > /dev/null 2>&1; then
    echo "Topic exists"
    # Check partition sync status
    kkl topics partitions user-events --output json
else
    echo "Creating topic..."
    kkl topics create user-events --partitions 6 --replication 3

    # Configure the topic
    kkl topics configs set user-events retention.ms=86400000
    kkl topics configs set user-events cleanup.policy=delete
fi

# Monitor consumer group with detailed member information
GROUPS=$(kkl groups list --output json)
kkl groups describe my-service --output json
LAG=$(kkl groups lag my-service --output json)
echo "Current lag: $LAG"
```

### Configuration Management

```bash
# Add clusters with single or multiple bootstrap servers
kkl config add dev --bootstrap "localhost:9092" --broker-metrics-port 9308
kkl config add staging --bootstrap "kafka-stage-1:9092,kafka-stage-2:9092" --broker-metrics-port 9308
kkl config add prod --bootstrap "kafka-prod-1:9092,kafka-prod-2:9092,kafka-prod-3:9092" --broker-metrics-port 9308

# Update existing cluster configurations
kkl config update dev --bootstrap "localhost:9092,localhost:9093" --broker-metrics-port 9309
kkl config update prod --zookeeper "zk-prod-1:2181,zk-prod-2:2181,zk-prod-3:2181"

# View complete cluster information
kkl config list     # Shows all clusters with metrics port and zookeeper
kkl config current  # Shows detailed current cluster info

# Topic configuration management
kkl topics configs list                              # All topic configs
kkl topics configs get orders                        # Specific topic
kkl topics configs set orders retention.ms=604800000 # Update setting
kkl topics configs delete orders cleanup.policy      # Remove override
```

### Broker Configuration Management

```bash
# List all broker configurations
kkl brokers configs list

# View specific broker configuration
kkl brokers configs get 1

# Update broker settings
kkl brokers configs set 1 log.retention.hours=72
```

### Monitoring & Health Checks

```bash
# Configure cluster with metrics support
kkl config add prod --bootstrap "kafka-1:9092,kafka-2:9092,kafka-3:9092" --broker-metrics-port 9308

# Comprehensive cluster health monitoring
kkl health check           # Full cluster diagnostics
kkl health brokers         # Broker connectivity
kkl health topics          # Topic accessibility
kkl health groups          # Consumer group health

# Partition health and sync monitoring
kkl topics partitions                    # All topics with INSYNC status
kkl topics partitions critical-topic     # Specific topic partitions

# Consumer group monitoring with member details
kkl groups list                          # Groups with state and member count
kkl groups describe payment-service      # Detailed member information
kkl groups lag payment-service           # Partition lag metrics

# Broker metrics monitoring (Prometheus)
kkl brokers metrics 1                    # Kafka server and JVM metrics
kkl brokers metrics 2                    # Network I/O and process stats

# AI-powered metrics analysis (optional)
export OPENAI_API_KEY="your-openai-key"              # Configure API key
kkl brokers metrics 1 --analyze                      # OpenAI analysis with gpt-4o (default)
kkl brokers metrics 1 --analyze --provider claude    # Use Claude with default model
kkl brokers metrics 1 --analyze --provider grok      # Use Grok with default model
kkl brokers metrics 1 --analyze --provider gemini    # Use Gemini with default model

# Custom model examples
kkl brokers metrics 1 --analyze --model gpt-4o-mini        # Use cheaper OpenAI model
kkl brokers metrics 1 --analyze --provider claude --model claude-3-haiku-20240307  # Use faster Claude model
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
   kkl brokers metrics 1 --analyze
   
   # Use specific AI provider with default model
   kkl brokers metrics 1 --analyze --provider claude
   
   # Use specific AI provider with custom model
   kkl brokers metrics 1 --analyze --provider openai --model gpt-4o-mini
   kkl brokers metrics 1 --analyze --provider claude --model claude-3-haiku-20240307
   kkl brokers metrics 1 --analyze --provider gemini --model gemini-1.5-pro
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
kkl health brokers

# Verify current configuration
kkl config current

# Test with specific cluster
kkl config use dev
kkl brokers list
```

### Topic Issues

```bash
# Check if topic exists and view details
kkl topics describe my-topic

# List all topics to verify
kkl topics list

# Check topic configurations
kkl topics configs list      # All topics
kkl topics configs get my-topic  # Specific topic
```

### Consumer Issues

```bash
# Check consumer groups and lag
kkl groups list
kkl groups describe my-group
kkl groups lag my-group

# Reset consumer offsets if needed
kkl groups reset my-group --to-earliest
```

### Configuration Issues

```bash
# Verify cluster configuration
kkl config current

# Export and inspect full config
kkl config export --output yaml

# Test cluster connectivity
kkl health check
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

For more help with any command, use `kkl <command> --help` or enable tab completion for the best experience.