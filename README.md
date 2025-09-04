# 🚀 kaf - A Unified CLI for Kafka

A comprehensive Kafka productivity CLI tool that simplifies Kafka operations with a **kubectl-like design philosophy**. Replace complex native `kafka-*` shell scripts with intuitive, short commands.

## 🎯 Features

- **Context-aware operations** - Switch between dev, staging, and prod clusters seamlessly
- **Unified command structure** - kubectl-inspired commands for all Kafka operations
- **Multiple output formats** - Human-readable tables, JSON, and YAML for automation
- **Topic management** - Create, list, describe, and delete topics with ease
- **Message operations** - Produce and consume messages with various formats
- **Consumer group management** - Monitor and manage consumer groups
- **Health monitoring** - Check cluster and broker health
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

- Go 1.19+ (for building)
- Access to Kafka cluster(s)

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

## 📖 Command Reference

### Configuration Commands

| Command | Description |
|---------|-------------|
| `kaf config list` | List all configured clusters |
| `kaf config current` | Show current active cluster |
| `kaf config use <name>` | Switch to different cluster |
| `kaf config add <name> --bootstrap <server>` | Add new cluster |
| `kaf config delete <name>` | Remove cluster |
| `kaf config export` | Export config to YAML/JSON |

### Topic Commands

| Command | Description |
|---------|-------------|
| `kaf topics list` | List all topics |
| `kaf topics describe <topic>` | Show topic details |
| `kaf topics create <topic> --partitions <n> --replication <n>` | Create topic |
| `kaf topics delete <topic>` | Delete topic |

### Consumer Group Commands

| Command | Description |
|---------|-------------|
| `kaf groups list` | List all consumer groups |
| `kaf groups describe <group>` | Show group details |
| `kaf groups lag <group>` | Show lag metrics |

### Message Commands

| Command | Description |
|---------|-------------|
| `kaf produce <topic>` | Produce messages interactively |
| `kaf produce <topic> --file <path>` | Produce from file |
| `kaf produce <topic> --count <n>` | Generate test messages |
| `kaf consume <topic>` | Consume messages |
| `kaf consume <topic> --from-beginning` | Consume from start |
| `kaf consume <topic> --limit <n>` | Limit message count |

### Monitoring Commands

| Command | Description |
|---------|-------------|
| `kaf brokers list` | List all brokers |
| `kaf health check` | Run all health checks |
| `kaf health brokers` | Check broker connectivity |

### Utility Commands

| Command | Description |
|---------|-------------|
| `kaf util random-key` | Generate random key |

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
```

### Production Operations

```bash
# Switch to production
kaf config use prod

# Check cluster health
kaf health check

# Monitor consumer lag
kaf groups list
kaf groups describe my-service

# List topic details
kaf topics list
kaf topics describe critical-topic
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
fi
```

## 🆘 Troubleshooting

### Connection Issues

```bash
# Test broker connectivity
kaf health brokers

# Verify current configuration
kaf config current
```

### Topic Issues

```bash
# Check if topic exists
kaf topics describe my-topic

# List all topics to verify
kaf topics list
```

### Consumer Issues

```bash
# Check consumer groups
kaf groups list

# Monitor specific group
kaf groups describe my-group
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Happy Kafka-ing! 🎉**

For more help with any command, use `kaf <command> --help`