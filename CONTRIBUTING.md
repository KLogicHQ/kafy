# Contributing to kafy

Thank you for your interest in contributing to kafy! This guide will help you get started with development, building, and releasing the project.

## 📋 Prerequisites

- **Go 1.21+** - Required for building the project
- **Git** - For version control
- **Docker** (optional) - For cross-platform builds
- **Kafka cluster** (optional) - For testing against real Kafka

## 🚀 Getting Started

### 1. Clone and Setup

```bash
git clone <repository-url>
cd kafy
go mod download
```

### 2. Running from Source

```bash
# Run directly with go run
go run . --help

# Example commands
go run . config --help
go run . topics list
go run . version
```

### 3. Building the Project

#### Development Build
```bash
# Build for current platform
go build -o kafy .

# Run the built binary
./kafy --help
```

#### Production Builds
```bash
# Use the automated build script for multiple platforms
./build.sh

# This creates:
# - Raw binaries in release/
# - Distribution packages in release/dist/
```

#### Docker Build (Cross-platform)
```bash
# Build for Linux platforms using Docker
docker buildx build --platform linux/amd64 -t kafy:latest .
docker buildx build --platform linux/arm64 -t kafy:latest .
```

## 🧪 Development Workflow

### Project Structure
```
kafy/
├── cmd/                    # CLI commands and subcommands
├── config/                 # Configuration management
├── internal/
│   ├── kafka/             # Kafka client wrapper
│   ├── output/            # Output formatting
│   └── ai/                # AI integration for metrics analysis
├── main.go                # Application entry point
├── build.sh               # Multi-platform build script
└── Dockerfile             # Container build definition
```

### Adding New Commands

1. Create new command files in `cmd/` directory
2. Follow existing patterns in `cmd/topics.go`, `cmd/groups.go`, etc.
3. Add the command to `cmd/root.go` in the `init()` function
4. Implement tab completion in `cmd/completions.go`

### Code Conventions

- Use `kafkaClient` alias for internal Kafka client imports
- Follow Go naming conventions and add proper documentation
- Use the existing formatter pattern for consistent output
- Add appropriate error handling and validation

## 🔄 Testing

### Manual Testing
```bash
# Test basic functionality
go run . config add test --bootstrap localhost:9092
go run . config current
go run . topics list

# Test with different output formats
go run . topics list --output json
go run . brokers list --output yaml
```

### Integration Testing
```bash
# Test against real Kafka cluster (requires running Kafka)
go run . config add local --bootstrap localhost:9092
go run . config use local
go run . health check
```

## 📦 Release Process

### 1. Version Update

Update the version constant in `cmd/root.go`:

```go
const version = "x.y.z"  // Update this line
```

### 2. Create Release Build

```bash
# Build all platform binaries
./build.sh

# Verify build success
ls -la release/dist/
```

### 3. Git Tag and Release

```bash
# Commit version change
git add cmd/root.go
git commit -m "Bump version to x.y.z"

# Create and push tag
git tag -a vx.y.z -m "Release version x.y.z"
git push origin main
git push origin vx.y.z
```

### 4. Distribution

The `build.sh` script creates distribution packages in `release/dist/`:
- `kafy-vx.y.z-linux-amd64.tar.gz`
- `kafy-vx.y.z-linux-arm64.tar.gz`
- `kafy-vx.y.z-darwin-amd64.tar.gz`
- `kafy-vx.y.z-darwin-arm64.tar.gz`
- `kafy-vx.y.z-windows-amd64.zip`

These can be uploaded to GitHub releases or distributed through package managers.

## 🐛 Debugging

### Common Issues

1. **Build failures**: Check Go version and dependencies
   ```bash
   go version  # Should be 1.21+
   go mod tidy
   ```

2. **Kafka connection issues**: Verify cluster configuration
   ```bash
   go run . config current
   go run . health brokers
   ```

3. **Import path issues**: Ensure all imports use `kafy/` prefix

### LSP/IDE Setup

The project uses standard Go modules. Most Go-compatible IDEs will work out of the box:

```bash
# For VS Code with Go extension
go mod tidy
```

## 🤝 Contribution Guidelines

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/my-feature`
3. **Make** your changes following the code conventions
4. **Test** your changes thoroughly
5. **Commit** with clear messages: `git commit -m "Add feature: description"`
6. **Push** to your fork: `git push origin feature/my-feature`
7. **Submit** a pull request with a clear description

### Commit Message Format

```
type: brief description

Longer description explaining the change and why it was needed.

- Bullet points for key changes
- Reference issues: Closes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## 📄 License

This project is licensed under the MIT License. By contributing, you agree that your contributions will be licensed under the same license.

## ❓ Getting Help

- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
- **Documentation**: Check README.md for usage examples

Happy contributing! 🎉