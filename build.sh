#!/bin/bash
set -e

VERSION="v1.0.0"
OUTPUT_DIR="release"
DIST_DIR="$OUTPUT_DIR/dist"

# Clean and create directories
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR
mkdir -p $DIST_DIR

echo "Building kaf $VERSION for multiple platforms..."
echo "Note: Cross-compilation requires librdkafka to be available for each target platform"

# Function to create archive packages
create_package() {
    local os=$1
    local arch=$2
    local ext=$3
    local binary_name="kaf"
    
    if [ "$os" = "windows" ]; then
        binary_name="kaf.exe"
    fi
    
    local binary_path="$OUTPUT_DIR/kaf-$os-$arch$ext"
    local package_name="kaf-$VERSION-$os-$arch"
    local package_dir="$OUTPUT_DIR/$package_name"
    
    if [ -f "$binary_path" ]; then
        echo "Creating package for $os-$arch..."
        
        # Create package directory
        mkdir -p "$package_dir"
        
        # Copy binary and rename appropriately
        cp "$binary_path" "$package_dir/$binary_name"
        
        # Copy additional files
        cp README.md "$package_dir/" 2>/dev/null || true
        cp LICENSE "$package_dir/" 2>/dev/null || true
        
        # Create archive based on platform
        cd "$OUTPUT_DIR"
        if [ "$os" = "windows" ]; then
            echo "Creating ZIP archive for Windows..."
            zip -r "$DIST_DIR/$package_name.zip" "$package_name"
        else
            echo "Creating tar.gz archive for $os..."
            tar -czf "$DIST_DIR/$package_name.tar.gz" "$package_name"
        fi
        cd - > /dev/null
        
        # Clean up temporary package directory
        rm -rf "$package_dir"
    fi
}

# Build for current platform with CGO enabled (should always work)
echo "Building for current platform..."
CGO_ENABLED=1 go build -o $OUTPUT_DIR/kaf-$(go env GOOS)-$(go env GOARCH)$([ "$(go env GOOS)" = "windows" ] && echo ".exe" || echo "")

# For cross-compilation, we need to build using Docker or use static linking
echo "Attempting cross-compilation..."

# Linux AMD64
echo "Building for Linux AMD64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-linux-amd64 2>/dev/null || {
    echo "Warning: Linux AMD64 build failed - librdkafka not available for cross-compilation"
}

# Linux ARM64
echo "Building for Linux ARM64..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-linux-arm64 2>/dev/null || {
    echo "Warning: Linux ARM64 build failed - librdkafka not available for cross-compilation"
}

# macOS builds
echo "Building for macOS AMD64..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-darwin-amd64 2>/dev/null || {
    echo "Warning: macOS AMD64 build failed - librdkafka not available for cross-compilation"
}

echo "Building for macOS ARM64..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-darwin-arm64 2>/dev/null || {
    echo "Warning: macOS ARM64 build failed - librdkafka not available for cross-compilation"
}

# Windows builds
echo "Building for Windows AMD64..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-windows-amd64.exe 2>/dev/null || {
    echo "Warning: Windows AMD64 build failed - librdkafka not available for cross-compilation"
}

echo ""
echo "Creating distribution packages..."

# Create packages for each successful build
create_package "linux" "amd64" ""
create_package "linux" "arm64" ""
create_package "darwin" "amd64" ""
create_package "darwin" "arm64" ""
create_package "windows" "amd64" ".exe"

echo ""
echo "Build completed! Available packages:"
ls -la $DIST_DIR/ 2>/dev/null || echo "No packages were created"

echo ""
echo "Raw binaries:"
ls -la $OUTPUT_DIR/kaf-* 2>/dev/null || echo "No binaries were created"

echo ""
echo "Version: $VERSION"
echo "Distribution packages saved to: $DIST_DIR/"