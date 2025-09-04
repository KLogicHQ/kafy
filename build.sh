#!/bin/bash
set -e

VERSION="v0.1.0"
OUTPUT_DIR="release"

mkdir -p $OUTPUT_DIR

echo "Building kaf $VERSION for multiple platforms..."
echo "Note: Cross-compilation requires librdkafka to be available for each target platform"

# Build for current platform with CGO enabled (should always work)
echo "Building for current platform..."
CGO_ENABLED=1 go build -o $OUTPUT_DIR/kaf-$(go env GOOS)-$(go env GOARCH)

# For cross-compilation, we need to build using Docker or use static linking
# This is a simplified approach that may require librdkafka development libraries
echo "Attempting cross-compilation..."

# Linux AMD64
echo "Building for Linux AMD64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-linux-amd64 2>/dev/null || {
    echo "Warning: Linux AMD64 build failed - librdkafka not available for cross-compilation"
    echo "Consider using Docker or building on the target platform"
}

# Linux ARM64
echo "Building for Linux ARM64..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-linux-arm64 2>/dev/null || {
    echo "Warning: Linux ARM64 build failed - librdkafka not available for cross-compilation"
}

# macOS builds
echo "Building for macOS..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-macos-amd64 2>/dev/null || {
    echo "Warning: macOS AMD64 build failed - librdkafka not available for cross-compilation"
}

CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-macos-arm64 2>/dev/null || {
    echo "Warning: macOS ARM64 build failed - librdkafka not available for cross-compilation"
}

# Windows builds (typically more challenging with CGO)
echo "Building for Windows..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-windows-amd64.exe 2>/dev/null || {
    echo "Warning: Windows AMD64 build failed - librdkafka not available for cross-compilation"
}

echo ""
echo "Build completed! Available binaries:"
ls -la $OUTPUT_DIR/kaf-* 2>/dev/null || echo "No binaries were created"
echo ""
echo "Version: $VERSION"
echo ""
echo "For full cross-platform builds, consider:"
echo "1. Using Docker with appropriate cross-compilation toolchains"
echo "2. Building on each target platform individually"
echo "3. Using GitHub Actions with matrix builds (see .github/workflows/build-release.yml)"