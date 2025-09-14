#!/bin/bash
set -e

VERSION="v1.0.0"
OUTPUT_DIR="release"
DIST_DIR="$OUTPUT_DIR/dist"

# Clean and create directories
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR
mkdir -p $DIST_DIR

echo "Building kafy $VERSION for multiple platforms..."

# Function to create archive packages
create_package() {
    local os=$1
    local arch=$2
    local ext=$3
    local binary_name="kafy"
    
    if [ "$os" = "windows" ]; then
        binary_name="kafy.exe"
    fi
    
    local binary_path="$OUTPUT_DIR/kafy-$os-$arch$ext"
    local package_name="kafy-$VERSION-$os-$arch"
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
        
        echo "✅ Package created: $package_name"
    else
        echo "❌ Binary not found for $os-$arch, skipping package creation"
    fi
}

# Function to build using Docker (for Linux targets)
build_with_docker() {
    local arch=$1
    local platform="linux/$arch"
    
    echo "Building for Linux $arch using Docker..."
    
    if command -v docker >/dev/null 2>&1; then
        if docker buildx build \
            --platform "$platform" \
            --output type=local,dest="$OUTPUT_DIR/docker-out-$arch" \
            --target final \
            . 2>/dev/null; then
            
            # Extract binary from Docker output
            if [ -f "$OUTPUT_DIR/docker-out-$arch/root/kafy" ]; then
                cp "$OUTPUT_DIR/docker-out-$arch/root/kafy" "$OUTPUT_DIR/kafy-linux-$arch"
                rm -rf "$OUTPUT_DIR/docker-out-$arch"
                echo "✅ Linux $arch build successful via Docker"
                return 0
            fi
        fi
    fi
    
    echo "❌ Linux $arch build failed - Docker not available or build failed"
    return 1
}

# Function for native build attempts
try_native_build() {
    local os=$1
    local arch=$2
    local output_name="$OUTPUT_DIR/kafy-$os-$arch"
    
    if [ "$os" = "windows" ]; then
        output_name="$output_name.exe"
    fi
    
    echo "Attempting native build for $os $arch..."
    
    if CGO_ENABLED=1 GOOS="$os" GOARCH="$arch" go build -o "$output_name" . 2>/dev/null; then
        echo "✅ $os $arch build successful (native)"
        return 0
    else
        echo "❌ $os $arch build failed (native)"
        return 1
    fi
}

echo ""
echo "=== Building for current platform ==="
current_os=$(go env GOOS)
current_arch=$(go env GOARCH)
current_ext=""
if [ "$current_os" = "windows" ]; then
    current_ext=".exe"
fi

if CGO_ENABLED=1 go build -o "$OUTPUT_DIR/kafy-$current_os-$current_arch$current_ext" .; then
    echo "✅ Current platform ($current_os-$current_arch) build successful"
else
    echo "❌ Current platform build failed"
    exit 1
fi

echo ""
echo "=== Cross-compilation attempts ==="

# Track build results
declare -A build_results

# Try Linux builds with Docker first, fallback to native
for arch in "amd64" "arm64"; do
    if build_with_docker "$arch"; then
        build_results["linux-$arch"]="success"
    elif try_native_build "linux" "$arch"; then
        build_results["linux-$arch"]="success"
    else
        build_results["linux-$arch"]="failed"
    fi
done

# Try macOS builds (native only)
for arch in "amd64" "arm64"; do
    if try_native_build "darwin" "$arch"; then
        build_results["darwin-$arch"]="success"
    else
        build_results["darwin-$arch"]="failed"
    fi
done

# Try Windows build (native only)
if try_native_build "windows" "amd64"; then
    build_results["windows-amd64"]="success"
else
    build_results["windows-amd64"]="failed"
fi

echo ""
echo "=== Creating distribution packages ==="

# Create packages for successful builds
create_package "linux" "amd64" ""
create_package "linux" "arm64" ""
create_package "darwin" "amd64" ""
create_package "darwin" "arm64" ""
create_package "windows" "amd64" ".exe"

echo ""
echo "=== Build Summary ==="
echo "Version: $VERSION"

success_count=0
total_count=0

echo ""
echo "Build Results:"
for target in "${!build_results[@]}"; do
    result="${build_results[$target]}"
    total_count=$((total_count + 1))
    if [ "$result" = "success" ]; then
        echo "  ✅ $target"
        success_count=$((success_count + 1))
    else
        echo "  ❌ $target"
    fi
done

echo ""
echo "Success rate: $success_count/$total_count targets"

echo ""
echo "Distribution packages:"
if [ -d "$DIST_DIR" ] && [ "$(ls -A $DIST_DIR 2>/dev/null)" ]; then
    ls -la "$DIST_DIR/"
else
    echo "  No packages were created"
fi

echo ""
echo "Raw binaries:"
if ls "$OUTPUT_DIR"/kafy-* 1> /dev/null 2>&1; then
    ls -la "$OUTPUT_DIR"/kafy-*
else
    echo "  No binaries were created"
fi

# Exit with error if no builds succeeded
if [ $success_count -eq 0 ]; then
    echo ""
    echo "❌ All builds failed!"
    exit 1
fi

echo ""
echo "Distribution packages saved to: $DIST_DIR/"