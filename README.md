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

### Build from Source

```bash
# Clone and build
git clone git@github.com:KLogicHQ/kaf.git
cd kaf
go mod tidy
go build -o kaf .

# Make it globally available (optional)
sudo mv kaf /usr/local/bin/
```

### Requirements

- Go 1.21+ (for building)
- Access to Kafka cluster(s)

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
# Add a development cluster
kaf config add dev --bootstrap "localhost:9092"

# Add a production cluster with authentication
kaf config add prod --bootstrap "kafka-prod:9092"

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

# Consume with specific consumer group
kaf consume orders --group my-service

# Output in JSON format
kaf consume orders --output json --limit 5
```

## 📖 Complete Command Reference

### Configuration Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf config list` | List all configured clusters | Show cluster overview with current context |
| `kaf config current` | Show current active cluster | Display active cluster details |
| `kaf config use <name>` | Switch to different cluster | `kaf config use prod` |
| `kaf config add <name> --bootstrap <server>` | Add new cluster | `kaf config add dev --bootstrap localhost:9092` |
| `kaf config delete <name>` | Remove cluster | `kaf config delete old-cluster` |
| `kaf config rename <old> <new>` | Rename cluster | `kaf config rename dev development` |
| `kaf config export` | Export config to YAML/JSON | Backup or share configurations |
| `kaf config import <file>` | Import config from file | Restore from backup |

### Topic Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf topics list` | List all topics | Show topics with partition/replication info |
| `kaf topics describe <topic>` | Show detailed topic information | `kaf topics describe orders` |
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
| `kaf groups list` | List all consumer groups | Show groups with member counts |
| `kaf groups describe <group>` | Show detailed group information | `kaf groups describe my-service` |
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
| `kaf consume <topic> --group <group>` | Consume with group | `kaf consume orders --group my-app` |

### Offset Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf offsets show <topic>` | Show partition offsets for topic | `kaf offsets show orders` (fixed group.id error) |
| `kaf offsets reset <topic>` | Reset partition offsets | `kaf offsets reset orders --to-earliest` |

### Broker Management

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf brokers list` | List all brokers | Show broker IDs and connection info |
| `kaf brokers describe <broker-id>` | Show broker details | `kaf brokers describe 1` |
| `kaf brokers metrics` | Show broker metrics | Display broker performance data |

### Broker Configuration Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `kaf brokers configs list` | List configurations for all brokers | Shows 45+ comprehensive settings per broker |
| `kaf brokers configs get <broker-id>` | Show specific broker config | `kaf brokers configs get 1` (returns same 45+ configs) |
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
| `kaf util version` | Show version information | Display CLI version |
| `kaf completion <shell>` | Generate completion scripts | `kaf completion bash` |

## 🔧 Configuration

The CLI stores configuration in `~/.kaf/config.yml`:

```yaml
current-context: dev

clusters:
  dev:
    bootstrap: localhost:9092

  staging:
    bootstrap: kafka-staging:9092
    security:
      sasl:
        mechanism: PLAIN
        username: stage-user
        password: stage-pass

  prod:
    bootstrap: kafka-prod:9092
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
# List all broker configurations (shows 45+ comprehensive settings)
kaf brokers configs list

# View specific broker configuration (consistent with list - shows same 45+ configs)
kaf brokers configs get 1

# Update broker settings
kaf brokers configs set 1 log.retention.hours=72
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