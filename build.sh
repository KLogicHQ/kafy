#!/bin/bash
set -e

VERSION="v0.1.0"
OUTPUT_DIR="release"

mkdir -p $OUTPUT_DIR

# Linux
GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-linux-amd64
GOOS=linux GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-linux-arm64

# macOS
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-macos-amd64
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-macos-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/kaf-windows-amd64.exe
GOOS=windows GOARCH=arm64 go build -o $OUTPUT_DIR/kaf-windows-arm64.exe

echo "Build completed! Binaries created in $OUTPUT_DIR/"
echo "Version: $VERSION"
ls -la $OUTPUT_DIR/